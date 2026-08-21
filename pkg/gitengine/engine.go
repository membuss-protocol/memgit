package gitengine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/membuss-protocol/memgit/pkg/membuss"
	"github.com/membuss-protocol/memgit/pkg/models"
)

// Engine manages Git repository instances and interfaces with Membuss.
type Engine struct {
	baseDir string
	client  *membuss.Client
	mu      sync.RWMutex
}

// NewEngine creates a new Git engine instance.
func NewEngine(baseDir string, client *membuss.Client) (*Engine, error) {
	if baseDir == "" {
		baseDir = "data/repos"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	return &Engine{
		baseDir: baseDir,
		client:  client,
	}, nil
}

// RepoPath returns the on-disk bare repository path.
func (e *Engine) RepoPath(repoName string) string {
	cleanName := strings.TrimSuffix(repoName, ".git")
	return filepath.Join(e.baseDir, cleanName+".git")
}

// InitRepo creates a new bare Git repository and initializes it with an initial commit if requested.
func (e *Engine) InitRepo(repoName, defaultBranch, desc string, initReadme bool) (*git.Repository, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	repoPath := e.RepoPath(repoName)
	if _, err := os.Stat(repoPath); err == nil {
		return nil, fmt.Errorf("repository %q already exists", repoName)
	}

	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// Initialize bare repo
	repo, err := git.PlainInit(repoPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to init bare repo: %w", err)
	}

	// Set HEAD to default branch
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName(defaultBranch))
	if err := repo.Storer.SetReference(headRef); err != nil {
		return nil, err
	}

	if initReadme {
		// Create initial README and .gitignore via temporary worktree in memory
		if err := e.createInitialCommit(repo, repoName, defaultBranch, desc); err != nil {
			return nil, fmt.Errorf("failed to create initial commit: %w", err)
		}
	}

	return repo, nil
}

// createInitialCommit creates an initial commit on the bare repository.
func (e *Engine) createInitialCommit(repo *git.Repository, repoName, defaultBranch, desc string) error {
	readmeContent := fmt.Sprintf("# %s\n\n%s\n\n---\n*Decentralized repository hosted on [Membuss](https://membuss.network).*\n", repoName, desc)
	gitignoreContent := "# Build\nbin/\n*.exe\n*.out\n*.log\n\n# Node / Web\nnode_modules/\ndist/\n\n# OS & IDE\n.DS_Store\nThumbs.db\n.vscode/\n.idea/\n"

	// Create blobs
	readmeBlob := repo.Storer.NewEncodedObject()
	readmeBlob.SetType(plumbing.BlobObject)
	w, err := readmeBlob.Writer()
	if err != nil {
		return err
	}
	_, _ = w.Write([]byte(readmeContent))
	_ = w.Close()
	readmeSHA, err := repo.Storer.SetEncodedObject(readmeBlob)
	if err != nil {
		return err
	}

	gitignoreBlob := repo.Storer.NewEncodedObject()
	gitignoreBlob.SetType(plumbing.BlobObject)
	w2, err := gitignoreBlob.Writer()
	if err != nil {
		return err
	}
	_, _ = w2.Write([]byte(gitignoreContent))
	_ = w2.Close()
	gitignoreSHA, err := repo.Storer.SetEncodedObject(gitignoreBlob)
	if err != nil {
		return err
	}

	// Create root tree
	tree := object.Tree{
		Entries: []object.TreeEntry{
			{
				Name: ".gitignore",
				Mode: filemode.Regular,
				Hash: gitignoreSHA,
			},
			{
				Name: "README.md",
				Mode: filemode.Regular,
				Hash: readmeSHA,
			},
		},
	}
	treeObj := repo.Storer.NewEncodedObject()
	if err := tree.Encode(treeObj); err != nil {
		return err
	}
	treeSHA, err := repo.Storer.SetEncodedObject(treeObj)
	if err != nil {
		return err
	}

	// Create commit
	now := time.Now()
	commit := object.Commit{
		Author: object.Signature{
			Name:  "Membuss MemGit",
			Email: "memgit@membuss.network",
			When:  now,
		},
		Committer: object.Signature{
			Name:  "Membuss MemGit",
			Email: "memgit@membuss.network",
			When:  now,
		},
		Message:  "Initial commit\n\nCreated with MemGit on Membuss decentralized network.",
		TreeHash: treeSHA,
	}

	commitObj := repo.Storer.NewEncodedObject()
	if err := commit.Encode(commitObj); err != nil {
		return err
	}
	commitSHA, err := repo.Storer.SetEncodedObject(commitObj)
	if err != nil {
		return err
	}

	// Update branch reference
	branchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(defaultBranch), commitSHA)
	return repo.Storer.SetReference(branchRef)
}

