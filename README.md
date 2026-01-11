# macOS Supply-Chain Integrity Monitor

**A developer-focused security tool that watches package managers and tracks system changes to detect supply-chain attacks.**

## Overview

The macOS Supply-Chain Integrity Monitor is a lightweight daemon that monitors software installations across multiple package managers (Homebrew, npm, pip, cargo, go) and tracks critical system changes to detect potential supply-chain compromises.

## The Problem

Most macOS malware arrives via:
- Compromised Homebrew packages
- Malicious npm/pip dependencies
- `curl | bash` installation scripts
- Dev toolchain exploits

Current gap: **No tool provides visibility into what packages actually installed or changed on your system.**

## Key Features

### 1. Multi-Package Manager Tracking
- **Homebrew** (brew install/upgrade)
- **npm** (npm install -g)
- **pip** (pip install)
- **cargo** (cargo install)
- **go install**

### 2. Binary Integrity Monitoring
- Track SHA-256 hashes of all installed binaries
- Detect code-signing identity changes
- Alert on unsigned binary drops
- Monitor binary modifications

### 3. Persistence Detection
- LaunchAgents/LaunchDaemons creation
- Login Items additions
- Cron job modifications
- Shell profile changes (.zshrc, .bashrc)

### 4. Post-Install Script Analysis
- Capture all post-install script execution
- Flag suspicious behaviors:
  - Network connections
  - File downloads
  - Privilege escalation attempts
  - Hidden file creation

### 5. Git-Style History
- Timeline of all system changes
- Diff support: "What changed between yesterday and today?"
- Rollback information

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   CLI Interface (scm)                     │
│  Commands: status, history, diff, watch, alert, scan    │
└─────────────────┬───────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────┐
│                  Core Monitor Daemon                      │
│  - Event processing                                       │
│  - Package manager hooks                                  │
│  - File system watcher                                    │
└─────────────────┬───────────────────────────────────────┘
                  │
         ┌────────┼────────┐
         │        │        │
    ┌────▼───┐ ┌─▼────┐ ┌▼─────┐
    │ Brew   │ │ npm  │ │ pip  │  ... (Package Watchers)
    │ Watcher│ │Watch │ │Watch │
    └────┬───┘ └─┬────┘ └┬─────┘
         │       │       │
    ┌────▼───────▼───────▼─────┐
    │   Event Collection Bus    │
    └────┬──────────────────────┘
         │
    ┌────▼──────────────────────┐
    │   SQLite Event Store      │
    │  - Install events          │
    │  - File changes            │
    │  - Binary hashes           │
    │  - Code signatures         │
    └───────────────────────────┘
```

## Data Model

### Events Table
```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT, -- 'install', 'modify', 'persist', 'script'
    package_manager TEXT, -- 'brew', 'npm', 'pip', etc.
    package_name TEXT,
    package_version TEXT,
    risk_score INTEGER, -- 0-100
    details JSON
);
```

### Binaries Table
```sql
CREATE TABLE binaries (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE,
    sha256 TEXT,
    code_sign_identity TEXT,
    first_seen DATETIME,
    last_modified DATETIME,
    package_source TEXT,
    risk_flags JSON
);
```

### Persistence Items
```sql
CREATE TABLE persistence (
    id INTEGER PRIMARY KEY,
    type TEXT, -- 'launchagent', 'launchdaemon', 'loginitem', 'cron'
    path TEXT,
    target_binary TEXT,
    created_at DATETIME,
    created_by_event INTEGER REFERENCES events(id),
    is_suspicious BOOLEAN
);
```

## Risk Scoring

Each event gets scored 0-100 based on:

- **High Risk (70-100)**:
  - Unsigned binary installed
  - LaunchAgent created by install script
  - Network activity during install to non-standard hosts
  - Obfuscated script content
  - Modified existing system binary

- **Medium Risk (40-69)**:
  - Code-sign identity changed
  - Post-install script with elevated privileges
  - Binary installed outside standard paths
  - Package from unknown repository

- **Low Risk (0-39)**:
  - Normal package manager install
  - Signed binary from known developer
  - No persistence mechanisms
  - Standard installation paths

## CLI Commands

### Basic Operations
```bash
# Start monitoring daemon
scm daemon start

# Check status
scm status

# View recent activity
scm history --last 24h

# Show all installations today
scm list --today --type install
```

### Diff and Comparison
```bash
# What changed since yesterday?
scm diff --since yesterday

# Compare two snapshots
scm diff --from 2026-01-10 --to 2026-01-11

# Show only high-risk changes
scm diff --risk high
```

### Scanning
```bash
# Scan current system state
scm scan

# Check specific package manager
scm scan --brew

# Verify all binary hashes
scm verify --all
```

### Alerts
```bash
# Configure alerts
scm alert config --webhook https://example.com/hook

# Test alert system
scm alert test

