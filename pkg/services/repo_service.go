package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/membuss-protocol/memgit/pkg/gitengine"
	"github.com/membuss-protocol/memgit/pkg/membuss"
	"github.com/membuss-protocol/memgit/pkg/models"
)

// RepoService handles repository lifecycle, metadata persistence, and Membuss synchronization.
type RepoService struct {
	metaDir string
	engine  *gitengine.Engine
	client  *membuss.Client
	mu      sync.RWMutex
	repos   map[string]*models.Repository
}

// NewRepoService creates a new repository service.
func NewRepoService(metaDir string, engine *gitengine.Engine, client *membuss.Client) (*RepoService, error) {
	if metaDir == "" {
		metaDir = "data/meta"
	}
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return nil, err
	}

	s := &RepoService{
		metaDir: metaDir,
		engine:  engine,
		client:  client,
		repos:   make(map[string]*models.Repository),
	}
	_ = s.loadAll()
	return s, nil
}

func (s *RepoService) metaPath(name string) string {
	return filepath.Join(s.metaDir, name+".json")
}

func (s *RepoService) loadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.metaDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(s.metaDir, entry.Name()))
			if err == nil {
				var repo models.Repository
				if err := json.Unmarshal(data, &repo); err == nil {
					s.repos[repo.Name] = &repo
				}
			}
		}
	}
	return nil
}

// Save writes repository metadata to disk.
func (s *RepoService) Save(repo *models.Repository) error {
	data, err := json.MarshalIndent(repo, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(repo.Name), data, 0o644)
}

// CreateRepo creates a new decentralized repository.
func (s *RepoService) CreateRepo(ctx context.Context, name, description, defaultBranch string, isPrivate, initReadme bool, baseURL string) (*models.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.repos[name]; exists {
		return nil, fmt.Errorf("repository %q already exists", name)
	}

	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// 1. Initialize bare git repository
	_, err := s.engine.InitRepo(name, defaultBranch, description, initReadme)
	if err != nil {
		return nil, err
	}

	// 2. Generate a dedicated Ed25519 MemNS key for this repository
	memnsKeyName := fmt.Sprintf("memgit-%s-%d", name, time.Now().Unix())
	keyCtx, keyCancel := context.WithTimeout(ctx, 500*time.Millisecond)
	keyInfo, err := s.client.GenerateMemNSKey(keyCtx, memnsKeyName)
	keyCancel()

	memnsName := ""
	if err == nil && keyInfo != nil {
		memnsName = keyInfo.MemNSName
	} else {
		// Fallback mock identity if daemon is offline
		memnsName = "memns1z" + name
	}

	// 3. Take initial snapshot to Membuss if initialized
	latestMID := ""
	if initReadme {
		snapCtx, snapCancel := context.WithTimeout(ctx, 1*time.Second)
		mid, err := s.engine.SnapshotToMembuss(snapCtx, name, memnsKeyName)
		snapCancel()
		if err == nil {
			latestMID = mid
		}
	}

	repo := &models.Repository{
		Name:          name,
		Description:   description,
		DefaultBranch: defaultBranch,
		MemNSKey:      memnsKeyName,
		MemNSName:     memnsName,
		LatestMID:     latestMID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		StarCount:     0,
		ForkCount:     0,
		IsPrivate:     isPrivate,
		Topics:        []string{"membuss", "p2p", "decentralized-git"},
		CloneHTTPS:    fmt.Sprintf("%s/git/%s.git", baseURL, name),
		CloneGateway:  fmt.Sprintf("%s/memns/%s", s.client.GatewayBase(), memnsName),
		CloneMembuss:  fmt.Sprintf("membuss://%s", memnsName),
	}

	s.repos[name] = repo
	_ = s.Save(repo)
	return repo, nil
}

// GetRepo returns a repository by name.
func (s *RepoService) GetRepo(name string) (*models.Repository, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repo, exists := s.repos[name]
	if !exists {
		return nil, fmt.Errorf("repository %q not found", name)
	}
	if repo.CloneGateway == "" && repo.MemNSName != "" {
		repo.CloneGateway = fmt.Sprintf("%s/memns/%s", s.client.GatewayBase(), repo.MemNSName)
	}
	return repo, nil
}

// ListRepos returns all repositories.
func (s *RepoService) ListRepos() []*models.Repository {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*models.Repository, 0, len(s.repos))
	for _, r := range s.repos {
		if r.CloneGateway == "" && r.MemNSName != "" {
			r.CloneGateway = fmt.Sprintf("%s/memns/%s", s.client.GatewayBase(), r.MemNSName)
		}
		list = append(list, r)
	}
	return list
}

// SyncToMembuss triggers a fresh snapshot and publishes the update to MemNS.
func (s *RepoService) SyncToMembuss(ctx context.Context, name string) (string, error) {
	repo, err := s.GetRepo(name)
	if err != nil {
		return "", err
	}

	mid, err := s.engine.SnapshotToMembuss(ctx, name, repo.MemNSKey)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	repo.LatestMID = mid
	repo.UpdatedAt = time.Now()
	_ = s.Save(repo)
	s.mu.Unlock()

	return mid, nil
}

// StarRepo toggles a star on a repository.
func (s *RepoService) StarRepo(name string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, exists := s.repos[name]
	if !exists {
		return 0, fmt.Errorf("repository %q not found", name)
	}
	repo.StarCount++
	_ = s.Save(repo)
	return repo.StarCount, nil
}

// ForkRepo creates a fork of an existing repository.
func (s *RepoService) ForkRepo(ctx context.Context, sourceName, newName, baseURL string) (*models.Repository, error) {
	sourceRepo, err := s.GetRepo(sourceName)
	if err != nil {
		return nil, err
	}

	forked, err := s.CreateRepo(ctx, newName, fmt.Sprintf("Forked from %s", sourceName), sourceRepo.DefaultBranch, false, false, baseURL)
	if err != nil {
		return nil, err
	}

	// Copy bare repo objects from source to forked
	srcPath := s.engine.RepoPath(sourceName)
	dstPath := s.engine.RepoPath(newName)
	_ = filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(srcPath, path)
		target := filepath.Join(dstPath, rel)
		_ = os.MkdirAll(filepath.Dir(target), 0o755)
		data, rerr := os.ReadFile(path)
		if rerr == nil {
			_ = os.WriteFile(target, data, info.Mode())
		}
		return nil
	})

	s.mu.Lock()
	sourceRepo.ForkCount++
	_ = s.Save(sourceRepo)

	forked.ForkedFrom = sourceName
	_ = s.Save(forked)
	s.mu.Unlock()

	// Snapshot forked repo
	_, _ = s.SyncToMembuss(ctx, newName)

	return forked, nil
}
