# Quick Start Guide

## Installation

### Option 1: Build from source

```bash
# Clone the repository
git clone https://github.com/straticus1/macos-supply-chain-monitor
cd macos-supply-chain-monitor

# Download dependencies
make deps

# Build
make build

# Install (optional)
make install
```

### Option 2: Download pre-built binary (future)

```bash
curl -L https://github.com/straticus1/macos-supply-chain-monitor/releases/latest/download/scm-darwin-amd64 -o scm
chmod +x scm
sudo mv scm /usr/local/bin/
```

## First Run

### 1. Initialize Configuration

```bash
# Create config directory and sample config
make init-config

# Or manually
mkdir -p ~/.config/scm
cp config.example.yaml ~/.config/scm/config.yaml
```

### 2. Scan Your Current System

Get a baseline of what's installed:

```bash
scm scan
```

Output:
```
🔍 Scanning system...

📊 Scan Results:
─────────────────────────────────────────
  homebrew: 127 packages
  npm: 45 packages
  pip: 89 packages
─────────────────────────────────────────
  Total: 261 packages
```

### 3. Start the Monitoring Daemon

```bash
scm daemon start
```

Output:
```
🚀 Starting Supply-Chain Monitor daemon...
📊 Database: /Users/ryan/.config/scm/scm.db
✅ Daemon started successfully
📡 Monitoring package managers: [homebrew npm pip cargo go]

💡 Use 'scm status' to check activity
💡 Use 'scm history' to view recent events

⏸  Press Ctrl+C to stop
```

### 4. Install Something and Watch It Get Detected

In another terminal:

```bash
brew install htop
```

The daemon will log:
```
✓ homebrew installed htop (score: 15)
```

### 5. View History

```bash
scm history --last 1h
```

Output:
```
┌─────────────────────────────────────────────────────────────────┐
│ Supply-Chain Activity                                            │
└─────────────────────────────────────────────────────────────────┘

🟢 LOW RISK
  [14:32:15] homebrew htop
  └─ Version: unknown
  └─ Risk Score: 15/100

─────────────────────────────────────────────────────────────────
Risk Summary: 0 High, 0 Medium, 1 Low
```

## Common Commands

### Monitoring

```bash
# Start daemon (runs in foreground)
scm daemon start

# Check daemon status (not yet implemented)
scm daemon status
```

### History & Analysis

```bash
# View last 24 hours (default)
scm history

# View last hour
scm history --last 1h

# View last week
scm history --last 168h

# Show only high-risk events
scm history --risk high

# Limit results
scm history --limit 10
```

### Scanning

```bash
# Scan all package managers
scm scan

# Scan only Homebrew
scm scan --brew

# Scan only npm
scm scan --npm

# Verify binary hashes (future)
scm scan --verify
```

## Example Workflows

### Scenario 1: Daily Security Check

```bash
# What got installed today?
scm history --today

# Any high-risk installations?
scm history --today --risk high
```

### Scenario 2: Before/After Comparison

```bash
# Take snapshot before
scm scan > before.txt

# ... make changes ...

# Take snapshot after
scm scan > after.txt

# Compare
diff before.txt after.txt
```

### Scenario 3: Continuous Monitoring

```bash
# Run daemon in background (macOS)
nohup scm daemon start > ~/Library/Logs/scm-daemon.log 2>&1 &

# Check logs
tail -f ~/Library/Logs/scm-daemon.log

# View recent activity
scm history --last 1h
```

### Scenario 4: Alert on High-Risk

Edit `~/.config/scm/config.yaml`:

```yaml
alerts:
  risk_threshold: 70
  webhooks:
    - url: https://hooks.slack.com/services/YOUR/WEBHOOK
```

Now any installation with risk score >= 70 will trigger a webhook.

## What Gets Monitored?

### Package Managers
- ✅ **Homebrew**: `/usr/local/Cellar`, `/opt/homebrew/Cellar`
- ⏳ **npm**: Global packages (partially implemented)
- ⏳ **pip**: Python packages (partially implemented)
- ⏳ **cargo**: Rust binaries (planned)
- ⏳ **go**: Go binaries (planned)

### Risk Factors (Current MVP)
- Executable files
- Large binaries (> 50MB)
- Installation location

### Risk Factors (Planned)
- Unsigned binaries
- LaunchAgent/LaunchDaemon creation
- Post-install network connections
- Code signature changes
- Obfuscated scripts
- Suspicious paths

## Troubleshooting

### "No events found"

The daemon might not be running. Start it with:
```bash
scm daemon start
```

### "No valid paths to watch"

Homebrew might not be installed, or is in a non-standard location.

Check if Homebrew is installed:
```bash
which brew
```

### Database errors

Delete and recreate the database:
```bash
rm ~/.config/scm/scm.db
scm scan  # This will recreate it
```

## Next Steps

- **Configure alerts**: Edit `~/.config/scm/config.yaml` to add webhooks
- **Run continuously**: Set up launchd or cron to keep daemon running
- **Contribute**: This is MVP - many features are stubs waiting for implementation!

## Getting Help

```bash
# Show help
scm --help

# Command-specific help
scm history --help
scm scan --help
scm daemon --help
```

---

**Happy monitoring! Stay safe out there. 🔒**
