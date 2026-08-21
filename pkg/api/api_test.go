package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/membuss-protocol/memgit/pkg/gitengine"
	"github.com/membuss-protocol/memgit/pkg/identity"
	"github.com/membuss-protocol/memgit/pkg/membuss"
	"github.com/membuss-protocol/memgit/pkg/services"
	"github.com/membuss-protocol/memgit/pkg/swarm"
)

func TestAPI_EndToEnd(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "memgit-api-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Mock Membuss Node API
	mockMembuss := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/node/info" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok": true,
				"data": map[string]interface{}{
					"id":        "12D3KooWTestPeer12345",
					"addresses": []string{"/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWTestPeer12345"},
					"version":   "0.9.0",
				},
			})
			return
		}
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
					"size": 2048,
				},
			})
			return
		}
		if r.URL.Path == "/api/v1/peers" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ok": true,
				"data": []map[string]interface{}{
					{
						"id":      "12D3KooWAnchorNode99",
						"latency": "12ms",
					},
				},
			})
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

	repoService, err := services.NewRepoService(filepath.Join(tempDir, "meta"), engine, client)
	if err != nil {
		t.Fatal(err)
	}
	issueService, err := services.NewIssueService(filepath.Join(tempDir, "issues"))
	if err != nil {
		t.Fatal(err)
	}
	prService, err := services.NewPRService(filepath.Join(tempDir, "pulls"), engine)
	if err != nil {
		t.Fatal(err)
	}
	releaseService, err := services.NewReleaseService(filepath.Join(tempDir, "releases"))
	if err != nil {
		t.Fatal(err)
	}

	idManager, err := identity.NewManager(filepath.Join(tempDir, "config"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := swarm.NewCatalog(repoService, idManager)

	gitHandler := NewGitHTTPHandler(engine, repoService)
	apiHandler := NewHandler(engine, repoService, issueService, prService, releaseService, client, idManager, catalog, "http://localhost:8500")
	router := SetupRouter(gitHandler, apiHandler, "")

	server := httptest.NewServer(router)
	defer server.Close()

	// 1. Test System Status
	resp, err := http.Get(server.URL + "/api/v1/system/status")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("system status failed: %v, code %d", err, resp.StatusCode)
	}

	// 2. Test Create Repository
	createPayload, _ := json.Marshal(map[string]interface{}{
		"name":           "decentralized-docs",
		"description":    "Decentralized documentation repository on Membuss",
		"default_branch": "main",
		"init_readme":    true,
	})
	resp, err = http.Post(server.URL+"/api/v1/repos", "application/json", bytes.NewReader(createPayload))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("create repo failed: %v, code %d", err, resp.StatusCode)
	}

	// 3. Test List Repositories
	resp, err = http.Get(server.URL + "/api/v1/repos")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("list repos failed: %v", err)
	}
	var repoListEnv struct {
		OK   bool                     `json:"ok"`
		Data []map[string]interface{} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&repoListEnv)
	if len(repoListEnv.Data) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repoListEnv.Data))
	}

	// 4. Test Tree & Blob
	resp, err = http.Get(server.URL + "/api/v1/repos/decentralized-docs/tree/main")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("get tree failed: %v", err)
	}

	resp, err = http.Get(server.URL + "/api/v1/repos/decentralized-docs/blob/main/README.md")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("get blob failed: %v", err)
	}

	// 5. Test Issues API
	issuePayload, _ := json.Marshal(map[string]interface{}{
		"title":  "Support markdown tables",
		"body":   "Tables should align columns properly.",
		"author": "Alice",
		"labels": []string{"ui", "markdown"},
	})
	resp, err = http.Post(server.URL+"/api/v1/repos/decentralized-docs/issues", "application/json", bytes.NewReader(issuePayload))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("create issue failed: %v", err)
	}

	resp, err = http.Get(server.URL + "/api/v1/repos/decentralized-docs/issues")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("list issues failed: %v", err)
	}

	// 6. Test Swarm Telemetry
	resp, err = http.Get(server.URL + "/api/v1/repos/decentralized-docs/network")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("get network telemetry failed: %v", err)
	}
	var swarmEnv struct {
		OK   bool                   `json:"ok"`
		Data map[string]interface{} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&swarmEnv)
	if !swarmEnv.OK {
		t.Fatalf("swarm telemetry response not ok: %+v", swarmEnv)
	}
}
