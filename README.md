# Denver

**D**iscourse **ENV**ironments manag**ER**

A CLI tool for managing multiple isolated Discourse development environments.

## The Problem

Testing Discourse core changes with different plugin combinations is painful.
Existing tools (discourse-cp, dev containers) make it hard to maintain
multiple separate instances with different plugin sets.

## The Solution

Denver makes it easy to create, manage, and switch between multiple isolated
Discourse development environments. Each environment gets its own cloned
discourse repo, plugin set, and dev container configuration.

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
# Create a minimal environment
denver create minimal --profile base

# Create environment with custom branch
denver create test-pr --profile base --branch fix/my-feature

# Destroy an environment
denver destroy minimal
```

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
- `--branch, -b`: Discourse core branch to checkout
- `--plugin`: Add or override plugin (format: name:branch, repeatable)

Examples:

```bash
# Basic environment
denver create yaks --profile base

# Test a core PR
denver create test-buttons --profile base --branch fix/button-refactor

# Add plugins to base profile
denver create yaks-dev --profile base --plugin discourse-yaks:feature/new-stuff

# Full environment with core branch
denver create test-epic --profile epic-games --branch my-pr
```

### destroy

Destroy an existing environment.

```bash
denver destroy <name>
```

This removes the entire environment directory. Cannot be undone.

## Directory Structure

```
~/.denver/
├── profiles/
│   ├── base.yml
│   ├── full.yml
│   └── epic-games.yml
└── environments/
    ├── minimal/
    │   ├── discourse/
    │   └── .denver.yml
    └── yaks/
        ├── discourse/
        └── .denver.yml
```

## Status

Version 20251109

Currently implemented:
- ✅ Profile loading from YAML
- ✅ Environment creation with discourse cloning
- ✅ Branch selection for discourse core
- ✅ Environment destruction

Coming soon:
- ⏳ Plugin cloning based on profiles
- ⏳ Plugin branch overrides
- ⏳ Dev container config generation
- ⏳ List environments command
- ⏳ Open environment command

## Use Cases

**Testing Core PRs**: Create environment with specific plugin set to test core
changes.

```bash
denver create test-pr --profile epic-games --branch fix/my-feature
```

**Plugin Development**: Work on plugins in isolation without affecting your
main dev environment.

```bash
denver create yaks --profile base --plugin discourse-yaks:feature/new-stuff
```

**Multiple Projects**: Maintain separate environments for different plugins or
features.

```bash
denver create yaks --profile base --plugin discourse-yaks
denver create transit --profile base --plugin discourse-transit-tracker
```

## License

MIT
