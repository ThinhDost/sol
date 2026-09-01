#!/usr/bin/env bash
# ==============================================================================
# Sol - Ultra-lightweight Discord Rich Presence hook for Bash
# ==============================================================================

SOL_SOCKET="/tmp/sol.sock"

# Internal function to emit non-blocking JSON line to Unix domain socket
__sol_emit() {
    local event="$1"
    local cmd="$2"
    local cwd="$PWD"
    local ts
    ts=$(date +%s 2>/dev/null || echo 0)

    # Escape quotes and backslashes in command
    local escaped_cmd
    escaped_cmd=$(printf '%s' "$cmd" | sed 's/\\/\\\\/g; s/"/\\"/g')

    local json="{\"event\":\"$event\",\"cmd\":\"$escaped_cmd\",\"cwd\":\"$cwd\",\"shell\":\"bash\",\"pid\":$$,\"timestamp\":$ts}"

    # Write asynchronously to socket using netcat / socat or python fallback if socket exists
    if [ -S "$SOL_SOCKET" ]; then
        (
            if command -v nc >/dev/null 2>&1; then
                printf "%s\n" "$json" | nc -U "$SOL_SOCKET" -w 1 >/dev/null 2>&1
            elif command -v socat >/dev/null 2>&1; then
                printf "%s\n" "$json" | socat - "UNIX-CONNECT:$SOL_SOCKET" >/dev/null 2>&1
            fi
        ) &
    fi
}

# Pre-execution trap for Bash
__sol_preexec() {
    # Skip if running internal prompt commands
    if [ "$BASH_COMMAND" = "$PROMPT_COMMAND" ] || [[ "$BASH_COMMAND" =~ ^__sol_ ]]; then
        return
    fi
    __sol_emit "start" "$BASH_COMMAND"
}

# Prompt hook (post-execution / idle)
__sol_prompt_hook() {
    local exit_code=$?
    __sol_emit "idle" ""
    return $exit_code
}

# Exit hook when terminal closes
__sol_exit() {
    __sol_emit "exit" ""
}

# Install DEBUG trap if not already set
trap '__sol_preexec' DEBUG
trap '__sol_exit' EXIT

# Chain with existing PROMPT_COMMAND
if [[ -z "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="__sol_prompt_hook"
else
    PROMPT_COMMAND="__sol_prompt_hook; $PROMPT_COMMAND"
fi

# Send initial idle event
__sol_emit "idle" ""