// OpenRepo opens an existing bare repository.
func (e *Engine) OpenRepo(repoName string) (*git.Repository, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	repoPath := e.RepoPath(repoName)
	return git.PlainOpen(repoPath)
}

// ResolveRef resolves a branch name, tag, or SHA into a commit object.
func (e *Engine) ResolveRef(repo *git.Repository, refName string) (*object.Commit, error) {
	if refName == "" || refName == "HEAD" {
		head, err := repo.Head()
		if err != nil {
			return nil, err
		}
		return repo.CommitObject(head.Hash())
	}

	// Check if ref is a commit SHA
	if plumbing.IsHash(refName) {
		h := plumbing.NewHash(refName)
		if c, err := repo.CommitObject(h); err == nil {
			return c, nil
		}
	}

	// Check branch
	branchRef, err := repo.Reference(plumbing.NewBranchReferenceName(refName), true)
	if err == nil {
		return repo.CommitObject(branchRef.Hash())
	}

	// Check tag
	tagRef, err := repo.Reference(plumbing.NewTagReferenceName(refName), true)
	if err == nil {
		return repo.CommitObject(tagRef.Hash())
	}

	return nil, fmt.Errorf("reference %q not found", refName)
}

// GetCommits returns a list of commits starting from the given ref.
func (e *Engine) GetCommits(repoName, refName string, limit int) ([]models.Commit, error) {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return nil, err
	}

	commit, err := e.ResolveRef(repo, refName)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	var results []models.Commit
	iter := object.NewCommitPreorderIter(commit, nil, nil)
	defer iter.Close()

	count := 0
	_ = iter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return errors.New("limit reached")
		}
		count++

		parents := make([]string, len(c.ParentHashes))
		for i, p := range c.ParentHashes {
			parents[i] = p.String()
		}

		mc := models.Commit{
			SHA:      c.Hash.String(),
			ShortSHA: c.Hash.String()[:8],
			TreeSHA:  c.TreeHash.String(),
			Parents:  parents,
			Author: models.Signature{
				Name:  c.Author.Name,
				Email: c.Author.Email,
				When:  c.Author.When,
			},
			Committer: models.Signature{
				Name:  c.Committer.Name,
				Email: c.Committer.Email,
				When:  c.Committer.When,
			},
			Message:   strings.TrimSpace(c.Message),
			Timestamp: c.Author.When,
		}
		results = append(results, mc)
		return nil
	})

	return results, nil
}

// GetCommitDiff returns single commit details with full file diffs.
func (e *Engine) GetCommitDiff(repoName, sha string) (*models.CommitDiff, error) {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return nil, err
	}

	commitObj, err := repo.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		return nil, err
	}

	currentTree, err := commitObj.Tree()
	if err != nil {
		return nil, err
	}

	var prevTree *object.Tree
	if commitObj.NumParents() > 0 {
		parentCommit, perr := commitObj.Parent(0)
		if perr == nil {
			prevTree, _ = parentCommit.Tree()
		}
	}

	changes, err := object.DiffTree(prevTree, currentTree)
	if err != nil {
		return nil, err
	}

	var fileDiffs []models.FileDiff
	totalAdditions := 0
	totalDeletions := 0

	for _, change := range changes {
		action, _ := change.Action()
		status := "modified"
		switch action {
		case 1:
			status = "added"
		case 2:
			status = "deleted"
		case 3:
			status = "modified"
		}

		patch, _ := change.Patch()
		patchStr := ""
		additions := 0
		deletions := 0
		if patch != nil {
			patchStr = patch.String()
			for _, stat := range patch.Stats() {
				additions += stat.Addition
				deletions += stat.Deletion
			}
		}

		totalAdditions += additions
		totalDeletions += deletions

		fileDiffs = append(fileDiffs, models.FileDiff{
			OldPath:   change.From.Name,
			NewPath:   change.To.Name,
			Status:    status,
			Additions: additions,
			Deletions: deletions,
			Patch:     patchStr,
		})
	}

	parents := make([]string, len(commitObj.ParentHashes))
	for i, p := range commitObj.ParentHashes {
		parents[i] = p.String()
	}

	mc := models.Commit{
		SHA:      commitObj.Hash.String(),
		ShortSHA: commitObj.Hash.String()[:8],
		TreeSHA:  commitObj.TreeHash.String(),
		Parents:  parents,
		Author: models.Signature{
			Name:  commitObj.Author.Name,
			Email: commitObj.Author.Email,
			When:  commitObj.Author.When,
		},
		Committer: models.Signature{
			Name:  commitObj.Committer.Name,
			Email: commitObj.Committer.Email,
			When:  commitObj.Committer.When,
		},
		Message:   strings.TrimSpace(commitObj.Message),
		Timestamp: commitObj.Author.When,
	}
	mc.Stats.FilesChanged = len(fileDiffs)
	mc.Stats.Additions = totalAdditions
	mc.Stats.Deletions = totalDeletions

	return &models.CommitDiff{
		Commit: mc,
		Files:  fileDiffs,
	}, nil
}

