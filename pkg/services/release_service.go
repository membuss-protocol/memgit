package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/membuss-protocol/memgit/pkg/models"
)

// ReleaseService manages releases and tagged snapshots for repositories.
type ReleaseService struct {
	baseDir  string
	mu       sync.RWMutex
	releases map[string]map[int]*models.Release
}

// NewReleaseService creates a new release service instance.
func NewReleaseService(baseDir string) (*ReleaseService, error) {
	if baseDir == "" {
		baseDir = "data/releases"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	s := &ReleaseService{
		baseDir:  baseDir,
		releases: make(map[string]map[int]*models.Release),
	}
	_ = s.loadAll()
	return s, nil
}

func (s *ReleaseService) repoPath(repoName string) string {
	return filepath.Join(s.baseDir, repoName)
}

func (s *ReleaseService) relFile(repoName string, id int) string {
	return filepath.Join(s.repoPath(repoName), fmt.Sprintf("%d.json", id))
}

func (s *ReleaseService) loadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return err
	}

	for _, d := range entries {
		if d.IsDir() {
			repoName := d.Name()
			s.releases[repoName] = make(map[int]*models.Release)

			files, _ := os.ReadDir(s.repoPath(repoName))
			for _, f := range files {
				if filepath.Ext(f.Name()) == ".json" {
					data, err := os.ReadFile(filepath.Join(s.repoPath(repoName), f.Name()))
					if err == nil {
						var rel models.Release
						if err := json.Unmarshal(data, &rel); err == nil {
							s.releases[repoName][rel.ID] = &rel
						}
					}
				}
			}
		}
	}
	return nil
}

func (s *ReleaseService) save(rel *models.Release) error {
	dir := s.repoPath(rel.RepoName)
	_ = os.MkdirAll(dir, 0o755)
	data, err := json.MarshalIndent(rel, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.relFile(rel.RepoName, rel.ID), data, 0o644)
}

// CreateRelease creates a new Release.
func (s *ReleaseService) CreateRelease(repoName, tagName, title, desc, commitSHA, mid string, assets []models.Asset) (*models.Release, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.releases[repoName]; !exists {
		s.releases[repoName] = make(map[int]*models.Release)
	}

	nextID := len(s.releases[repoName]) + 1
	rel := &models.Release{
		ID:          nextID,
		RepoName:    repoName,
		TagName:     tagName,
		Title:       title,
		Description: desc,
		CommitSHA:   commitSHA,
		MID:         mid,
		PublishedAt: time.Now(),
		Assets:      assets,
	}

	s.releases[repoName][nextID] = rel
	_ = s.save(rel)
	return rel, nil
}

// ListReleases lists all releases for a repository.
func (s *ReleaseService) ListReleases(repoName string) []*models.Release {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*models.Release
	repoRels, exists := s.releases[repoName]
	if !exists {
		return list
	}

	for _, rel := range repoRels {
		list = append(list, rel)
	}
	return list
}
