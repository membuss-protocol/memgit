package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/membuss-protocol/memgit/pkg/gitengine"
	"github.com/membuss-protocol/memgit/pkg/identity"
	"github.com/membuss-protocol/memgit/pkg/membuss"
	"github.com/membuss-protocol/memgit/pkg/models"
	"github.com/membuss-protocol/memgit/pkg/services"
	"github.com/membuss-protocol/memgit/pkg/swarm"
)

// Handler contains all REST API handler methods.
type Handler struct {
	engine         *gitengine.Engine
	repoService    *services.RepoService
	issueService   *services.IssueService
	prService      *services.PRService
	releaseService *services.ReleaseService
	membussClient  *membuss.Client
	idManager      *identity.Manager
	catalog        *swarm.Catalog
	serverBaseURL  string
}

// NewHandler creates a new API handler.
func NewHandler(
	engine *gitengine.Engine,
	repoService *services.RepoService,
	issueService *services.IssueService,
	prService *services.PRService,
	releaseService *services.ReleaseService,
	membussClient *membuss.Client,
	idManager *identity.Manager,
	catalog *swarm.Catalog,
	serverBaseURL string,
) *Handler {
	return &Handler{
		engine:         engine,
		repoService:    repoService,
		issueService:   issueService,
		prService:      prService,
		releaseService: releaseService,
		membussClient:  membussClient,
		idManager:      idManager,
		catalog:        catalog,
		serverBaseURL:  serverBaseURL,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"ok":    false,
		"error": message,
	})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}

// GetCurrentUser returns the profile of the current active user.
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := h.idManager.CurrentUser()
	user.RepoCount = len(h.repoService.ListRepos())
	writeSuccess(w, user)
}

// UpdateCurrentUserProfile updates the current user's profile information.
func (h *Handler) UpdateCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
		Email       string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.idManager.UpdateProfile(req.Username, req.DisplayName, req.Bio, req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, updated)
}

// GetUserProfile returns a public user profile and their repositories.
func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	user, exists := h.idManager.GetUser(username)
	if !exists {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Collect repos owned by this user
	allRepos := h.catalog.AllRepositories()
	userRepos := make([]*models.Repository, 0)
	starredRepos := make([]*models.Repository, 0)

	for _, repo := range allRepos {
		if repo.Owner == username {
			userRepos = append(userRepos, repo)
		}
		for _, pk := range repo.StarredBy {
			if pk == user.PublicKey {
				starredRepos = append(starredRepos, repo)
				break
			}
		}
	}

	user.RepoCount = len(userRepos)
	user.StarCount = len(starredRepos)

	writeSuccess(w, map[string]interface{}{
		"user":          user,
		"repositories":  userRepos,
		"starred_repos": starredRepos,
	})
}

// ListUsers returns all known developers on the swarm.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users := h.idManager.ListUsers()
	writeSuccess(w, users)
}

// ExploreRepos returns all public repositories across the entire global Membuss swarm.
func (h *Handler) ExploreRepos(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	filter := r.URL.Query().Get("filter") // "all", "local", "trending", "starred"

	allRepos := h.catalog.AllRepositories()
	currentUser := h.idManager.CurrentUser()
	filtered := make([]*models.Repository, 0)

	for _, repo := range allRepos {
		if filter == "local" && !repo.IsLocal {
			continue
		}
		if filter == "starred" {
			hasStar := false
			for _, pk := range repo.StarredBy {
				if pk == currentUser.PublicKey {
					hasStar = true
					break
				}
			}
			if !hasStar {
				continue
			}
		}

		if query != "" {
			nameMatch := strings.Contains(strings.ToLower(repo.Name), query)
			ownerMatch := strings.Contains(strings.ToLower(repo.Owner), query)
			descMatch := strings.Contains(strings.ToLower(repo.Description), query)
			topicMatch := false
			for _, t := range repo.Topics {
				if strings.Contains(strings.ToLower(t), query) {
					topicMatch = true
					break
				}
			}
			if !nameMatch && !ownerMatch && !descMatch && !topicMatch {
				continue
			}
		}

		filtered = append(filtered, repo)
	}

	writeSuccess(w, filtered)
}

// GetActivityFeed returns the global activity stream.
func (h *Handler) GetActivityFeed(w http.ResponseWriter, r *http.Request) {
	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	feed := h.catalog.ActivityFeed(limit)
	writeSuccess(w, feed)
}

// ListRepos returns local repositories.
func (h *Handler) ListRepos(w http.ResponseWriter, r *http.Request) {
	repos := h.repoService.ListRepos()
	currentUser := h.idManager.CurrentUser()
	for _, repo := range repos {
		if repo.Owner == "" {
			repo.Owner = currentUser.Username
		}
		if repo.FullName == "" {
			repo.FullName = repo.Owner + "/" + repo.Name
		}
		repo.IsLocal = true
	}
	writeSuccess(w, repos)
}