// GetTree returns a directory listing under subPath for a given ref.
func (e *Engine) GetTree(repoName, refName, subPath string) ([]models.FileNode, error) {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return nil, err
	}

	commit, err := e.ResolveRef(repo, refName)
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	subPath = strings.Trim(subPath, "/")
	targetTree := tree
	if subPath != "" {
		targetTree, err = tree.Tree(subPath)
		if err != nil {
			return nil, fmt.Errorf("path %q is not a directory: %w", subPath, err)
		}
	}

	var nodes []models.FileNode
	for _, entry := range targetTree.Entries {
		entryPath := entry.Name
		if subPath != "" {
			entryPath = subPath + "/" + entry.Name
		}

		nodeType := "blob"
		size := int64(0)
		if entry.Mode == filemode.Dir {
			nodeType = "tree"
		} else {
			if blob, berr := repo.BlobObject(entry.Hash); berr == nil {
				size = blob.Size
			}
		}

		nodes = append(nodes, models.FileNode{
			Path:    entryPath,
			Name:    entry.Name,
			Type:    nodeType,
			Size:    size,
			Mode:    entry.Mode.String(),
			BlobSHA: entry.Hash.String(),
		})
	}

	return nodes, nil
}

// GetBlob returns file content and metadata for a specific path.
func (e *Engine) GetBlob(repoName, refName, filePath string) (*models.BlobContent, error) {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return nil, err
	}

	commit, err := e.ResolveRef(repo, refName)
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	filePath = strings.Trim(filePath, "/")
	file, err := tree.File(filePath)
	if err != nil {
		return nil, fmt.Errorf("file %q not found: %w", filePath, err)
	}

	content, err := file.Contents()
	if err != nil {
		return nil, err
	}

	// Sniff binary vs text
	isBinary := false
	var sniffBuf [512]byte
	n := copy(sniffBuf[:], []byte(content))
	mimeType := http.DetectContentType(sniffBuf[:n])
	if strings.Contains(mimeType, "application/octet-stream") || bytes.IndexByte([]byte(content), 0) != -1 {
		isBinary = true
		content = "" // Don't send raw binary over JSON
	}

	return &models.BlobContent{
		Path:     filePath,
		Name:     filepath.Base(filePath),
		Size:     file.Size,
		IsBinary: isBinary,
		Content:  content,
		MimeType: mimeType,
		BlobSHA:  file.Hash.String(),
	}, nil
}

// GetRawBlob returns raw bytes and detected mime-type for a file.
func (e *Engine) GetRawBlob(repoName, refName, filePath string) ([]byte, string, error) {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return nil, "", err
	}

	commit, err := e.ResolveRef(repo, refName)
	if err != nil {
		return nil, "", err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, "", err
	}

	filePath = strings.Trim(filePath, "/")
	file, err := tree.File(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("file %q not found: %w", filePath, err)
	}

	reader, err := file.Reader()
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, "", err
	}

	data := buf.Bytes()
	var sniffBuf [512]byte
	n := copy(sniffBuf[:], data)
	mimeType := http.DetectContentType(sniffBuf[:n])

	return data, mimeType, nil
}

// GetBranches returns all local branch references.
func (e *Engine) GetBranches(repoName string) ([]models.Branch, error) {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return nil, err
	}

	head, _ := repo.Head()
	headName := ""
	if head != nil && head.Name().IsBranch() {
		headName = head.Name().Short()
	}

	iter, err := repo.Branches()
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var branches []models.Branch
	_ = iter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		commitSHA := ref.Hash().String()
		msg := ""
		updatedAt := time.Now()

		if c, err := repo.CommitObject(ref.Hash()); err == nil {
			msg = strings.Split(strings.TrimSpace(c.Message), "\n")[0]
			updatedAt = c.Author.When
		}

		branches = append(branches, models.Branch{
			Name:      name,
			CommitSHA: commitSHA,
			IsDefault: name == headName,
			Message:   msg,
			UpdatedAt: updatedAt,
		})
		return nil
	})

	return branches, nil
}

