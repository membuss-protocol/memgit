package membuss

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a type-safe HTTP client for the local Membuss Node API and Mem-Gate.
type Client struct {
	apiBase     string
	gatewayBase string
	httpClient  *http.Client
}

// NewClient returns a new Membuss API client.
func NewClient(apiBase, gatewayBase string) *Client {
	if apiBase == "" {
		apiBase = "http://127.0.0.1:5001"
	}
	if gatewayBase == "" {
		gatewayBase = "http://127.0.0.1:8080"
	}
	return &Client{
		apiBase:     strings.TrimRight(apiBase, "/"),
		gatewayBase: strings.TrimRight(gatewayBase, "/"),
		httpClient:  &http.Client{Timeout: 2 * time.Second},
	}
}

// GatewayBase returns the configured gateway URL.
func (c *Client) GatewayBase() string {
	return c.gatewayBase
}

// APIBase returns the configured node API URL.
func (c *Client) APIBase() string {
	return c.apiBase
}

// APIResponse is the standard JSON envelope returned by Membuss Node API.
type APIResponse struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

// NodeInfo represents the local Membuss node identity.
type NodeInfo struct {
	ID        string   `json:"peer_id"`
	AltID     string   `json:"id,omitempty"`
	Addresses []string `json:"addrs"`
	AltAddrs  []string `json:"addresses,omitempty"`
	Version   string   `json:"version"`
}

// PeerInfo represents a connected swarm peer.
type PeerInfo struct {
	ID        string   `json:"peer_id"`
	AltID     string   `json:"id,omitempty"`
	Addresses []string `json:"addresses"`
	Latency   string   `json:"latency"`
	Protocols []string `json:"protocols"`
}

// KeyInfo represents a MemNS signing key.
type KeyInfo struct {
	Name      string    `json:"name"`
	MemNSName string    `json:"memns_name"`
	CreatedAt time.Time `json:"created_at"`
}

// DirectoryPart represents a file inside a multipart directory upload.
type DirectoryPart struct {
	Path    string
	Content []byte
}

// CheckHealth verifies connection to the Membuss daemon.
func (c *Client) CheckHealth(ctx context.Context) (*NodeInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/api/v1/node/info", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("membuss daemon unreachable at %s: %w", c.apiBase, err)
	}
	defer resp.Body.Close()

	var env struct {
		OK   bool     `json:"ok"`
		Data NodeInfo `json:"data"`
		Err  string   `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if !env.OK {
		return nil, errors.New(env.Err)
	}
	if env.Data.ID == "" && env.Data.AltID != "" {
		env.Data.ID = env.Data.AltID
	}
	if len(env.Data.Addresses) == 0 && len(env.Data.AltAddrs) > 0 {
		env.Data.Addresses = env.Data.AltAddrs
	}
	return &env.Data, nil
}

// GenerateMemNSKey generates a new cryptographic Ed25519 key for a repository.
func (c *Client) GenerateMemNSKey(ctx context.Context, keyName string) (*KeyInfo, error) {
	payload, _ := json.Marshal(map[string]string{"name": keyName})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/api/v1/memns/key/generate", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var env struct {
		OK   bool    `json:"ok"`
		Data KeyInfo `json:"data"`
		Err  string  `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if !env.OK {
		return nil, errors.New(env.Err)
	}
	return &env.Data, nil
}

// ListMemNSKeys lists all registered signing keys.
func (c *Client) ListMemNSKeys(ctx context.Context) ([]KeyInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/api/v1/memns/key/list", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var env struct {
		OK   bool      `json:"ok"`
		Data []KeyInfo `json:"data"`
		Err  string    `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if !env.OK {
		return nil, errors.New(env.Err)
	}
	return env.Data, nil
}

// PublishMemNS signs and broadcasts an update for a mutable repository pointer.
func (c *Client) PublishMemNS(ctx context.Context, keyName, targetMID string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"key":    keyName,
		"target": "/mem/" + strings.TrimPrefix(targetMID, "/mem/"),
		"ttl":    ttl.String(),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/api/v1/memns/publish", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var env APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return err
	}
	if !env.OK {
		return errors.New(env.Error)
	}
	return nil
}

// ResolveMemNS resolves a mutable MemNS name or domain to an immutable MID.
func (c *Client) ResolveMemNS(ctx context.Context, name string) (string, error) {
	name = strings.TrimPrefix(name, "/memns/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/api/v1/memns/resolve/"+url.PathEscape(name), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			TargetMID string `json:"target_mid"`
		} `json:"data"`
		Err string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return "", err
	}
	if !env.OK {
		return "", errors.New(env.Err)
	}
	return strings.TrimPrefix(env.Data.TargetMID, "/mem/"), nil
}

// IngestDirectory ingests a multi-file tree into MemFS, returning the resulting root MID.
func (c *Client) IngestDirectory(ctx context.Context, repoName string, files []DirectoryPart) (string, uint64, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, f := range files {
		partHeader := make(map[string][]string)
		partHeader["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="file"; filename="%s"`, f.Path)}
		partWriter, err := writer.CreatePart(partHeader)
		if err != nil {
			return "", 0, err
		}
		if _, err := partWriter.Write(f.Content); err != nil {
			return "", 0, err
		}
	}
	if err := writer.Close(); err != nil {
		return "", 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/api/v1/add/dir?name="+url.QueryEscape(repoName), body)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			MID  string `json:"mid"`
			Size uint64 `json:"size"`
		} `json:"data"`
		Err string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return "", 0, err
	}
	if !env.OK {
		return "", 0, errors.New(env.Err)
	}

	// Automatically Seal the root MID to preserve it permanently
	_ = c.Seal(ctx, env.Data.MID, true)

	return env.Data.MID, env.Data.Size, nil
}

// Seal pins an MID recursively in BadgerDB.
func (c *Client) Seal(ctx context.Context, mid string, recursive bool) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"mid":       mid,
		"recursive": recursive,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiBase+"/api/v1/seal", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetCat returns raw block bytes or file content by MID.
func (c *Client) GetCat(ctx context.Context, mid string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/api/v1/cat/"+mid, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch MID %s: status %d", mid, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// GetPeers returns connected swarm peers.
func (c *Client) GetPeers(ctx context.Context) ([]PeerInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+"/api/v1/peers", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var env struct {
		OK   bool       `json:"ok"`
		Data []PeerInfo `json:"data"`
		Err  string     `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	return env.Data, nil
}