# View alert history
scm alert history
```

## Example Output

```bash
$ scm history --last 1h

┌─────────────────────────────────────────────────────────────────┐
│ Supply-Chain Activity (Last 1 hour)                             │
└─────────────────────────────────────────────────────────────────┘

🔴 HIGH RISK
  [14:32:15] brew install suspicious-cli
  └─ ⚠️  Unsigned binary: /usr/local/bin/suspicious-cli
  └─ ⚠️  Created LaunchAgent: ~/Library/LaunchAgents/com.suspicious.plist
  └─ ⚠️  Post-install script contacted: 192.168.1.100:4444

🟡 MEDIUM RISK
  [14:25:03] npm install -g @company/internal-tool
  └─ ⚠️  Code-sign identity changed
  └─ ℹ️  Binary: /usr/local/bin/internal-tool

🟢 LOW RISK
  [14:15:42] brew upgrade node
  └─ ✓ Signed by: Homebrew
  └─ ✓ Hash verified: abc123...

─────────────────────────────────────────────────────────────────
Risk Summary: 1 High, 1 Medium, 1 Low
```

## Installation

```bash
# Install via Homebrew (future)
brew tap afterdark/security
brew install supply-chain-monitor

# Or build from source
git clone https://github.com/afterdark/supply-chain-monitor
cd supply-chain-monitor
make install
```

## Configuration

Config file: `~/.config/scm/config.yaml`

```yaml
# Package managers to monitor
package_managers:
  - homebrew
  - npm
  - pip
  - cargo
  - go

# Alert thresholds
alerts:
  risk_threshold: 70  # Alert on high risk only
  webhooks:
    - url: https://hooks.slack.com/services/YOUR/WEBHOOK
    - url: https://discord.com/api/webhooks/YOUR/WEBHOOK

# Monitoring paths
watch_paths:
  - /usr/local/bin
  - /opt/homebrew/bin
  - ~/.cargo/bin
  - ~/go/bin

# Database location
database: ~/.config/scm/scm.db

# Logging
log_level: info
log_file: ~/.config/scm/scm.log
```

## Technical Stack

- **Language**: Go 1.21+
- **Database**: SQLite3 with FTS5 for search
- **File Monitoring**: FSEvents (macOS native)
- **Code Signing**: Use `codesign` command-line tool
- **Package Hooks**:
  - Homebrew: Monitor `/usr/local/Cellar` and `/opt/homebrew/Cellar`
  - npm: Monitor global node_modules
  - pip: Monitor site-packages
  - cargo: Monitor `~/.cargo/bin`

## Security Considerations

1. **Daemon runs as user** (not root) to minimize attack surface
2. **Database encrypted at rest** (optional, via SQLCipher)
3. **No network access required** for core functionality
4. **Webhooks use TLS** with certificate validation
5. **Audit log** for all daemon operations

## Roadmap

### MVP (Week 1)
- [x] Core daemon architecture
- [ ] Homebrew monitoring
- [ ] Basic CLI (status, history, scan)
- [ ] SQLite event storage
- [ ] Binary hash tracking

### V1.0 (Week 2-3)
- [ ] npm, pip, cargo, go monitoring
- [ ] Risk scoring engine
- [ ] LaunchAgent detection
- [ ] Diff functionality
- [ ] Webhook alerts

### V1.1 (Week 4+)
- [ ] Code signing verification
- [ ] Post-install script analysis
- [ ] TUI (terminal UI) with charts
- [ ] Export to SIEM formats
- [ ] Integration with security tools

### Future
- [ ] Cloud sync (optional)
- [ ] Team dashboards
- [ ] ML-based anomaly detection
- [ ] Automatic rollback
- [ ] Integration with Veribits for hash lookups

## Use Cases

### Individual Developers
"I want to know if a package I installed did something sketchy."

### Security Researchers
"Track all software installations during malware analysis."

### Teams
"Audit what got installed on developer machines this week."

### Incident Response
"What changed on this system before the breach?"

## Why This Will Work

1. **Real pain point**: Developers have zero visibility today
2. **Simple UX**: Git-like commands, familiar mental model
3. **Local-first**: No cloud required, privacy-preserving
4. **Open source**: Build trust, get contributions
5. **Extensible**: Easy to add new package managers

## Differentiation

| Feature | SCM | Existing Tools |
|---------|-----|----------------|
| Package manager awareness | ✅ | ❌ |
| Developer-focused UX | ✅ | ❌ |
| Local-only operation | ✅ | ❌ (most are cloud EDR) |
| Supply-chain specific | ✅ | ❌ (generic monitoring) |
| Open source | ✅ | ❌ (most are commercial) |
| Zero-config start | ✅ | ❌ |

## License

MIT License (for maximum adoption)

## Contributing

Contributions welcome! See CONTRIBUTING.md

---

**Built with ❤️ for developers who care about supply-chain security**
