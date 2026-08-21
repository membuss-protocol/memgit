package gitengine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/membuss-protocol/memgit/pkg/membuss"
)

func TestGitEngine_InitAndCommit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "memgit-engine-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	client := membuss.NewClient("http://127.0.0.1:5001", "http://127.0.0.1:8080")
	engine, err := NewEngine(filepath.Join(tempDir, "repos"), client)
	if err != nil {
		t.Fatal(err)
	}

	repoName := "test-repo"
	_, err = engine.InitRepo(repoName, "main", "Test repository on Membuss", true)
	if err != nil {
		t.Fatalf("InitRepo failed: %v", err)
	}

	// Verify branches
	branches, err := engine.GetBranches(repoName)
	if err != nil {
		t.Fatalf("GetBranches failed: %v", err)
	}
	if len(branches) == 0 {
		t.Fatalf("expected at least 1 branch, got 0")
	}
	if branches[0].Name != "main" {
		t.Errorf("expected default branch 'main', got %q", branches[0].Name)
	}

	// Verify commits
	commits, err := engine.GetCommits(repoName, "main", 10)
	if err != nil {
		t.Fatalf("GetCommits failed: %v", err)
	}
	if len(commits) == 0 {
		t.Fatalf("expected at least 1 commit, got 0")
	}

	// Verify tree listing
	tree, err := engine.GetTree(repoName, "main", "")
	if err != nil {
		t.Fatalf("GetTree failed: %v", err)
	}
	if len(tree) != 2 {
		t.Fatalf("expected 2 files (README.md, .gitignore), got %d", len(tree))
	}

	// Verify blob reading
	blob, err := engine.GetBlob(repoName, "main", "README.md")
	if err != nil {
		t.Fatalf("GetBlob failed: %v", err)
	}
	if blob.IsBinary {
		t.Errorf("README.md should not be binary")
	}
	if blob.Size == 0 {
		t.Errorf("README.md size should be > 0")
	}

	// Test branch creation
	err = engine.CreateBranch(repoName, "feature-x", commits[0].SHA)
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	branches, _ = engine.GetBranches(repoName)
	if len(branches) != 2 {
		t.Errorf("expected 2 branches after creation, got %d", len(branches))
	}
}
