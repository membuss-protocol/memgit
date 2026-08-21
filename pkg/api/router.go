package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// SetupRouter constructs the full Chi HTTP router for MemGit.
func SetupRouter(gitHandler *GitHTTPHandler, apiHandler *Handler, webDir string) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Enable CORS for web frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Range"},
		ExposedHeaders:   []string{"Link", "Content-Length", "Content-Range"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Git Smart HTTP Protocol Routes (supports /git/:repo and /git/:owner/:repo)
	r.Route("/git/{repo}", func(gr chi.Router) {
		gr.Get("/info/refs", gitHandler.InfoRefs)
		gr.Post("/git-upload-pack", gitHandler.UploadPack)
		gr.Post("/git-receive-pack", gitHandler.ReceivePack)
	})
	r.Route("/git/{owner}/{repo}", func(gr chi.Router) {
		gr.Get("/info/refs", gitHandler.InfoRefs)
		gr.Post("/git-upload-pack", gitHandler.UploadPack)
		gr.Post("/git-receive-pack", gitHandler.ReceivePack)
	})

	// REST API v1 Routes
	r.Route("/api/v1", func(api chi.Router) {
		// System
		api.Get("/system/status", apiHandler.SystemStatus)

		// User & Identity
		api.Get("/user", apiHandler.GetCurrentUser)
		api.Put("/user/profile", apiHandler.UpdateCurrentUserProfile)
		api.Get("/users", apiHandler.ListUsers)
		api.Get("/users/{username}", apiHandler.GetUserProfile)

		// Global Swarm Discovery & Activity
		api.Get("/explore/repos", apiHandler.ExploreRepos)
		api.Get("/activity/feed", apiHandler.GetActivityFeed)

		// Repositories
		api.Get("/repos", apiHandler.ListRepos)
		api.Post("/repos", apiHandler.CreateRepo)
		api.Get("/repos/{repo}", apiHandler.GetRepo)
		api.Post("/repos/{repo}/star", apiHandler.StarRepo)
		api.Post("/repos/{repo}/fork", apiHandler.ForkRepo)
		api.Post("/repos/{repo}/sync", apiHandler.SyncRepo)
		api.Get("/repos/{repo}/network", apiHandler.GetSwarmStatus)

		// Git Objects
		api.Get("/repos/{repo}/tree/{ref}", apiHandler.GetTree)
		api.Get("/repos/{repo}/tree/{ref}/*", apiHandler.GetTree)
		api.Get("/repos/{repo}/blob/{ref}/*", apiHandler.GetBlob)
		api.Get("/repos/{repo}/raw/{ref}/*", apiHandler.GetRawBlob)
		api.Get("/repos/{repo}/commits/{ref}", apiHandler.GetCommits)
		api.Get("/repos/{repo}/commit/{sha}", apiHandler.GetCommit)
		api.Get("/repos/{repo}/branches", apiHandler.GetBranches)
		api.Post("/repos/{repo}/branches", apiHandler.CreateBranch)
		api.Get("/repos/{repo}/tags", apiHandler.GetTags)

		// Issues
		api.Get("/repos/{repo}/issues", apiHandler.ListIssues)
		api.Post("/repos/{repo}/issues", apiHandler.CreateIssue)
		api.Get("/repos/{repo}/issues/{id}", apiHandler.GetIssue)
		api.Post("/repos/{repo}/issues/{id}/comment", apiHandler.AddIssueComment)
		api.Patch("/repos/{repo}/issues/{id}/state", apiHandler.UpdateIssueState)

		// Pull Requests
		api.Get("/repos/{repo}/pulls", apiHandler.ListPRs)
		api.Post("/repos/{repo}/pulls", apiHandler.CreatePR)
		api.Get("/repos/{repo}/pulls/{id}", apiHandler.GetPR)
		api.Post("/repos/{repo}/pulls/{id}/merge", apiHandler.MergePR)

		// Releases
		api.Get("/repos/{repo}/releases", apiHandler.ListReleases)
		api.Post("/repos/{repo}/releases", apiHandler.CreateRelease)
	})

	// Static Web Frontend Serving (SPA fallback)
	resolvedWebDir := findWebDir(webDir)
	if resolvedWebDir != "" {
		fileServer := http.FileServer(http.Dir(resolvedWebDir))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/git") {
				http.NotFound(w, r)
				return
			}
			path := filepath.Join(resolvedWebDir, r.URL.Path)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				// Fallback to index.html for SPA routing
				http.ServeFile(w, r, filepath.Join(resolvedWebDir, "index.html"))
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}

func findWebDir(webDir string) string {
	candidates := []string{
		webDir,
		"./web/dist",
		"web/dist",
		"../web/dist",
		"../../web/dist",
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "web", "dist"),
			filepath.Join(exeDir, "..", "web", "dist"),
			filepath.Join(exeDir, webDir),
		)
	}

	for _, c := range candidates {
		if c == "" {
			continue
		}
		if info, err := os.Stat(filepath.Join(c, "index.html")); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}