// CreateBranch creates a new branch pointing to targetSHA or HEAD.
func (e *Engine) CreateBranch(repoName, branchName, targetSHA string) error {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return err
	}

	var hash plumbing.Hash
	if targetSHA != "" {
		hash = plumbing.NewHash(targetSHA)
	} else {
		head, err := repo.Head()
		if err != nil {
			return err
		}
		hash = head.Hash()
	}

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), hash)
	return repo.Storer.SetReference(ref)
}

// DeleteBranch deletes a branch reference.
func (e *Engine) DeleteBranch(repoName, branchName string) error {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return err
	}
	return repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(branchName))
}

// GetTags returns all tags in the repository.
func (e *Engine) GetTags(repoName string) ([]models.Tag, error) {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return nil, err
	}

	iter, err := repo.Tags()
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var tags []models.Tag
	_ = iter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		sha := ref.Hash().String()
		msg := ""
		createdAt := time.Now()

		if tagObj, err := repo.TagObject(ref.Hash()); err == nil {
			msg = tagObj.Message
			createdAt = tagObj.Tagger.When
			sha = tagObj.Target.String()
		} else if commitObj, err := repo.CommitObject(ref.Hash()); err == nil {
			msg = strings.Split(strings.TrimSpace(commitObj.Message), "\n")[0]
			createdAt = commitObj.Author.When
		}

		tags = append(tags, models.Tag{
			Name:      name,
			CommitSHA: sha,
			Message:   msg,
			CreatedAt: createdAt,
		})
		return nil
	})

	return tags, nil
}

// CreateTag creates a new Git tag reference.
func (e *Engine) CreateTag(repoName, tagName, targetSHA, message string) error {
	repo, err := e.OpenRepo(repoName)
	if err != nil {
		return err
	}

	var hash plumbing.Hash
	if targetSHA != "" {
		hash = plumbing.NewHash(targetSHA)
	} else {
		head, err := repo.Head()
		if err != nil {
			return err
		}
		hash = head.Hash()
	}

	if message != "" {
		tagObj := &object.Tag{
			Name: tagName,
			Tagger: object.Signature{
				Name:  "Membuss MemGit",
				Email: "memgit@membuss.network",
				When:  time.Now(),
			},
			Message:    message,
			TargetType: plumbing.CommitObject,
			Target:     hash,
		}
		obj := repo.Storer.NewEncodedObject()
		if err := tagObj.Encode(obj); err != nil {
			return err
		}
		tagHash, err := repo.Storer.SetEncodedObject(obj)
		if err != nil {
			return err
		}
		ref := plumbing.NewHashReference(plumbing.NewTagReferenceName(tagName), tagHash)
		return repo.Storer.SetReference(ref)
	}

	ref := plumbing.NewHashReference(plumbing.NewTagReferenceName(tagName), hash)
	return repo.Storer.SetReference(ref)
}

// SnapshotToMembuss walks the bare repo on disk, packages it, uploads to Membuss MemFS,
// and signs the update to the repository's MemNS key.
func (e *Engine) SnapshotToMembuss(ctx context.Context, repoName, memnsKey string) (string, error) {
	e.mu.RLock()
	repoPath := e.RepoPath(repoName)
	e.mu.RUnlock()

	// Update Dumb HTTP server-info so any Membuss Gateway can serve `git clone http://<gateway>/memns/<name>`
	_ = exec.CommandContext(ctx, "git", "-C", repoPath, "update-server-info").Run()

	var parts []membuss.DirectoryPart
	err := filepath.WalkDir(repoPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(repoPath, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		parts = append(parts, membuss.DirectoryPart{
			Path:    rel,
			Content: content,
		})
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to package bare repo: %w", err)
	}

	// Ingest directory into Membuss
	rootMID, _, err := e.client.IngestDirectory(ctx, repoName, parts)
	if err != nil {
		return "", fmt.Errorf("failed to ingest repo to Membuss: %w", err)
	}

	// Publish to MemNS if key is provided
	if memnsKey != "" {
		if err := e.client.PublishMemNS(ctx, memnsKey, rootMID, 7*24*time.Hour); err != nil {
			return rootMID, fmt.Errorf("ingested MID %s but MemNS publish failed: %w", rootMID, err)
		}
	}

	return rootMID, nil
}
