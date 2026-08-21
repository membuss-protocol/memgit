#!/usr/bin/env bash
# ==============================================================================
# MemGit Build & Package Script (Linux / macOS / WSL)
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/bin"

echo "======================================================="
echo "  MemGit — Build System (Svelte 5 + Go Backend)"
echo "======================================================="

clean() {
    echo "🧹 Cleaning previous build artifacts..."
    rm -rf "${OUTPUT_DIR}" "${SCRIPT_DIR}/web/dist"
    echo "✔ Clean complete."
}

build_web() {
    echo "📦 [1/3] Building Svelte 5 Frontend..."
    cd "${SCRIPT_DIR}/web"
    if [ ! -d "node_modules" ]; then
        echo "  Installing npm dependencies..."
        npm install --silent
    fi
    npm run build
    echo "✔ Svelte frontend compiled successfully to web/dist/"
    cd "${SCRIPT_DIR}"
}

build_bin() {
    echo "🔨 [2/3] Compiling MemGit Server Binary..."
    mkdir -p "${OUTPUT_DIR}"

    echo "  -> Compiling memgit-server..."
    go build -ldflags="-s -w" -o "${OUTPUT_DIR}/memgit-server" "${SCRIPT_DIR}/cmd/memgit-server"

    echo "✔ Server binary compiled to bin/memgit-server"
}

run_tests() {
    echo "🧪 [3/3] Running Unit & Integration Tests..."
    cd "${SCRIPT_DIR}"
    go test -v ./...
    echo "✔ All tests passed 100%!"
}

case "${1:-all}" in
    clean)
        clean
        ;;
    web)
        build_web
        ;;
    bin)
        build_bin
        ;;
    test)
        run_tests
        ;;
    all)
        build_web
        build_bin
        run_tests
        echo ""
        echo "🚀 MemGit built successfully! Launch with: ./bin/memgit-server -port 8500"
        ;;
    *)
        echo "Usage: $0 {all|web|bin|test|clean}"
        exit 1
        ;;
esac
