#!/usr/bin/env bash
# ==============================================================================
# Sol - Bash Hook Installer for ~/.bashrc
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOK_FILE="$SCRIPT_DIR/sol.bash"
BASHRC="$HOME/.bashrc"

if [ ! -f "$HOOK_FILE" ]; then
    echo "❌ Error: Could not find sol.bash at $HOOK_FILE"
    exit 1
fi

SOURCE_LINE="source \"$HOOK_FILE\""

if grep -Fxq "$SOURCE_LINE" "$BASHRC" 2>/dev/null; then
    echo "⚠️  Sol Bash hook is already configured in $BASHRC"
else
    echo "" >> "$BASHRC"
    echo "# Sol Discord Rich Presence" >> "$BASHRC"
    echo "$SOURCE_LINE" >> "$BASHRC"
    echo "✅ Successfully added Sol Bash hook to $BASHRC"
    echo "Run 'source ~/.bashrc' or restart your terminal to activate."
fi
