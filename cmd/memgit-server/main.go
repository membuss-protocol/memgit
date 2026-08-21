package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/membuss-protocol/memgit/pkg/api"
	"github.com/membuss-protocol/memgit/pkg/config"
	"github.com/membuss-protocol/memgit/pkg/gitengine"
	"github.com/membuss-protocol/memgit/pkg/identity"
	"github.com/membuss-protocol/memgit/pkg/membuss"
	"github.com/membuss-protocol/memgit/pkg/services"
	"github.com/membuss-protocol/memgit/pkg/swarm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	var (
		port       int
		portShort  int
		host       string
		dataDir    string
		webDir     string
		apiURL     string
		membussAPI string
		gwURL      string
		membussGW  string
	)

	flag.IntVar(&port, "port", 8500, "HTTP server port")
	flag.IntVar(&portShort, "p", 0, "HTTP server port (shorthand)")
	flag.StringVar(&host, "host", "0.0.0.0", "HTTP server host interface")
	flag.StringVar(&dataDir, "data-dir", cfg.DataDir, "Directory for local repository cache and metadata")
	flag.StringVar(&dataDir, "datadir", cfg.DataDir, "Directory for local repository cache and metadata")
	flag.StringVar(&webDir, "web-dir", cfg.WebDir, "Directory containing the built frontend SPA")
	flag.StringVar(&webDir, "webdir", cfg.WebDir, "Directory containing the built frontend SPA")

	flag.StringVar(&apiURL, "api-url", "", "Membuss Node API endpoint")
	flag.StringVar(&membussAPI, "membuss-api", "", "Membuss Node API endpoint (alias)")
	flag.StringVar(&membussAPI, "api", "", "Membuss Node API endpoint (alias)")

	flag.StringVar(&gwURL, "gw-url", "", "Membuss Gateway CDN endpoint")
	flag.StringVar(&membussGW, "membuss-gw", "", "Membuss Gateway CDN endpoint (alias)")
	flag.StringVar(&membussGW, "gw", "", "Membuss Gateway CDN endpoint (alias)")
	flag.Parse()

	// Resolve effective port
	effectivePort := port
	if portShort > 0 {
		effectivePort = portShort
	}

	// Resolve effective API URL
	effectiveAPI := cfg.MembussAPI
	if apiURL != "" {
		effectiveAPI = strings.TrimRight(apiURL, "/")
	} else if membussAPI != "" {
		effectiveAPI = strings.TrimRight(membussAPI, "/")
	}

	// Resolve effective Gateway URL
	effectiveGW := cfg.MembussGateway
	if gwURL != "" {
		effectiveGW = strings.TrimRight(gwURL, "/")
	} else if membussGW != "" {
		effectiveGW = strings.TrimRight(membussGW, "/")
	}

	log.Printf("=======================================================")
	log.Printf("  MEMGIT — Decentralized Git Platform on Membuss")
	log.Printf("=======================================================")
	log.Printf("Data directory:    %s", dataDir)
	log.Printf("Membuss Node API:  %s", effectiveAPI)
	log.Printf("Membuss Gateway:   %s", effectiveGW)

	serverBaseURL := fmt.Sprintf("http://localhost:%d", effectivePort)

	// Initialize Membuss client
	client := membuss.NewClient(effectiveAPI, effectiveGW)

	// Check Membuss connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	if nodeInfo, err := client.CheckHealth(ctx); err != nil {
		log.Printf("[WARN] Membuss daemon not detected: %v (running in offline local-cache mode)", err)
	} else {
		log.Printf("[INFO] Connected to Membuss daemon! Node ID: %s (v%s)", nodeInfo.ID, nodeInfo.Version)
	}
	cancel()

	// Initialize Git Engine
	reposDir := filepath.Join(dataDir, "repos")
	engine, err := gitengine.NewEngine(reposDir, client)
	if err != nil {
		log.Fatalf("Failed to initialize Git engine: %v", err)
	}

	// Initialize Services
	metaDir := filepath.Join(dataDir, "meta")
	repoService, err := services.NewRepoService(metaDir, engine, client)
	if err != nil {
		log.Fatalf("Failed to initialize Repo service: %v", err)
	}

	issuesDir := filepath.Join(dataDir, "issues")
	issueService, err := services.NewIssueService(issuesDir)
	if err != nil {
		log.Fatalf("Failed to initialize Issue service: %v", err)
	}

	pullsDir := filepath.Join(dataDir, "pulls")
	prService, err := services.NewPRService(pullsDir, engine)
	if err != nil {
		log.Fatalf("Failed to initialize PR service: %v", err)
	}

	releasesDir := filepath.Join(dataDir, "releases")
	releaseService, err := services.NewReleaseService(releasesDir)
	if err != nil {
		log.Fatalf("Failed to initialize Release service: %v", err)
	}

	// Initialize Identity & Swarm Catalog
	idManager, err := identity.NewManager(config.ConfigDir())
	if err != nil {
		log.Fatalf("Failed to initialize Identity manager: %v", err)
	}
	catalog := swarm.NewCatalog(repoService, idManager)
	currentUser := idManager.CurrentUser()
	log.Printf("[INFO] Active Developer Identity: @%s (Key: %s...)", currentUser.Username, currentUser.PublicKey[:12])

	// Initialize Handlers & Router
	gitHandler := api.NewGitHTTPHandler(engine, repoService)
	apiHandler := api.NewHandler(engine, repoService, issueService, prService, releaseService, client, idManager, catalog, serverBaseURL)
	router := api.SetupRouter(gitHandler, apiHandler, webDir)

	addr := fmt.Sprintf("%s:%d", host, effectivePort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown listener
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[INFO] MemGit server listening on http://%s (Git HTTP: %s/git/<repo>.git)", addr, serverBaseURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failure: %v", err)
		}
	}()

	<-stop
	log.Println("[INFO] Shutting down MemGit server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	}
	log.Println("[INFO] MemGit exited cleanly.")
}
