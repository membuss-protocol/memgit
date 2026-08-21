package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/membuss-protocol/memgit/pkg/gitengine"
	"github.com/membuss-protocol/memgit/pkg/services"
)

// GitHTTPHandler implements standard Git Smart HTTP backend (git-upload-pack, git-receive-pack).
type GitHTTPHandler struct {
	engine      *gitengine.Engine
	repoService *services.RepoService
}

// NewGitHTTPHandler creates a new Git HTTP handler.
func NewGitHTTPHandler(engine *gitengine.Engine, repoService *services.RepoService) *GitHTTPHandler {
	return &GitHTTPHandler{
		engine:      engine,
		repoService: repoService,
	}
}

// InfoRefs handles GET /git/{repo}/info/refs (or /git/{repo}.git/info/refs)
func (h *GitHTTPHandler) InfoRefs(w http.ResponseWriter, r *http.Request) {
	repoParam := chi.URLParam(r, "repo")
	repoName := strings.TrimSuffix(repoParam, ".git")
	service := r.URL.Query().Get("service")

	if service != "git-upload-pack" && service != "git-receive-pack" {
		http.Error(w, "invalid git service (expected git-upload-pack or git-receive-pack)", http.StatusBadRequest)
		return
	}

	repoPath := h.engine.RepoPath(repoName)

	// Push-to-Create Autoprovisioning:
	// If repository doesn't exist and this is a receive-pack (push), auto-initialize repository!
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if service == "git-receive-pack" {
			log.Printf("[INFO] Auto-provisioning new repository %q via Git Push...", repoName)
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			_, err := h.repoService.CreateRepo(ctx, repoName, "Auto-created on Git push", "main", false, false, "http://"+r.Host)
			cancel()
			if err != nil {
				log.Printf("[ERROR] Failed to auto-provision %q: %v", repoName, err)
				http.Error(w, "failed to auto-provision repository: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "repository not found", http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")

	// Pkt-line protocol header: # service=git-xxx\n
	packet := fmt.Sprintf("# service=%s\n", service)
	length := len(packet) + 4
	fmt.Fprintf(w, "%04x%s0000", length, packet)

	subcmd := strings.TrimPrefix(service, "git-") // "upload-pack" or "receive-pack"
	cmd := exec.CommandContext(r.Context(), "git", subcmd, "--stateless-rpc", "--advertise-refs", repoPath)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("git %s info/refs error: %v", subcmd, err)
	}
}

// UploadPack handles POST /git/{repo}/git-upload-pack (Clone, Fetch, Pull)
func (h *GitHTTPHandler) UploadPack(w http.ResponseWriter, r *http.Request) {
	h.serviceRPC(w, r, "upload-pack", false)
}

// ReceivePack handles POST /git/{repo}/git-receive-pack (Push)
func (h *GitHTTPHandler) ReceivePack(w http.ResponseWriter, r *http.Request) {
	h.serviceRPC(w, r, "receive-pack", true)
}

func (h *GitHTTPHandler) serviceRPC(w http.ResponseWriter, r *http.Request, subcmd string, isPush bool) {
	repoParam := chi.URLParam(r, "repo")
	repoName := strings.TrimSuffix(repoParam, ".git")

	repoPath := h.engine.RepoPath(repoName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if isPush {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			_, _ = h.repoService.CreateRepo(ctx, repoName, "Auto-created on Git push", "main", false, false, "http://"+r.Host)
			cancel()
		} else {
			http.Error(w, "repository not found", http.StatusNotFound)
			return
		}
	}

	var reqBody io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gz.Close()
		reqBody = gz
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", subcmd))
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")

	cmd := exec.CommandContext(r.Context(), "git", subcmd, "--stateless-rpc", repoPath)
	cmd.Stdin = reqBody
	cmd.Stdout = w
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		log.Printf("git RPC %s error on %s: %v (%s)", subcmd, repoName, err, errBuf.String())
		return
	}

	// If this was a successful push, automatically snapshot to Membuss in background
	if isPush {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			mid, err := h.repoService.SyncToMembuss(ctx, repoName)
			if err != nil {
				log.Printf("[WARN] Membuss auto-sync after push failed for %s: %v", repoName, err)
			} else {
				log.Printf("[INFO] Membuss auto-sync complete for %s -> root MID: %s", repoName, mid)
			}
		}()
	}
}
