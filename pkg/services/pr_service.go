package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/membuss-protocol/memgit/pkg/gitengine"
	"github.com/membuss-protocol/memgit/pkg/models"
)

// PRService manages pull requests, branch comparisons, and merges.
type PRService struct {
	baseDir string
	engine  *gitengine.Engine
	mu      sync.RWMutex
	pulls   map[string]map[int]*models.PullRequest
}

// NewPRService creates a new Pull Request service.
func NewPRService(baseDir string, engine *gitengine.Engine) (*PRService, error) {
	if baseDir == "" {
		baseDir = "data/pulls"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	s := &PRService{
		baseDir: baseDir,
		engine:  engine,
		pulls:   make(map[string]map[int]*models.PullRequest),
	}
	_ = s.loadAll()
	return s, nil
}

func (s *PRService) repoPath(repoName string) string {
	return filepath.Join(s.baseDir, repoName)
}

func (s *PRService) prFile(repoName string, id int) string {
	return filepath.Join(s.repoPath(repoName), fmt.Sprintf("%d.json", id))
}

func (s *PRService) loadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return err
	}

	for _, d := range entries {
		if d.IsDir() {
			repoName := d.Name()
			s.pulls[repoName] = make(map[int]*models.PullRequest)

			files, _ := os.ReadDir(s.repoPath(repoName))
			for _, f := range files {
				if filepath.Ext(f.Name()) == ".json" {
					data, err := os.ReadFile(filepath.Join(s.repoPath(repoName), f.Name()))
					if err == nil {
						var pr models.PullRequest
						if err := json.Unmarshal(data, &pr); err == nil {
							s.pulls[repoName][pr.ID] = &pr
						}
					}
				}
			}
		}
	}
	return nil
}

func (s *PRService) save(pr *models.PullRequest) error {
	dir := s.repoPath(pr.RepoName)
	_ = os.MkdirAll(dir, 0o755)
	data, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.prFile(pr.RepoName, pr.ID), data, 0o644)
}

// CreatePR creates a new Pull Request.
func (s *PRService) CreatePR(repoName, title, body, author, sourceBranch, targetBranch, sourceRepo string) (*models.PullRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pulls[repoName]; !exists {
		s.pulls[repoName] = make(map[int]*models.PullRequest)
	}

	if sourceRepo == "" {
		sourceRepo = repoName
	}

	nextID := len(s.pulls[repoName]) + 1
	pr := &models.PullRequest{
		ID:           nextID,
		RepoName:     repoName,
		Title:        title,
		Body:         body,
		Author:       author,
		State:        "open",
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
		SourceRepo:   sourceRepo,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Comments:     []models.IssueComment{},
	}

	s.pulls[repoName][nextID] = pr
	_ = s.save(pr)
	return pr, nil
}

// GetPR returns a Pull Request by ID.
func (s *PRService) GetPR(repoName string, id int) (*models.PullRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repoPRs, exists := s.pulls[repoName]
	if !exists {
		return nil, fmt.Errorf("no pull requests found for %q", repoName)
	}
	pr, exists := repoPRs[id]
	if !exists {
		return nil, fmt.Errorf("pull request #%d not found", id)
	}
	return pr, nil
}

// ListPRs lists all pull requests for a repository.
func (s *PRService) ListPRs(repoName, state string) []*models.PullRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*models.PullRequest
	repoPRs, exists := s.pulls[repoName]
	if !exists {
		return list
	}

	for _, pr := range repoPRs {
		if state == "" || pr.State == state {
			list = append(list, pr)
		}
	}
	return list
}

// MergePR marks a pull request as merged and fast-forwards the target branch.
func (s *PRService) MergePR(repoName string, id int) (*models.PullRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repoPRs, exists := s.pulls[repoName]
	if !exists {
		return nil, fmt.Errorf("repository %q not found", repoName)
	}
	pr, exists := repoPRs[id]
	if !exists {
		return nil, fmt.Errorf("pull request #%d not found", id)
	}
	if pr.State == "merged" {
		return nil, fmt.Errorf("pull request #%d is already merged", id)
	}

	// Update git ref
	repo, err := s.engine.OpenRepo(repoName)
	if err == nil {
		sourceCommit, serr := s.engine.ResolveRef(repo, pr.SourceBranch)
		if serr == nil {
			_ = s.engine.CreateBranch(repoName, pr.TargetBranch, sourceCommit.Hash.String())
		}
	}

	now := time.Now()
	pr.State = "merged"
	pr.MergedAt = &now
	pr.UpdatedAt = now
	_ = s.save(pr)
	return pr, nil
}

// AddComment adds a comment to a PR.
func (s *PRService) AddComment(repoName string, id int, author, body string) (*models.IssueComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repoPRs, exists := s.pulls[repoName]
	if !exists {
		return nil, fmt.Errorf("repository %q not found", repoName)
	}
	pr, exists := repoPRs[id]
	if !exists {
		return nil, fmt.Errorf("pull request #%d not found", id)
	}

	comment := models.IssueComment{
		ID:        len(pr.Comments) + 1,
		IssueID:   id,
		Author:    author,
		Body:      body,
		CreatedAt: time.Now(),
	}

	pr.Comments = append(pr.Comments, comment)
	pr.UpdatedAt = time.Now()
	_ = s.save(pr)
	return &comment, nil
}
