package config

import (
	"os"
	"testing"
)

func TestConfig_DefaultAndSet(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ServerURL != "http://localhost:8500" {
		t.Errorf("unexpected default server URL: %s", cfg.ServerURL)
	}
	if cfg.MembussAPI != "http://127.0.0.1:5004" {
		t.Errorf("unexpected default API URL: %s", cfg.MembussAPI)
	}
	if cfg.MembussGateway != "https://gateway.membuss.dpdns.org" {
		t.Errorf("unexpected default Gateway URL: %s", cfg.MembussGateway)
	}

	tempHome, err := os.MkdirTemp("", "memgit-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	t.Setenv("USERPROFILE", tempHome)
	t.Setenv("HOME", tempHome)

	err = cfg.SetKey("membuss_api", "http://192.168.1.100:5001")
	if err != nil {
		t.Fatalf("SetKey failed: %v", err)
	}
	if cfg.MembussAPI != "http://192.168.1.100:5001" {
		t.Errorf("expected updated API, got %s", cfg.MembussAPI)
	}

	err = cfg.SetKey("membuss_gateway", "http://192.168.1.100:8080")
	if err != nil {
		t.Fatalf("SetKey failed: %v", err)
	}
	if cfg.MembussGateway != "http://192.168.1.100:8080" {
		t.Errorf("expected updated Gateway, got %s", cfg.MembussGateway)
	}

	// Verify reload from disk
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.MembussAPI != "http://192.168.1.100:5001" {
		t.Errorf("expected persisted API, got %s", loaded.MembussAPI)
	}
	if loaded.MembussGateway != "http://192.168.1.100:8080" {
		t.Errorf("expected persisted Gateway, got %s", loaded.MembussGateway)
	}
}