// CreateRepo creates a new repository.
func (h *Handler) CreateRepo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		DefaultBranch string `json:"default_branch"`
		IsPrivate     bool   `json:"is_private"`
		InitReadme    bool   `json:"init_readme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "repository name is required")
		return
	}

	repo, err := h.repoService.CreateRepo(r.Context(), req.Name, req.Description, req.DefaultBranch, req.IsPrivate, req.InitReadme, h.serverBaseURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	currentUser := h.idManager.CurrentUser()
	repo.Owner = currentUser.Username
	repo.FullName = currentUser.Username + "/" + repo.Name
	repo.IsLocal = true
	_ = h.repoService.Save(repo)

	h.catalog.RecordActivity("create_repo", currentUser.Username, repo.FullName, fmt.Sprintf("Created repository %s", repo.FullName))

	writeSuccess(w, repo)
}

// GetRepo returns a single repository's details.
func (h *Handler) GetRepo(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "repo")
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}

	repo, err := h.repoService.GetRepo(name)
	if err != nil {
		// Check swarm catalog for remote repo
		all := h.catalog.AllRepositories()
		for _, remote := range all {
			if remote.Name == name || remote.FullName == name {
				writeSuccess(w, remote)
				return
			}
		}
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	currentUser := h.idManager.CurrentUser()
	if repo.Owner == "" {
		repo.Owner = currentUser.Username
	}
	if repo.FullName == "" {
		repo.FullName = repo.Owner + "/" + repo.Name
	}
	repo.IsLocal = true
	writeSuccess(w, repo)
}

// StarRepo toggles a star on a repository.
func (h *Handler) StarRepo(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	repo, isStarred, err := h.catalog.ToggleStar(repoName)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, map[string]interface{}{
		"repo":       repo,
		"is_starred": isStarred,
		"star_count": repo.StarCount,
	})
}

// ForkRepo creates a decentralized fork.
func (h *Handler) ForkRepo(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	var req struct {
		NewName string `json:"new_name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	forked, err := h.catalog.ForkRepo(r.Context(), repoName, req.NewName, h.serverBaseURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, forked)
}

// SyncRepo triggers manual Merkle DAG snapshot to Membuss.
func (h *Handler) SyncRepo(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "repo")
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}

	mid, err := h.repoService.SyncToMembuss(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"repo":      name,
		"latest_mid": mid,
		"status":    "synced",
	})
}

// GetTree returns a directory listing.
func (h *Handler) GetTree(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	ref := chi.URLParam(r, "ref")
	subPath := chi.URLParam(r, "*")

	nodes, err := h.engine.GetTree(repoName, ref, subPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, nodes)
}

// GetBlob returns file content.
func (h *Handler) GetBlob(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	ref := chi.URLParam(r, "ref")
	filePath := chi.URLParam(r, "*")

	blob, err := h.engine.GetBlob(repoName, ref, filePath)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(w, blob)
}

