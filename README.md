# Denver

**D**iscourse **ENV**ironments manag**ER**

A CLI tool for managing multiple isolated Discourse development environments.

## The Problem

Testing Discourse core changes with different plugin combinations is painful.
Existing tools (discourse-cp, dev containers) make it hard to maintain
multiple separate instances with different plugin sets.

## The Solution

Denver makes it easy to create, manage, and switch between multiple isolated
Discourse development environments. Uses git worktrees for instant environment
creation (less than 1 second after initial setup). Each environment gets its
own branch, plugin set, and isolated working directory.

## Prerequisites

Denver requires PostgreSQL and Redis to be running locally. You can install and run them however you prefer (system packages, Docker, Homebrew, etc.).

**PostgreSQL**: Must be running and accessible (default port 5432)
**Redis**: Must be running and accessible (default port 6379)

Example with Docker:
```bash
docker run -d --name postgres -p 5432:5432 -e POSTGRES_HOST_AUTH_METHOD=trust postgres:16
docker run -d --name redis -p 6379:6379 redis:7
```

Or use your system package manager, Homebrew, etc.

## Installation

**Download pre-built binary** (recommended):

Visit the [releases page](https://github.com/ducks/denver/releases) and download the binary for your platform:
- Linux (amd64, arm64)
- macOS (Intel, Apple Silicon)

Make it executable and move to your PATH:
```bash
chmod +x denver-*
sudo mv denver-* /usr/local/bin/denver
```

**Or install with Go**:

```bash
go install github.com/ducks/denver@latest
```

**Or build from source**:

```bash
git clone https://github.com/ducks/denver.git
cd denver
go build -o denver
```

## Quick Start

```bash
# Check prerequisites
denver doctor

# Create a minimal environment (first run clones bare repo, takes a few minutes)
denver create minimal --profile base

# Setup dependencies (one-time per environment)
denver setup minimal

# Start Rails and Ember servers
denver start minimal

# View logs
denver logs              # Rails logs
denver logs --ember      # Ember logs

# Check status
denver status

# List all environments
denver list

# Create and switch to another environment
denver create yaks --profile base --plugin discourse-yaks
denver setup yaks
denver switch yaks          # Stops minimal, starts yaks

# Stop servers
denver stop

# Destroy environments when done
denver destroy minimal
denver destroy yaks
```

## How It Works

**Git Worktrees**: Denver uses git worktrees to share a single bare repository
across all environments. The first `create` command clones discourse to
`~/.denver/discourse.git` (takes a few minutes). Subsequent environments are
created instantly as worktrees.

**Benefits**:
- First environment: ~3 minutes (one-time bare repo clone)
- Additional environments: <1 second
- Disk savings: ~400MB per environment (shared .git)
- Each environment gets its own branch (named after the environment)

## Profiles

Profiles define the base configuration for an environment. They live in
`~/.denver/profiles/` and are written in YAML.

### Example: base.yml

```yaml
name: Base
description: Minimal Discourse environment with no plugins

plugins: []

site_settings:
  title: "Discourse Local"

seed:
  admin: true
  sample_users: 5
  sample_topics: 10
```

### Example: full.yml

```yaml
name: Full
description: Full-featured Discourse with common plugins

plugins:
  - name: discourse-chat
    repo: discourse/discourse-chat
  - name: discourse-automation
    repo: discourse/discourse-automation
  - name: discourse-voting
    repo: discourse/discourse-voting

site_settings:
  title: "Discourse Local (Full)"
  chat_enabled: true

seed:
  admin: true
  sample_users: 20
  sample_topics: 50
```

## Commands

### create

Create a new environment from a profile.

```bash
denver create <name> --profile <profile> [flags]
```

Flags:
- `--profile, -p`: Profile to use (required)
- `--base`: Base branch to branch from (default: main)
- `--plugin`: Add plugins beyond profile (repeatable)
  - Simple name: `discourse-chat` (uses discourse org)
  - Full path: `ducks/discourse-invite-stats` (custom org)
  - With branch: `discourse-yaks:feature/new-stuff`
- `--local`: Use local plugin path for development (repeatable)
  - Format: `plugin-name:~/path/to/plugin`
  - Symlinks directly to your local plugin directory
  - Changes appear instantly without git push/pull

Examples:

```bash
# Basic environment (creates branch "yaks" from main)
denver create yaks --profile base

# Test a core PR (creates branch "test-buttons" from fix/button-refactor)
denver create test-buttons --profile base --base fix/button-refactor

# Add plugins to base profile
denver create yaks-dev --profile base --plugin discourse-yaks

# Add plugins from custom GitHub org
denver create invite-stats --profile base --plugin ducks/discourse-invite-stats

# Use local plugin for active development
denver create frndr --profile base --local discourse-frndr:~/dev/discourse-frndr

# Mix cloned and local plugins
denver create test --profile base --plugin discourse-chat --local discourse-yaks:~/dev/discourse-yaks

# Full environment with core branch
denver create test-epic --profile epic-games --base my-pr
```

**Note**: Each environment gets a unique git branch named after the environment.
The `--base` flag specifies which branch to branch from (default: main).

### setup

Install dependencies for an environment (run once after creating).

```bash
denver setup <name>
```

This runs:
- `bundle install` (Ruby gems)
- `pnpm install` (JavaScript dependencies)
- `bundle exec rake db:create db:migrate` (database setup)

### start

Start Rails and Ember servers for an environment.

```bash
denver start <name>
```

Servers run on:
- Rails: http://localhost:3000
- Ember: http://localhost:4200

Only one environment can run at a time.

### stop

Stop the currently running environment.

```bash
denver stop
```

### sync

Update discourse core and/or plugins by pulling latest changes from git.

```bash
denver sync <name> [target]
```

Targets:
- (none): Update discourse core and all remote plugins
- `discourse`: Update only discourse core
- `<plugin-name>`: Update only the specified plugin

Local plugins (created with `--local`) are skipped since they're symlinks to your development directories.

Examples:

```bash
# Update everything (discourse + all remote plugins)
denver sync frndr

# Update only discourse core
denver sync frndr discourse

# Update specific plugin
denver sync frndr discourse-chat

# Local plugin (will be skipped)
denver sync frndr discourse-yaks  # "Plugin is local, skipping sync"
```

### switch

Switch to a different environment (stops current, starts new).

```bash
denver switch yaks
```

Convenience command that combines `denver stop` and `denver start`. If the target environment is already running, does nothing.

### status

Show the currently running environment.

```bash
denver status
```

### logs

View server logs for the running environment.

```bash
denver logs           # Rails logs (last 100 lines)
denver logs -f        # Follow Rails logs
denver logs --ember   # Ember logs
denver logs --ember -f  # Follow Ember logs
```

### list

List all environments.

```bash
denver list
```

Shows environment names, modification times, and which is currently running.

### doctor

Check that all prerequisites are installed and running.

```bash
denver doctor
```

Checks for:
- git, ruby, bundler, node, pnpm
- PostgreSQL and Redis availability

### destroy

Destroy an existing environment.

```bash
denver destroy <name>
```

This removes the environment directory, git worktree, and git branch. If the branch has uncommitted changes, you'll be prompted before deletion.

## Directory Structure

```
~/.denver/
├── discourse.git/          # Bare repo (shared across environments)
├── profiles/
│   ├── base.yml
│   ├── full.yml
│   └── epic-games.yml
└── environments/
    ├── minimal/
    │   ├── discourse/      # Worktree (branch: minimal)
    │   ├── plugins/        # Cloned plugins
    │   │   └── discourse-chat/
    │   └── .denver.yml     # Environment config
    └── yaks/
        ├── discourse/      # Worktree (branch: yaks)
        │   └── plugins/
        │       └── discourse-yaks -> ~/dev/discourse-yaks  # Local plugin symlink
        ├── plugins/
        │   └── discourse-chat/  # Cloned plugin
        └── .denver.yml
```

**Plugin Storage:**
- **Cloned plugins**: Stored in `environments/<name>/plugins/` and symlinked into `discourse/plugins/`
- **Local plugins** (`--local`): Symlinked directly from your local dev directory to `discourse/plugins/`

The `.denver.yml` file tracks which plugins are local vs cloned, so future
commands (like `denver sync`) can handle them appropriately.

## Status

Version 20251109

**Implemented:**
- ✅ Profile loading from YAML
- ✅ Git worktrees for fast environment creation (<1 second)
- ✅ Bare repo sharing (~400MB saved per environment)
- ✅ Plugin cloning with symlinks
- ✅ Command-line plugin additions (`--plugin`)
- ✅ Custom GitHub org support (e.g., `ducks/discourse-invite-stats`)
- ✅ Local plugin development (`--local` flag for instant updates)
- ✅ Branch selection for discourse core
- ✅ Environment setup (bundle, pnpm, database)
- ✅ Server management (start, stop, status)
- ✅ Log viewing (Rails and Ember)
- ✅ List environments
- ✅ Doctor command (prerequisite checking)
- ✅ Smart environment destruction (prompts if branch has changes)

**Coming soon:**
- ⏳ Plugin branch management (`denver plugin` command)
- ⏳ Database cloning from staging environments
- ⏳ Nix shell auto-detection and wrapping
- ⏳ Multi-environment support (auto port allocation)

## Use Cases

**Testing Core PRs**: Create environment with specific plugin set to test core
changes.

```bash
denver create test-pr --profile epic-games --base fix/my-feature
```

**Plugin Development**: Work on plugins in isolation without affecting your
main dev environment.

```bash
# Clone plugin from GitHub
denver create yaks --profile base --plugin discourse-yaks

# Or use local plugin for active development (instant updates)
denver create yaks --profile base --local discourse-yaks:~/dev/discourse-yaks
```

The `--local` flag symlinks directly to your local plugin directory, so changes
appear instantly without needing to commit/push to GitHub. Perfect for active
plugin development.

**Multiple Projects**: Maintain separate environments for different plugins or
features.

```bash
denver create yaks --profile base --plugin discourse-yaks
denver create transit --profile base --plugin discourse-transit-tracker
```

## License

MIT
