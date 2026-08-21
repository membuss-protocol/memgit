package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/membuss-protocol/memgit/pkg/gitengine"
	"github.com/membuss-protocol/memgit/pkg/membuss"
)

func TestServices_CRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "memgit-services-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Mock Membuss Node API server for testing
	mockMembuss := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/memns/key/generate" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok": true,
				"data": map[string]interface{}{
					"name":       "test-key",
					"memns_name": "memns1ztestkey123456",
					"created_at": time.Now(),
				},
			})
			return
		}
		if r.URL.Path == "/api/v1/add/dir" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok": true,
				"data": map[string]interface{}{
					"mid":  "mem1z4a2b3c4d5e6f7",
					"size": 1024,
				},
			})
			return
		}
		if r.URL.Path == "/api/v1/memns/publish" || r.URL.Path == "/api/v1/seal" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer mockMembuss.Close()

	client := membuss.NewClient(mockMembuss.URL, mockMembuss.URL)
	engine, err := gitengine.NewEngine(filepath.Join(tempDir, "repos"), client)
	if err != nil {
		t.Fatal(err)
	}

	repoService, err := NewRepoService(filepath.Join(tempDir, "meta"), engine, client)
	if err != nil {
		t.Fatal(err)
	}

	issueService, err := NewIssueService(filepath.Join(tempDir, "issues"))
	if err != nil {
		t.Fatal(err)
	}

	prService, err := NewPRService(filepath.Join(tempDir, "pulls"), engine)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// 1. Create Repository
	repo, err := repoService.CreateRepo(ctx, "awesome-p2p", "Decentralized P2P App", "main", false, true, "http://localhost:8500")
	if err != nil {
		t.Fatalf("CreateRepo failed: %v", err)
	}
	if repo.Name != "awesome-p2p" {
		t.Errorf("expected name 'awesome-p2p', got %q", repo.Name)
	}
	if repo.LatestMID != "mem1z4a2b3c4d5e6f7" {
		t.Errorf("expected LatestMID 'mem1z4a2b3c4d5e6f7', got %q", repo.LatestMID)
	}

	// 2. Star Repository
	stars, err := repoService.StarRepo("awesome-p2p")
	if err != nil || stars != 1 {
		t.Fatalf("StarRepo failed: %v, stars=%d", err, stars)
	}

	// 3. Issue Lifecycle
	issue, err := issueService.CreateIssue("awesome-p2p", "Bug in P2P stream", "Stream hangs when peer drops", "Alice", []string{"bug"})
	if err != nil {
		t.Fatalf("CreateIssue failed: %v", err)
	}
	if issue.ID != 1 || issue.State != "open" {
		t.Errorf("unexpected issue state: %+v", issue)
	}

	comment, err := issueService.AddComment("awesome-p2p", 1, "Bob", "Fix available on branch patch-1")
	if err != nil || comment.ID != 1 {
		t.Fatalf("AddComment failed: %v", err)
	}

	_, err = issueService.UpdateIssueState("awesome-p2p", 1, "closed")
	if err != nil {
		t.Fatalf("UpdateIssueState failed: %v", err)
	}

	// 4. PR Lifecycle
	pr, err := prService.CreatePR("awesome-p2p", "Fix peer stream drops", "Added keep-alive heartbeats", "Bob", "main", "main", "awesome-p2p")
	if err != nil {
		t.Fatalf("CreatePR failed: %v", err)
	}
	if pr.ID != 1 || pr.State != "open" {
		t.Errorf("unexpected PR state: %+v", pr)
	}

	mergedPR, err := prService.MergePR("awesome-p2p", 1)
	if err != nil {
		t.Fatalf("MergePR failed: %v", err)
	}
	if mergedPR.State != "merged" {
		t.Errorf("expected state 'merged', got %q", mergedPR.State)
	}
}
