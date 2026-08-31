# Sol - AI Agent & Developer Handover Guide

> **Project Name:** Sol  
> **Goal:** Ultra-lightweight, zero-latency, cross-platform Discord Rich Presence for Terminal & Shells (PowerShell, Bash, Zsh, Fish).  
> **Primary Tech Stack Target:** Rust (preferred for zero runtime overhead, sub-millisecond CLI startup, and <3MB RAM) or Go.

---

## 1. Project Philosophy & Core Directives

When modifying or expanding the **Sol** codebase, all contributors and AI agents must adhere to the following core constraints:

1. **Zero Shell Latency (< 5ms):**
   - The shell hooks (PowerShell / Bash) must **never** block the user's terminal prompt.
   - Shell hooks must only emit a fire-and-forget payload to the Sol local IPC socket. No heavy parsing or network calls should ever happen inside the shell hook process.
2. **Minimal Resource Footprint:**
   - **Daemon RAM:** Must remain strictly under **5MB** when idle and **10MB** under heavy load.
   - **Binary Size:** Final release binary should be optimized (e.g., via Rust LTO, stripping debug symbols, or Go `-ldflags="-s -w"`) to stay under **3–5MB**.
3. **Privacy First (Zero Leak Guarantee):**
   - Never send full un-sanitized command strings containing potential secrets (e.g., `--token`, `ghp_`, API keys, passwords, bearer tokens).
   - Sanitize or truncate absolute home paths (e.g., replace `C:\Users\Username\Projects\secret` with `~/Projects/secret` or project name `secret`).
4. **Living Documentation:**
   - Whenever an architectural change, protocol adjustment, or new feature is implemented, **both `AGENT_GUIDE.md` and `ARCHITECTURE.md` must be updated concurrently.**

---

## 2. Directory Layout & Module Structure

The intended repository structure for **Sol** is designed as a modular workspace:

```text
Sol/
├── AGENT_GUIDE.md          # Handover instructions, conventions, workflows (this file)
├── ARCHITECTURE.md         # In-depth technical knowledge base, protocol specs, dataflows
├── Cargo.toml              # (If Rust) Root workspace configuration
│
├── crates/ (or pkg/ if Go)
│   ├── sol-core/           # Protocol types, Discord IPC framer, Sanitizer, Git detector
│   ├── sol-daemon/         # Background service managing Discord RPC connection & rate limiting
│   └── sol-cli/            # Ultra-lightweight CLI to ping daemon or manage service
│
├── shells/                 # Shell integrations & hook scripts
│   ├── powershell/
│   │   ├── Sol.psm1        # PowerShell module with preexec/prompt hooks
│   │   └── install.ps1     # One-line installer for $PROFILE
│   └── bash/
│       ├── sol.bash        # Bash hook integration (preexec + PROMPT_COMMAND)
│       └── install.sh      # One-line installer for ~/.bashrc
│
└── assets/                 # Icons & Discord Application asset mapping tables
```

---

## 3. Step-by-Step Guide for Incoming AI Agents

When assigned a task on **Sol**, follow this standard protocol:

### Step 1: Read Knowledge Base First
Always read `ARCHITECTURE.md` to understand:
- The exact binary framing format for Discord IPC (`Opcode` + `Length` + `JSON`).
- The Sol Local IPC socket contract between the shell hooks and daemon.
- How rate limits and state transitions (`RUNNING` -> `IDLE` -> `AFK`) are managed.

### Step 2: Verify Latency Impact Before Implementing Shell Hooks
- Any code running in the shell's `preexec` or `prompt` event must execute in under **5 milliseconds**.
- If a background daemon is unreachable, the shell hook must fail silently and immediately (timeout <= 2ms).

### Step 3: Implement & Test Modularity
- Keep **Discord connection logic** isolated inside `sol-daemon`.
- Keep **Sanitization / Git parsing logic** inside `sol-core` with unit tests.
- Keep **Shell hooks** purely focused on capturing command names, working directory, and timestamps.

### Step 4: Update Documentation
- Update `ARCHITECTURE.md` if dataflow, IPC schemas, or config options change.
- Update the **Project Roadmap & Status** section below in this file.

---

## 4. Project Roadmap & Implementation Milestones

- [x] **Phase 1: Architecture & Technical Foundations** *(Completed)*
  - [x] Create `AGENT_GUIDE.md` and `ARCHITECTURE.md`.
  - [x] Finalize language selection (Go 1.27) and binary optimization flags (`-ldflags="-s -w"` -> 3.6MB binary).
- [x] **Phase 2: Core Daemon & Discord IPC Engine** *(Completed)*
  - [x] Implement cross-platform Discord IPC connector (Windows Named Pipes `\\.\pipe\discord-ipc-0` & Unix Domain Sockets).
  - [x] Implement packet framing (Opcode 0 Handshake, Opcode 1 Frame, Opcode 2 Close, Opcode 3/4 Ping/Pong).
  - [x] Implement Activity Dispatcher with 15-second rate limiter & debounce queue.
  - [x] Implement Sol Local IPC Server (Windows Named Pipe `\\.\pipe\sol-ipc` & Unix Socket `/tmp/sol.sock`).
- [x] **Phase 3: Shell Hook Modules (PowerShell & Bash)** *(Completed)*
  - [x] Develop `Sol.psm1` for PowerShell with in-process .NET NamedPipe stream (`< 0.5ms` overhead) and PSReadLine hooks.
  - [x] Develop `sol.bash` for Bash 4+ (with `DEBUG` trap and `PROMPT_COMMAND`).
  - [x] Automated one-line installers for PowerShell (`install.ps1`) and Bash (`install.sh`).
- [x] **Phase 4: Smart Detection & Privacy Filters** *(Completed)*
  - [x] Fast Git branch & repo detector (`.git/HEAD` reader without spawning child processes, `< 0.1ms`).
  - [x] Regex-based command sanitizer for secret stripping (`--password`, `--token`, `ghp_`, `sk_`, `Bearer`).
  - [x] Tool & Language asset mapper (Cargo, Go, Python, Docker, Neovim, Kubernetes, etc.).
- [x] **Phase 5: CLI, Packaging & Automated Installers** *(Completed)*
  - [x] CLI commands: `sol status`, `sol ping`, `sol notify`, `sol version`.
  - [x] Build script `build.ps1` producing lightweight stripped binaries in `bin/`.
  - [x] 100% automated test coverage across `discord`, `engine`, `git`, and `ipc` packages.

---

## 5. Coding & Contribution Rules

1. **No External Heavy Dependencies:** Prefer lightweight, native crates/packages. Avoid heavy web frameworks or bulky async runtimes if simple standard library threading or minimal async suffices.
2. **Error Handling:** Daemons must never crash on malformed IPC frames or Discord disconnections. Auto-reconnect with exponential backoff.
3. **Cross-Platform Path Handling:** Always use OS-agnostic path resolvers (handling both Windows `\` and Unix `/`).
