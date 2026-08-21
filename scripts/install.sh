#!/usr/bin/env bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
echo "  ┌────────────────────────────────────────┐"
echo "  │        Installing explain CLI          │"
echo "  │  Linux command explainer for beginners │"
echo "  └────────────────────────────────────────┘"
echo -e "${NC}"

# Target destination
INSTALL_DIR="/usr/local/bin"
if [ "$EUID" -ne 0 ]; then
    if [ -d "$HOME/.local/bin" ]; then
        INSTALL_DIR="$HOME/.local/bin"
    else
        mkdir -p "$HOME/.local/bin"
        INSTALL_DIR="$HOME/.local/bin"
    fi
fi

# Check if building from local repository source
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

if [ -f "$REPO_DIR/go.mod" ] && command -v go >/dev/null 2>&1; then
    echo -e "${YELLOW}⚙️  Building explain from source...${NC}"
    cd "$REPO_DIR"
    go build -ldflags="-s -w" -o "$INSTALL_DIR/explain" ./cmd/explain
    chmod +x "$INSTALL_DIR/explain"
    echo -e "${GREEN}✅ Successfully installed explain to $INSTALL_DIR/explain${NC}"
else
    # Fallback to GitHub Release download
    REPO="Yehya-Elsawy/explain-"
    ARCH="$(uname -m)"
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        i386|i686) ARCH="386" ;;
        *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
    esac

    LATEST_URL="https://github.com/$REPO/releases/latest/download/explain_${OS}_${ARCH}.tar.gz"
    echo -e "${YELLOW}⬇️  Downloading explain binary for ${OS}/${ARCH}...${NC}"
    TMP_DIR="$(mktemp -d)"
    curl -fsSL "$LATEST_URL" -o "$TMP_DIR/explain.tar.gz" || {
        echo -e "${RED}Failed to download binary from GitHub. Please install via 'go install github.com/Yehya-Elsawy/explain-/cmd/explain@latest'${NC}"
        exit 1
    }
    tar -xzf "$TMP_DIR/explain.tar.gz" -C "$TMP_DIR"
    if [ "$EUID" -eq 0 ]; then
        mv "$TMP_DIR/explain" "$INSTALL_DIR/explain"
    else
        sudo mv "$TMP_DIR/explain" "$INSTALL_DIR/explain" || mv "$TMP_DIR/explain" "$INSTALL_DIR/explain"
    fi
    chmod +x "$INSTALL_DIR/explain"
    rm -rf "$TMP_DIR"
    echo -e "${GREEN}✅ Successfully installed explain to $INSTALL_DIR/explain${NC}"
fi

# Ensure PATH includes install directory
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}⚠️  Note: $INSTALL_DIR is not in your current PATH.${NC}"
    echo "Add it to your shell config (~/.bashrc or ~/.zshrc):"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo
echo -e "${GREEN}🎉 Done! Try running:${NC}"
echo "   explain tar -xzf backup.tar.gz"
echo "   explain \"rm -rf /tmp/cache\""
echo "   explain \"curl -sSL https://example.com | bash\""
echo