// GetRawBlob returns raw file bytes.
func (h *Handler) GetRawBlob(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	ref := chi.URLParam(r, "ref")
	filePath := chi.URLParam(r, "*")

	data, mimeType, err := h.engine.GetRawBlob(repoName, ref, filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GetCommits returns commit history.
func (h *Handler) GetCommits(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	ref := chi.URLParam(r, "ref")

	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	commits, err := h.engine.GetCommits(repoName, ref, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, commits)
}

// GetCommit returns a specific commit and its file diffs.
func (h *Handler) GetCommit(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	sha := chi.URLParam(r, "sha")

	diff, err := h.engine.GetCommitDiff(repoName, sha)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(w, diff)
}

// GetBranches returns all branches.
func (h *Handler) GetBranches(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	branches, err := h.engine.GetBranches(repoName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, branches)
}

// CreateBranch creates a branch.
func (h *Handler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	var req struct {
		Name      string `json:"name"`
		TargetSHA string `json:"target_sha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.engine.CreateBranch(repoName, req.Name, req.TargetSHA); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, map[string]string{"branch": req.Name, "status": "created"})
}

// GetTags returns all tags.
func (h *Handler) GetTags(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	tags, err := h.engine.GetTags(repoName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, tags)
}

// ListIssues returns issues for a repo.
func (h *Handler) ListIssues(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	state := r.URL.Query().Get("state")
	issues := h.issueService.ListIssues(repoName, state)
	writeSuccess(w, issues)
}

// CreateIssue creates an issue.
func (h *Handler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	var req struct {
		Title  string   `json:"title"`
		Body   string   `json:"body"`
		Labels []string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	currentUser := h.idManager.CurrentUser()
	issue, err := h.issueService.CreateIssue(repoName, req.Title, req.Body, currentUser.Username, req.Labels)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.catalog.RecordActivity("issue", currentUser.Username, repoName, fmt.Sprintf("Opened issue #%d: %s", issue.ID, issue.Title))

	writeSuccess(w, issue)
}

// GetIssue returns a single issue.
func (h *Handler) GetIssue(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid issue id")
		return
	}

	issue, err := h.issueService.GetIssue(repoName, id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(w, issue)
}

// AddIssueComment adds a comment.
func (h *Handler) AddIssueComment(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid issue id")
		return
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	currentUser := h.idManager.CurrentUser()
	comment, err := h.issueService.AddComment(repoName, id, currentUser.Username, req.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, comment)
}

// UpdateIssueState updates state (open/closed).
func (h *Handler) UpdateIssueState(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid issue id")
		return
	}

	var req struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	issue, err := h.issueService.UpdateIssueState(repoName, id, req.State)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, issue)
}

// ListPRs returns PRs.
func (h *Handler) ListPRs(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	state := r.URL.Query().Get("state")
	prs := h.prService.ListPRs(repoName, state)
	writeSuccess(w, prs)
}

// CreatePR creates a PR.
func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	var req struct {
		Title        string `json:"title"`
		Body         string `json:"body"`
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
		SourceRepo   string `json:"source_repo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sourceRepo := req.SourceRepo
	if sourceRepo == "" {
		sourceRepo = repoName
	}

	currentUser := h.idManager.CurrentUser()
	pr, err := h.prService.CreatePR(repoName, req.Title, req.Body, currentUser.Username, req.SourceBranch, req.TargetBranch, sourceRepo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.catalog.RecordActivity("pr", currentUser.Username, repoName, fmt.Sprintf("Opened Pull Request #%d: %s", pr.ID, pr.Title))

	writeSuccess(w, pr)
}

// GetPR returns a single PR.
func (h *Handler) GetPR(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pr id")
		return
	}

	pr, err := h.prService.GetPR(repoName, id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(w, pr)
}

// MergePR merges a PR.
func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pr id")
		return
	}

	pr, err := h.prService.MergePR(repoName, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	currentUser := h.idManager.CurrentUser()
	h.catalog.RecordActivity("pr_merged", currentUser.Username, repoName, fmt.Sprintf("Merged Pull Request #%d into %s", pr.ID, pr.TargetBranch))

	// Sync updated merged tree to Membuss
	_, _ = h.repoService.SyncToMembuss(r.Context(), repoName)

	writeSuccess(w, pr)
}

// ListReleases returns releases.
func (h *Handler) ListReleases(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	releases := h.releaseService.ListReleases(repoName)
	writeSuccess(w, releases)
}

// CreateRelease creates a tagged release.
func (h *Handler) CreateRelease(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	var req struct {
		TagName     string `json:"tag_name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		TargetSHA   string `json:"target_sha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Ensure tag exists
	_ = h.engine.CreateTag(repoName, req.TagName, req.TargetSHA, req.Title)

	// Snapshot to Membuss
	mid, _ := h.repoService.SyncToMembuss(r.Context(), repoName)

	rel, err := h.releaseService.CreateRelease(repoName, req.TagName, req.Title, req.Description, req.TargetSHA, mid, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	currentUser := h.idManager.CurrentUser()
	h.catalog.RecordActivity("release", currentUser.Username, repoName, fmt.Sprintf("Published release %s: %s", rel.TagName, rel.Title))

	writeSuccess(w, rel)
}

// GetSwarmStatus returns P2P health metrics from Membuss.
func (h *Handler) GetSwarmStatus(w http.ResponseWriter, r *http.Request) {
	repoName := chi.URLParam(r, "repo")
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		repoName = parts[len(parts)-1]
	}

	repo, _ := h.repoService.GetRepo(repoName)
	nodeInfo, _ := h.membussClient.CheckHealth(r.Context())
	peers, _ := h.membussClient.GetPeers(r.Context())

	status := models.SwarmStatus{
		ConnectedPeers: len(peers),
		DHTProviders:   len(peers) + 1,
		AnchorSynced:   true,
		ErasureParity:  "10+4 Reed-Solomon Parity Enabled",
	}

	if nodeInfo != nil {
		status.NodeID = nodeInfo.ID
		status.Addresses = nodeInfo.Addresses
	}

	if repo != nil {
		status.ReplicationList = []models.PeerReplication{
			{
				PeerID:    status.NodeID,
				Latency:   "0ms (Local Host)",
				IsAnchor:  false,
				HasShards: true,
			},
		}
		for _, p := range peers {
			status.ReplicationList = append(status.ReplicationList, models.PeerReplication{
				PeerID:    p.ID,
				Latency:   p.Latency,
				IsAnchor:  strings.Contains(p.ID, "anchor"),
				HasShards: true,
			})
		}
	}

	writeSuccess(w, status)
}

// SystemStatus returns overall health of MemGit and Membuss daemon connection.
func (h *Handler) SystemStatus(w http.ResponseWriter, r *http.Request) {
	nodeInfo, err := h.membussClient.CheckHealth(r.Context())
	online := err == nil
	allRepos := h.catalog.AllRepositories()
	status := map[string]interface{}{
		"memgit_version":  "2.0.0",
		"membuss_online":  online,
		"total_repos":     len(allRepos),
		"local_repos":     len(h.repoService.ListRepos()),
		"storage_backend": "Membuss P2P (BadgerDB + 10+4 Reed-Solomon)",
		"current_user":    h.idManager.CurrentUser(),
	}
	if nodeInfo != nil {
		status["membuss_node_id"] = nodeInfo.ID
		status["membuss_version"] = nodeInfo.Version
	}
	writeSuccess(w, status)
}
