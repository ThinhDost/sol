# ☀️ Sol - Terminal Discord Rich Presence

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=for-the-badge&logo=go)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-blue?style=for-the-badge&logo=windows)
![Shells](https://img.shields.io/badge/Shells-PowerShell%20%7C%20Bash-4EAA25?style=for-the-badge&logo=powershell)
![Discord](https://img.shields.io/badge/Discord-Rich%20Presence-5865F2?style=for-the-badge&logo=discord&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-orange?style=for-the-badge)

**An ultra-lightweight, zero-latency, cross-platform Discord Rich Presence daemon & shell hook system.**

[Features](#-key-features) • [Quick Start](#-quick-start) • [Configuration](#-configuration) • [Testing](#-testing--verification) • [Roadmap](#-roadmap)

</div>

---

## ⚡ Key Features

- **🚀 Zero Shell Latency (< 0.5ms):** PowerShell hooks use direct in-memory `.NET` Named Pipe streams (`System.IO.Pipes`). Commands start instantly with zero keystroke delay.
- **📁 Zero-Process Git Context (< 0.1ms):** Reads `.git/HEAD` directly in memory to extract branch names and repository details without spawning heavy `git.exe` child processes.
- **🛡️ Privacy-First Secret Sanitizer:** Automatically scrubs passwords (`--password`, `-p`), API keys (`sk-...`), personal access tokens (`ghp_...`), and Bearer auth headers before broadcasting to Discord.
- **⏱️ Smart Rate Limiting & Debouncing:** Buffers rapid command execution to strictly comply with Discord's 15-second IPC rate limit without dropping state transitions.
- **💤 Intelligent Idle & AFK Detection:** Automatically transitions from active command execution (`RUNNING`) to prompt waiting (`IDLE`), and marks as `AFK / Taking a break` after inactivity.
- **📦 Single Self-Contained Binary (< 3.6MB):** Compiled with stripped debug symbols and zero external runtime dependencies.

---

## 🚀 1-Line Quick Install

Run one command in your terminal. Sol will automatically download, configure your shell profile, and start the daemon silently in the background:

### 🪟 Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/ThinhDost/sol/main/install.ps1 | iex
```

### 🐧 Linux / 🍎 macOS (Bash & Zsh)
```bash
curl -fsSL https://raw.githubusercontent.com/ThinhDost/sol/main/install.sh | bash
```

> **That's it!** Discord Rich Presence is now active and will automatically update whenever you open a terminal and run commands.

---

### 🛠️ Manual Build (Optional)
If you prefer building from source:
```powershell
git clone https://github.com/ThinhDost/sol.git
cd sol
.\build.ps1
```

---

## ⚙️ Configuration

Sol works out of the box, but you can customize settings via `sol.config.json` or CLI flags:

```json
{
  "client_id": "1543849393432698940",
  "github_url": "https://github.com/ThinhDost/sol",
  "idle_timeout_minutes": 10,
  "rate_limit_seconds": 15,
  "mask_secrets": true,
  "anonymize_home_path": true
}
```

### CLI Options:
```text
Usage of sol-daemon:
  --client-id string      Discord Application Client ID
  --github-url string     GitHub repository link for interactive profile button
  --idle-timeout int      Minutes of inactivity before switching to AFK (default 10)
  --rate-limit int        Minimum seconds between Discord updates (default 15)
  --version               Print version and exit
```

---

## 🧪 Testing & Verification

Run automated test suite covering binary packet framing, Git parser, secret sanitizer, and local IPC:
```bash
go test -v ./...
```

---

## 🗺️ Roadmap

- [x] High-performance Go daemon with Windows Named Pipes & Unix sockets.
- [x] PowerShell hook via in-memory .NET streams.
- [x] Bash hook integration.
- [x] Zero-process `.git/HEAD` reader.
- [x] Privacy sanitizer with regex secret scrubbing.
- [x] Interactive Discord Profile button linking to GitHub.
- [ ] Zsh & Fish shell integrations.
- [ ] Customizable theme presets for status text.

---

## 📄 License

Distributed under the **MIT License**. See `LICENSE` for more information.
