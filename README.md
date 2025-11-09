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

```bash
go install github.com/ducks/denver@latest
```

Or build from source:

```bash
git clone https://github.com/ducks/denver.git
cd denver
go build -o denver
```

## Quick Start

```bash
# Create a minimal environment (first run clones bare repo, takes a few minutes)
denver create minimal --profile base

# Create another environment (instant - uses worktree)
denver create test-pr --profile base --branch fix/my-feature

# Add plugins to an environment
denver create with-chat --profile base --plugin discourse-chat

# Destroy an environment
denver destroy minimal
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
- `--branch, -b`: Base branch for worktree (default: main)
- `--plugin`: Add plugins beyond profile (format: name, repeatable)

Examples:

```bash
# Basic environment (creates branch "yaks" from main)
denver create yaks --profile base

# Test a core PR (creates branch "test-buttons" from fix/button-refactor)
denver create test-buttons --profile base --branch fix/button-refactor

# Add plugins to base profile
denver create yaks-dev --profile base --plugin discourse-yaks

# Full environment with core branch
denver create test-epic --profile epic-games --branch my-pr
```

**Note**: Each environment gets a unique git branch named after the environment.
The `--branch` flag specifies which branch to base it on (default: main).

### destroy

Destroy an existing environment.

```bash
denver destroy <name>
```

This removes the entire environment directory. Cannot be undone.

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
    │   └── .denver.yml
    └── yaks/
        ├── discourse/      # Worktree (branch: yaks)
        ├── plugins/
        │   └── discourse-yaks/
        └── .denver.yml
```

Plugins are cloned to `environments/<name>/plugins/` and symlinked into
`discourse/plugins/` for each environment.

## Status

Version 20251109

Currently implemented:
- ✅ Profile loading from YAML
- ✅ Git worktrees for fast environment creation (<1 second)
- ✅ Bare repo sharing (~400MB saved per environment)
- ✅ Plugin cloning with symlinks
- ✅ Command-line plugin additions (`--plugin`)
- ✅ Branch selection for discourse core
- ✅ Environment destruction

Coming soon:
- ⏳ Plugin branch management (`denver plugin` command)
- ⏳ Database service management (auto-start postgres/redis with Docker)
- ⏳ Dev container config generation
- ⏳ List environments command
- ⏳ Open environment command (launch VSCode)

## Use Cases

**Testing Core PRs**: Create environment with specific plugin set to test core
changes.

```bash
denver create test-pr --profile epic-games --branch fix/my-feature
```

**Plugin Development**: Work on plugins in isolation without affecting your
main dev environment.

```bash
denver create yaks --profile base --plugin discourse-yaks
```

**Multiple Projects**: Maintain separate environments for different plugins or
features.

```bash
denver create yaks --profile base --plugin discourse-yaks
denver create transit --profile base --plugin discourse-transit-tracker
```

## License

MIT
