#!/usr/bin/env bash
# ==============================================================================
# Sol One-Line Automated Installer for Linux / macOS (Bash & Zsh)
# ==============================================================================

set -e

echo ""
echo "  ☀️ Installing Sol - Terminal Discord Rich Presence..."
echo ""

SOL_HOME="$HOME/.sol"
SOL_BIN="$SOL_HOME/bin"

mkdir -p "$SOL_BIN"

REPO_RAW_URL="https://raw.githubusercontent.com/ThinhDost/sol/main"

# Download module and config
curl -fsSL "$REPO_RAW_URL/shells/bash/sol.bash" -o "$SOL_HOME/sol.bash"
curl -fsSL "$REPO_RAW_URL/sol.config.json" -o "$SOL_HOME/sol.config.json"

chmod +x "$SOL_HOME/sol.bash"

# Compile or install binaries if Go is present
if command -v go >/dev/null 2>&1; then
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
    chmod +x "$SOL_BIN/sol-daemon" "$SOL_BIN/sol" 2>/dev/null || true
    echo "  🚀 Launching Sol background daemon..."
    nohup "$SOL_BIN/sol-daemon" >/dev/null 2>&1 &
fi

echo ""
echo "  🎉 Sol is installed and running successfully!"
echo "  👉 Run 'source ~/.bashrc' or restart your terminal to activate."
echo ""
