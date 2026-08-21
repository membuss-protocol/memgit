package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/membuss-protocol/memgit/pkg/models"
)

// Manager manages local developer identities, cryptographic keys, and profiles.
type Manager struct {
	mu           sync.RWMutex
	identityPath string
	currentUser  *models.User
	privateKey   ed25519.PrivateKey
	users        map[string]*models.User // cached profiles by username
}

// NewManager creates or loads the local developer identity.
func NewManager(configDir string) (*Manager, error) {
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, err
	}

	identityPath := filepath.Join(configDir, "identity.json")
	mgr := &Manager{
		identityPath: identityPath,
		users:        make(map[string]*models.User),
	}

	if err := mgr.loadOrCreate(); err != nil {
		return nil, err
	}

	return mgr, nil
}

type identityFile struct {
	User       models.User `json:"user"`
	PrivateKey string      `json:"private_key"`
	PublicKey  string      `json:"public_key"`
}

func (m *Manager) loadOrCreate() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if data, err := os.ReadFile(m.identityPath); err == nil {
		var idFile identityFile
		if err := json.Unmarshal(data, &idFile); err == nil && len(idFile.PrivateKey) > 0 {
			privBytes, err := hex.DecodeString(idFile.PrivateKey)
			if err == nil && len(privBytes) == ed25519.PrivateKeySize {
				m.privateKey = ed25519.PrivateKey(privBytes)
				user := idFile.User
				if user.Username == "" {
					user.Username = "developer"
				}
				if user.DisplayName == "" {
					user.DisplayName = "Membuss Developer"
				}
				m.currentUser = &user
				m.users[user.Username] = &user
				return nil
			}
		}
	}

	// Generate new Ed25519 keypair for default user
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate identity keypair: %w", err)
	}

	pubHex := hex.EncodeToString(pub)
	shortKey := pubHex
	if len(shortKey) > 8 {
		shortKey = shortKey[:8]
	}

	user := models.User{
		Username:    "dev-" + shortKey,
		DisplayName: "Membuss Contributor",
		Bio:         "Building decentralized software on Membuss Network.",
		Email:       "dev@" + shortKey + ".membuss.network",
		PublicKey:   pubHex,
		JoinedAt:    time.Now(),
		StarCount:   0,
		RepoCount:   0,
	}

	m.privateKey = priv
	m.currentUser = &user
	m.users[user.Username] = &user

	return m.saveCurrent()
}

func (m *Manager) saveCurrent() error {
	idFile := identityFile{
		User:       *m.currentUser,
		PrivateKey: hex.EncodeToString(m.privateKey),
		PublicKey:  hex.EncodeToString(m.currentUser.PublicKeyBytes()),
	}

	data, err := json.MarshalIndent(idFile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.identityPath, data, 0o600)
}

// CurrentUser returns the currently active user profile.
func (m *Manager) CurrentUser() *models.User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copy := *m.currentUser
	return &copy
}

// UpdateProfile updates profile fields for the current user.
func (m *Manager) UpdateProfile(username, displayName, bio, email string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if username != "" {
		delete(m.users, m.currentUser.Username)
		m.currentUser.Username = username
	}
	if displayName != "" {
		m.currentUser.DisplayName = displayName
	}
	if bio != "" {
		m.currentUser.Bio = bio
	}
	if email != "" {
		m.currentUser.Email = email
	}

	m.users[m.currentUser.Username] = m.currentUser
	if err := m.saveCurrent(); err != nil {
		return nil, err
	}

	copy := *m.currentUser
	return &copy, nil
}

// GetUser returns a user profile by username or public key.
func (m *Manager) GetUser(usernameOrKey string) (*models.User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if u, exists := m.users[usernameOrKey]; exists {
		copy := *u
		return &copy, true
	}

	for _, u := range m.users {
		if u.PublicKey == usernameOrKey {
			copy := *u
			return &copy, true
		}
	}

	return nil, false
}

// RegisterOrUpdateUser caches a remote user profile discovered from the swarm.
func (m *Manager) RegisterOrUpdateUser(user *models.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user != nil && user.Username != "" {
		m.users[user.Username] = user
	}
}

// ListUsers returns all cached users.
func (m *Manager) ListUsers() []*models.User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*models.User, 0, len(m.users))
	for _, u := range m.users {
		copy := *u
		res = append(res, &copy)
	}
	return res
}
