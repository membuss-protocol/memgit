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

// IssueService manages decentralized issues and comments for repositories.
type IssueService struct {
	baseDir string
	mu      sync.RWMutex
	issues  map[string]map[int]*models.Issue
}

// NewIssueService creates a new issue service instance.
func NewIssueService(baseDir string) (*IssueService, error) {
	if baseDir == "" {
		baseDir = "data/issues"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	s := &IssueService{
		baseDir: baseDir,
		issues:  make(map[string]map[int]*models.Issue),
	}
	_ = s.loadAll()
	return s, nil
}

func (s *IssueService) repoPath(repoName string) string {
	return filepath.Join(s.baseDir, repoName)
}

func (s *IssueService) issueFile(repoName string, id int) string {
	return filepath.Join(s.repoPath(repoName), fmt.Sprintf("%d.json", id))
}

func (s *IssueService) loadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return err
	}

	for _, d := range entries {
		if d.IsDir() {
			repoName := d.Name()
			s.issues[repoName] = make(map[int]*models.Issue)

			files, _ := os.ReadDir(s.repoPath(repoName))
			for _, f := range files {
				if filepath.Ext(f.Name()) == ".json" {
					data, err := os.ReadFile(filepath.Join(s.repoPath(repoName), f.Name()))
					if err == nil {
						var issue models.Issue
						if err := json.Unmarshal(data, &issue); err == nil {
							s.issues[repoName][issue.ID] = &issue
						}
					}
				}
			}
		}
	}
	return nil
}

func (s *IssueService) save(issue *models.Issue) error {
	dir := s.repoPath(issue.RepoName)
	_ = os.MkdirAll(dir, 0o755)
	data, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.issueFile(issue.RepoName, issue.ID), data, 0o644)
}

// CreateIssue creates a new issue in the repository.
func (s *IssueService) CreateIssue(repoName, title, body, author string, labels []string) (*models.Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.issues[repoName]; !exists {
		s.issues[repoName] = make(map[int]*models.Issue)
	}

	nextID := len(s.issues[repoName]) + 1
	issue := &models.Issue{
		ID:        nextID,
		RepoName:  repoName,
		Title:     title,
		Body:      body,
		Author:    author,
		State:     "open",
		Labels:    labels,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Comments:  []models.IssueComment{},
	}

	s.issues[repoName][nextID] = issue
	_ = s.save(issue)
	return issue, nil
}

// GetIssue returns a specific issue by ID.
func (s *IssueService) GetIssue(repoName string, id int) (*models.Issue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	repoIssues, exists := s.issues[repoName]
	if !exists {
		return nil, fmt.Errorf("no issues found for repository %q", repoName)
	}
	issue, exists := repoIssues[id]
	if !exists {
		return nil, fmt.Errorf("issue #%d not found", id)
	}
	return issue, nil
}

// ListIssues lists all issues for a repository, optionally filtered by state.
func (s *IssueService) ListIssues(repoName, state string) []*models.Issue {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*models.Issue
	repoIssues, exists := s.issues[repoName]
	if !exists {
		return list
	}

	for _, issue := range repoIssues {
		if state == "" || issue.State == state {
			list = append(list, issue)
		}
	}
	return list
}

// UpdateIssueState toggles open / closed state.
func (s *IssueService) UpdateIssueState(repoName string, id int, state string) (*models.Issue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repoIssues, exists := s.issues[repoName]
	if !exists {
		return nil, fmt.Errorf("repository %q not found", repoName)
	}
	issue, exists := repoIssues[id]
	if !exists {
		return nil, fmt.Errorf("issue #%d not found", id)
	}

	issue.State = state
	issue.UpdatedAt = time.Now()
	_ = s.save(issue)
	return issue, nil
}

// AddComment appends a comment to an issue.
func (s *IssueService) AddComment(repoName string, id int, author, body string) (*models.IssueComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repoIssues, exists := s.issues[repoName]
	if !exists {
		return nil, fmt.Errorf("repository %q not found", repoName)
	}
	issue, exists := repoIssues[id]
	if !exists {
		return nil, fmt.Errorf("issue #%d not found", id)
	}

	comment := models.IssueComment{
		ID:        len(issue.Comments) + 1,
		IssueID:   id,
		Author:    author,
		Body:      body,
		CreatedAt: time.Now(),
	}

	issue.Comments = append(issue.Comments, comment)
	issue.UpdatedAt = time.Now()
	_ = s.save(issue)
	return &comment, nil
}
