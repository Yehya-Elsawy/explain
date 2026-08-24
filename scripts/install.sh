#!/usr/bin/env bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
echo "  ┌─────────────────────────────────────────────────────────────┐"
echo "  │    EXPLAIN CLI — Understand the command before you run it.  │"
echo "  └─────────────────────────────────────────────────────────────┘"
echo -e "${NC}"

# Target destination
INSTALL_DIR="/usr/local/bin"
if [ "$EUID" -ne 0 ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

# Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;;
    *) echo -e "${RED}[!] Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

REPO="Yehya-Elsawy/explain"
LATEST_URL="https://github.com/$REPO/releases/latest/download/explain_${OS}_${ARCH}.tar.gz"

echo -e "${YELLOW}[>] Downloading explain binary for ${OS}/${ARCH}...${NC}"

# Smooth progress bar animation
show_progress() {
    local width=30
    for ((i=1; i<=width; i++)); do
        local filled=$(printf '█%.0s' $(seq 1 $i))
        local empty=$(printf '░%.0s' $(seq 1 $((width - i))))
        local percent=$(( (i * 100) / width ))
        echo -ne "\r  ${CYAN}[${filled}${empty}]${NC} ${percent}%"
        sleep 0.02
    done
    echo -e "\r  ${GREEN}[██████████████████████████████] 100%${NC}"
}

TMP_DIR="$(mktemp -d)"
download_success=false

if curl -fSL --connect-timeout 10 "$LATEST_URL" -o "$TMP_DIR/explain.tar.gz" 2>/dev/null; then
    show_progress
    tar -xzf "$TMP_DIR/explain.tar.gz" -C "$TMP_DIR"
    if [ -f "$TMP_DIR/explain" ]; then
        if [ "$EUID" -eq 0 ]; then
            mv "$TMP_DIR/explain" "$INSTALL_DIR/explain"
        else
            if [ -w "$INSTALL_DIR" ]; then
                mv "$TMP_DIR/explain" "$INSTALL_DIR/explain"
            else
                sudo mv "$TMP_DIR/explain" "$INSTALL_DIR/explain"
            fi
        fi
        chmod +x "$INSTALL_DIR/explain"
        download_success=true
    fi
fi

rm -rf "$TMP_DIR"

# If release download didn't succeed, check if Go is installed to build from source
if [ "$download_success" = false ]; then
    echo -e "${YELLOW}[i] Pre-built release binary not found, compiling from source...${NC}"
    if command -v go >/dev/null 2>&1; then
        show_progress
        GOBIN="$INSTALL_DIR" go install "github.com/$REPO/cmd/explain@latest" || {
            TMP_SRC="$(mktemp -d)"
            git clone --depth 1 "https://github.com/$REPO.git" "$TMP_SRC"
            cd "$TMP_SRC"
            go build -ldflags="-s -w" -o "$INSTALL_DIR/explain" ./cmd/explain
            rm -rf "$TMP_SRC"
        }
        download_success=true
    else
        echo -e "${RED}[!] Could not download binary and Go is not installed on this machine.${NC}"
        echo -e "Please ensure a GitHub release is created on https://github.com/$REPO/releases"
        echo -e "Or install Go: https://golang.org/doc/install"
        exit 1
    fi
fi

echo -e "${GREEN}✓ Successfully installed explain to $INSTALL_DIR/explain${NC}"

# Ensure PATH includes install directory
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Note: $INSTALL_DIR is not in your current PATH.${NC}"
    echo "Add it to your shell config (~/.bashrc or ~/.zshrc):"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo
echo -e "${GREEN}🎉 Done! Try running:${NC}"
echo "   explain tar -xzf backup.tar.gz"
echo "   explain cd /home/"
echo "   explain \"ps aux | grep nginx | awk '{print \$2}' | xargs kill -9\""
echo "   explain \"curl -fsSL https://get.docker.com | sh\""
echo
