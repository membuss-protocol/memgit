package swarm

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/membuss-protocol/memgit/pkg/identity"
	"github.com/membuss-protocol/memgit/pkg/models"
	"github.com/membuss-protocol/memgit/pkg/services"
)

// Catalog manages global swarm repository discovery, multi-user star counters, and the decentralized activity stream.
type Catalog struct {
	mu          sync.RWMutex
	repoService *services.RepoService
	idManager   *identity.Manager
	remoteRepos map[string]*models.Repository
	activities  []models.ActivityEvent
}

// NewCatalog initializes the Swarm Catalog.
func NewCatalog(repoService *services.RepoService, idManager *identity.Manager) *Catalog {
	return &Catalog{
		repoService: repoService,
		idManager:   idManager,
		remoteRepos: make(map[string]*models.Repository),
		activities:  make([]models.ActivityEvent, 0),
	}
}

// AllRepositories returns both local and discovered remote swarm repositories.
func (c *Catalog) AllRepositories() []*models.Repository {
	c.mu.RLock()
	defer c.mu.RUnlock()

	local := c.repoService.ListRepos()
	allMap := make(map[string]*models.Repository)

	// Add local repos first
	currentUser := c.idManager.CurrentUser()
	for _, r := range local {
		copy := *r
		copy.IsLocal = true
		if copy.Owner == "" {
			copy.Owner = currentUser.Username
		}
		if copy.FullName == "" {
			copy.FullName = copy.Owner + "/" + copy.Name
		}
		allMap[copy.FullName] = &copy
	}

	// Add remote repos
	for name, r := range c.remoteRepos {
		if _, exists := allMap[name]; !exists {
			copy := *r
			allMap[name] = &copy
		}
	}

	res := make([]*models.Repository, 0, len(allMap))
	for _, r := range allMap {
		res = append(res, r)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].UpdatedAt.After(res[j].UpdatedAt)
	})

	return res
}

// ToggleStar toggles the star status of a repository for the current user.
func (c *Catalog) ToggleStar(repoKey string) (*models.Repository, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	user := c.idManager.CurrentUser()

	// Check if repo is local
	shortName := repoKey
	if parts := strings.Split(repoKey, "/"); len(parts) == 2 {
		shortName = parts[1]
	}

	isStarred := false
	var target *models.Repository

	if localRepo, err := c.repoService.GetRepo(shortName); err == nil {
		target = localRepo
	} else if remote, exists := c.remoteRepos[repoKey]; exists {
		target = remote
	} else {
		for _, r := range c.remoteRepos {
			if r.Name == shortName || r.FullName == repoKey || r.Name == repoKey {
				target = r
				break
			}
		}
	}

	if target == nil {
		return nil, false, fmt.Errorf("repository %q not found", repoKey)
	}

	// Toggle user public key in StarredBy
	foundIdx := -1
	for i, pk := range target.StarredBy {
		if pk == user.PublicKey {
			foundIdx = i
			break
		}
	}

	if foundIdx >= 0 {
		// Unstar
		target.StarredBy = append(target.StarredBy[:foundIdx], target.StarredBy[foundIdx+1:]...)
		if target.StarCount > 0 {
			target.StarCount--
		}
		isStarred = false
	} else {
		// Star
		target.StarredBy = append(target.StarredBy, user.PublicKey)
		target.StarCount++
		isStarred = true

		c.activities = append([]models.ActivityEvent{{
			ID:        fmt.Sprintf("star-%s-%d", target.Name, time.Now().Unix()),
			Type:      "star",
			Actor:     user.Username,
			RepoName:  target.FullName,
			Summary:   fmt.Sprintf("Starred %s", target.FullName),
			Timestamp: time.Now(),
		}}, c.activities...)
	}

	copy := *target
	return &copy, isStarred, nil
}

// ForkRepo creates a local fork of a target repository for the current user.
func (c *Catalog) ForkRepo(ctx context.Context, sourceFullName, newName, baseURL string) (*models.Repository, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	user := c.idManager.CurrentUser()
	if newName == "" {
		parts := strings.Split(sourceFullName, "/")
		if len(parts) == 2 {
			newName = parts[1]
		} else {
			newName = sourceFullName
		}
	}

	// Create repository locally
	desc := fmt.Sprintf("Decentralized fork of %s", sourceFullName)
	forked, err := c.repoService.CreateRepo(ctx, newName, desc, "main", false, false, baseURL)
	if err != nil {
		return nil, err
	}

	forked.Owner = user.Username
	forked.FullName = user.Username + "/" + newName
	forked.ForkedFrom = sourceFullName
	_ = c.repoService.Save(forked)

	c.activities = append([]models.ActivityEvent{{
		ID:        fmt.Sprintf("fork-%s-%d", forked.Name, time.Now().Unix()),
		Type:      "fork",
		Actor:     user.Username,
		RepoName:  forked.FullName,
		Summary:   fmt.Sprintf("Forked %s to %s", sourceFullName, forked.FullName),
		Timestamp: time.Now(),
	}}, c.activities...)

	return forked, nil
}

// ActivityFeed returns the latest global activity feed across the swarm.
func (c *Catalog) ActivityFeed(limit int) []models.ActivityEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 || limit > len(c.activities) {
		limit = len(c.activities)
	}

	res := make([]models.ActivityEvent, limit)
	copy(res, c.activities[:limit])
	return res
}

// RecordActivity records a new activity event.
func (c *Catalog) RecordActivity(evtType, actor, repoName, summary string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.activities = append([]models.ActivityEvent{{
		ID:        fmt.Sprintf("act-%d", time.Now().UnixNano()),
		Type:      evtType,
		Actor:     actor,
		RepoName:  repoName,
		Summary:   summary,
		Timestamp: time.Now(),
	}}, c.activities...)

	if len(c.activities) > 200 {
		c.activities = c.activities[:200]
	}
}
