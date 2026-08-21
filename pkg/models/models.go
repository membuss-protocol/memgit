package models

import (
	"encoding/hex"
	"time"
)

// User represents a developer identity on the Membuss network.
type User struct {
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	PublicKey   string    `json:"public_key"` // Ed25519 hex string
	JoinedAt    time.Time `json:"joined_at"`
	StarCount   int       `json:"star_count"`
	RepoCount   int       `json:"repo_count"`
}

// PublicKeyBytes returns the raw byte slice of the public key.
func (u *User) PublicKeyBytes() []byte {
	b, _ := hex.DecodeString(u.PublicKey)
	return b
}

// Repository represents a decentralized Git repository on Membuss.
type Repository struct {
	Name          string    `json:"name"`           // Short name (e.g. "quantum-engine")
	Owner         string    `json:"owner"`          // Username of author/owner (e.g. "alice")
	FullName      string    `json:"full_name"`      // Namespaced "owner/name" or "name"
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	MemNSKey      string    `json:"memns_key"`      // Local KeyRing key name (e.g. "memgit-my-repo")
	MemNSName     string    `json:"memns_name"`     // Cryptographic MemNS name (e.g. "memns1z...")
	LatestMID     string    `json:"latest_mid"`     // Immutable Merkle DAG MID of latest commit tree
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	StarCount     int       `json:"star_count"`
	ForkCount     int       `json:"fork_count"`
	IsPrivate     bool      `json:"is_private"`
	IsLocal       bool      `json:"is_local"`       // True if hosted locally; false if discovered from peer swarm
	OriginNode    string    `json:"origin_node"`    // Origin peer ID
	Topics        []string  `json:"topics"`
	ForkedFrom    string    `json:"forked_from,omitempty"`
	StarredBy     []string  `json:"starred_by,omitempty"` // User public keys
	CloneHTTPS    string    `json:"clone_https"`
	CloneGateway  string    `json:"clone_gateway"`
	CloneMembuss  string    `json:"clone_membuss"`
}

// Branch represents a Git branch reference.
type Branch struct {
	Name      string    `json:"name"`
	CommitSHA string    `json:"commit_sha"`
	IsDefault bool      `json:"is_default"`
	Message   string    `json:"message"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tag represents a Git tag reference.
type Tag struct {
	Name      string    `json:"name"`
	CommitSHA string    `json:"commit_sha"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// Commit represents a single Git commit with metadata and statistics.
type Commit struct {
	SHA       string    `json:"sha"`
	ShortSHA  string    `json:"short_sha"`
	TreeSHA   string    `json:"tree_sha"`
	Parents   []string  `json:"parents"`
	Author    Signature `json:"author"`
	Committer Signature `json:"committer"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	MID       string    `json:"mid,omitempty"` // Corresponding Membuss MID for this tree
	Stats     struct {
		FilesChanged int `json:"files_changed"`
		Additions    int `json:"additions"`
		Deletions    int `json:"deletions"`
	} `json:"stats"`
}

// Signature represents an author/committer signature.
type Signature struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	When  time.Time `json:"when"`
}

// FileNode represents a file or folder in a tree listing.
type FileNode struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Type         string   `json:"type"` // "blob" or "tree"
	Size         int64    `json:"size"`
	Mode         string   `json:"mode"`
	BlobSHA      string   `json:"blob_sha"`
	LatestCommit *Commit  `json:"latest_commit,omitempty"`
}

// BlobContent represents the text/binary content of a file.
type BlobContent struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	IsBinary bool   `json:"is_binary"`
	Content  string `json:"content"`
	MimeType string `json:"mime_type"`
	BlobSHA  string `json:"blob_sha"`
}

// CommitDiff represents changes in a single commit.
type CommitDiff struct {
	Commit Commit     `json:"commit"`
	Files  []FileDiff `json:"files"`
}

// FileDiff represents changes to an individual file.
type FileDiff struct {
	OldPath   string `json:"old_path"`
	NewPath   string `json:"new_path"`
	Status    string `json:"status"` // "added", "modified", "deleted", "renamed"
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch"`
}

// Issue represents a decentralized repository issue.
type Issue struct {
	ID        int            `json:"id"`
	RepoName  string         `json:"repo_name"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Author    string         `json:"author"`
	State     string         `json:"state"` // "open", "closed"
	Labels    []string       `json:"labels"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Comments  []IssueComment `json:"comments"`
}

// IssueComment represents a comment on an issue.
type IssueComment struct {
	ID        int       `json:"id"`
	IssueID   int       `json:"issue_id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// PullRequest represents a decentralized pull request / proposal.
type PullRequest struct {
	ID           int            `json:"id"`
	RepoName     string         `json:"repo_name"`
	Title        string         `json:"title"`
	Body         string         `json:"body"`
	Author       string         `json:"author"`
	State        string         `json:"state"` // "open", "closed", "merged"
	SourceBranch string         `json:"source_branch"`
	TargetBranch string         `json:"target_branch"`
	SourceRepo   string         `json:"source_repo"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	MergedAt     *time.Time     `json:"merged_at,omitempty"`
	Diffs        []FileDiff     `json:"diffs,omitempty"`
	Comments     []IssueComment `json:"comments"`
}

// Release represents a tagged version release with downloadable assets and MID.
type Release struct {
	ID          int       `json:"id"`
	RepoName    string    `json:"repo_name"`
	TagName     string    `json:"tag_name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CommitSHA   string    `json:"commit_sha"`
	MID         string    `json:"mid"` // Direct snapshot MID on Membuss
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a downloadable file attached to a release.
type Asset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url"`
	MID         string `json:"mid"`
}

// SwarmStatus represents P2P health and replication metrics from Membuss.
type SwarmStatus struct {
	NodeID          string            `json:"node_id"`
	Addresses       []string          `json:"addresses"`
	ConnectedPeers  int               `json:"connected_peers"`
	DHTProviders    int               `json:"dht_providers"`
	AnchorSynced    bool              `json:"anchor_synced"`
	ErasureParity   string            `json:"erasure_parity"` // "10+4 Reed-Solomon"
	MemNSVersions   []MemNSLogEntry   `json:"memns_versions"`
	ReplicationList []PeerReplication `json:"replication_list"`
}

// MemNSLogEntry represents one historical signed record published to a MemNS key.
type MemNSLogEntry struct {
	Sequence  uint64    `json:"sequence"`
	TargetMID string    `json:"target_mid"`
	Timestamp time.Time `json:"timestamp"`
	Signature string    `json:"signature"`
}

// PeerReplication represents peer status for a repository MID.
type PeerReplication struct {
	PeerID    string `json:"peer_id"`
	Latency   string `json:"latency"`
	IsAnchor  bool   `json:"is_anchor"`
	HasShards bool   `json:"has_shards"`
}

// ActivityEvent represents an action in the global decentralized activity feed.
type ActivityEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "commit", "create_repo", "star", "fork", "issue", "pr", "release"
	Actor     string    `json:"actor"`
	RepoName  string    `json:"repo_name"`
	Summary   string    `json:"summary"`
	Detail    string    `json:"detail,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
