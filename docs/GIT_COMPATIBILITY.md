# 🔌 Native Git CLI Compatibility Guide

MemGit was built from the ground up to provide **100% protocol compatibility** with the official `git` command-line interface, Git GUI clients (GitKraken, SourceTree, VS Code), and CI/CD pipelines.

---

## 📑 Contents
1. [Git Smart HTTP Transport](#1-git-smart-http-transport)
2. [Push-to-Create Autoprovisioning](#2-push-to-create-autoprovisioning)
3. [Day-to-Day Git Workflow](#3-day-to-day-git-workflow)
4. [Native Git Subcommand (`git memgit`)](#4-native-git-subcommand-git-memgit)
5. [Native P2P Remote Transport Helper (`membuss://`)](#5-native-p2p-remote-transport-helper-membuss)

---

## 1. Git Smart HTTP Transport

MemGit implements the standard **Git Smart HTTP Protocol v1 & v2** over HTTP/HTTPS:

- **Ref Discovery**: `GET /git/<repo>.git/info/refs?service=git-upload-pack` (or `git-receive-pack`)
- **Pack Negotiation & Fetch**: `POST /git/<repo>.git/git-upload-pack`
- **Pack Ingestion & Push**: `POST /git/<repo>.git/git-receive-pack`

### Protocol Characteristics:
- **Packet-Line Framing**: All ref advertisements and status reports use standard 4-byte hexadecimal length-prefixed packet lines (e.g. `001e# service=git-upload-pack\n0000`).
- **Gzip Compression**: Automatically decompresses and negotiates gzipped packfiles sent by modern Git clients.
- **Stateless RPC**: Runs fully stateless RPC invocations against bare Git repositories, ensuring maximum concurrent throughput.

---

## 2. Push-to-Create Autoprovisioning

Unlike traditional Git servers that require you to open a web browser and manually click "New Repository" before pushing, MemGit supports **Push-to-Create**:

```bash
cd my-new-project
git init
git add .
git commit -m "Initial commit"
git remote add origin http://localhost:8500/git/my-new-project.git
git push -u origin main
```

### What happens behind the scenes:
1. When MemGit receives the initial `git-receive-pack` request for a repository that does not exist yet:
   - A new bare Git repository is created in `./data/repos/my-new-project.git`.
   - A dedicated **Ed25519 MemNS Key** (`memgit-my-new-project-<timestamp>`) is generated via the Membuss Node API.
   - Metadata is indexed in `./data/meta/`.
2. The incoming packfile is ingested by `git receive-pack`.
3. An asynchronous worker packages the repository into a Merkle DAG, stores it in BadgerDB with **Reed-Solomon (10+4) erasure coding**, and signs the new root MID to your MemNS key.

---

## 3. Day-to-Day Git Workflow

### Cloning
```bash
git clone http://localhost:8500/git/<repo-name>.git
cd <repo-name>
```

### Creating Branches & Pushing
```bash
git checkout -b feature/p2p-streaming
# ... make code changes ...
git commit -am "Implement chunked P2P block streaming"
git push origin feature/p2p-streaming
```

### Pulling & Fetching
```bash
git pull origin main
git fetch --all --prune
```

### Tagging & Releases
```bash
git tag -a v1.0.0 -m "Production Release 1.0"
git push origin v1.0.0
```
*Creating and pushing a Git tag triggers a permanent Merkle DAG snapshot on Membuss, making the release downloadable by root MID.*

---

## 4. Native Git Subcommand (`git memgit`)

Git allows any binary named `git-<subcommand>` in your `PATH` to be executed as a native Git subcommand:

1. Add `D:\projects\go\rsrc\membuss\memgit\bin` to your system `PATH`.
2. Run any MemGit command directly with `git`:

```bash
# List all hosted decentralized repositories
git memgit list

# Inspect repository status & swarm telemetry
git memgit info demo-p2p-app

# Force snapshot to Membuss
git memgit sync demo-p2p-app

# Create an issue
git memgit issue create demo-p2p-app -t "Add WebSocket live sync" -a "Alice"
```

---

## 5. Native P2P Remote Transport Helper (`membuss://`)

MemGit provides a custom transport helper binary: **`git-remote-membuss.exe`**.

When installed in your `PATH`, Git recognizes the `membuss://` scheme and invokes the helper to stream objects directly over the Membuss P2P overlay:

```bash
# Clone using a MemNS name
git clone membuss://memns1zdemo-p2p-app

# Clone using an immutable Merkle DAG root MID
git clone membuss://mem1z4a2b3c4d5e6f7...
```
