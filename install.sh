#!/usr/bin/env bash
# ==============================================================================
# Sol One-Line Automated Installer for Linux / macOS (Bash & Zsh)
# ZERO dependencies required (No Go, No Git needed).
# ==============================================================================

set -e

echo ""
echo "  ☀️ Installing Sol - Terminal Discord Rich Presence..."
echo ""

SOL_HOME="$HOME/.sol"
SOL_BIN="$SOL_HOME/bin"

mkdir -p "$SOL_BIN"

REPO_RAW_URL="https://raw.githubusercontent.com/ThinhDost/sol/main"
RELEASE_URL="https://github.com/ThinhDost/sol/releases/latest/download"

# Download hook and config
curl -fsSL "$REPO_RAW_URL/shells/bash/sol.bash" -o "$SOL_HOME/sol.bash"
curl -fsSL "$REPO_RAW_URL/sol.config.json" -o "$SOL_HOME/sol.config.json"
chmod +x "$SOL_HOME/sol.bash"

# Detect OS and Architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

BINARY_NAME=""
CLI_NAME=""

case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64)
                BINARY_NAME="sol-linux-amd64"
                CLI_NAME="sol-linux-amd64-cli"
                ;;
            *)
                echo "  [!] Unsupported Linux architecture: $ARCH, will try compiling if Go is present."
                ;;
        esac
        ;;
    Darwin)
        case "$ARCH" in
            arm64)
                BINARY_NAME="sol-darwin-arm64"
                CLI_NAME="sol-darwin-arm64-cli"
                ;;
            x86_64)
                BINARY_NAME="sol-darwin-amd64"
                CLI_NAME="sol-darwin-amd64-cli"
                ;;
            *)
                echo "  [!] Unsupported macOS architecture: $ARCH, will try compiling if Go is present."
                ;;
        esac
        ;;
esac

# Download pre-built standalone binaries
if [ -n "$BINARY_NAME" ]; then
    echo "  [*] Downloading pre-built binary ($BINARY_NAME) from GitHub Release..."
    curl -fsSL "$RELEASE_URL/$BINARY_NAME" -o "$SOL_BIN/sol-daemon"
    curl -fsSL "$RELEASE_URL/$CLI_NAME" -o "$SOL_BIN/sol"
    chmod +x "$SOL_BIN/sol-daemon" "$SOL_BIN/sol"
elif command -v go >/dev/null 2>&1; then
    echo "  🔨 Compiling Sol binaries with Go..."
    TEMP_DIR=$(mktemp -d)
    git clone --depth 1 https://github.com/ThinhDost/sol.git "$TEMP_DIR" 2>/dev/null
    cd "$TEMP_DIR"
    go build -ldflags="-s -w" -o "$SOL_BIN/sol-daemon" ./cmd/sol-daemon
    go build -ldflags="-s -w" -o "$SOL_BIN/sol" ./cmd/sol-cli
    cd - >/dev/null
    rm -rf "$TEMP_DIR"
fi

# Add hook to ~/.bashrc or ~/.zshrc
HOOK_LINE="source \"$SOL_HOME/sol.bash\""

if [ -f "$HOME/.bashrc" ]; then
    if ! grep -Fxq "$HOOK_LINE" "$HOME/.bashrc" 2>/dev/null; then
        echo "" >> "$HOME/.bashrc"
        echo "# Sol Discord Rich Presence" >> "$HOME/.bashrc"
        echo "$HOOK_LINE" >> "$HOME/.bashrc"
        echo "  ✅ Added Sol hook to ~/.bashrc"
    fi
fi

if [ -f "$HOME/.zshrc" ]; then
    if ! grep -Fxq "$HOOK_LINE" "$HOME/.zshrc" 2>/dev/null; then
        echo "" >> "$HOME/.zshrc"
        echo "# Sol Discord Rich Presence" >> "$HOME/.zshrc"
        echo "$HOOK_LINE" >> "$HOME/.zshrc"
        echo "  ✅ Added Sol hook to ~/.zshrc"
    fi
fi

# Kill old daemon if running and restart in background
pkill -f sol-daemon >/dev/null 2>&1 || true

if [ -f "$SOL_BIN/sol-daemon" ]; then
    echo "  🚀 Launching Sol background daemon..."
    nohup "$SOL_BIN/sol-daemon" >/dev/null 2>&1 &
fi

echo ""
echo "  🎉 Sol is installed and running successfully!"
echo "  👉 Run 'source ~/.bashrc' or restart your terminal to activate."
echo ""
