# Denver - AI Agent Context

This file provides context for AI coding assistants (Claude, Cursor, etc.)
working on the Denver project.

## What is Denver?

Denver (Discourse ENVironments managER) is a CLI tool for managing multiple
isolated Discourse development environments. It solves the problem of testing
core changes or working on plugins with different plugin combinations.

## Architecture

**Language**: Go 1.23
**CLI Framework**: cobra
**Config Format**: YAML (gopkg.in/yaml.v3)
**Versioning**: Date-based (YYYYMMDD format, e.g., 20251109)

**Directory Structure**:
```
denver/
├── main.go                    # Entry point
├── cmd/                       # Cobra commands
│   ├── root.go               # Root command
│   ├── create.go             # Create environments
│   └── destroy.go            # Destroy environments
└── internal/
    └── config/
        └── profile.go        # Profile loading and config
```

**User Data Structure**:
```
~/.denver/
├── profiles/                  # YAML profile templates
│   ├── base.yml
│   └── full.yml
└── environments/             # Created environments
    └── <name>/
        ├── discourse/        # Cloned discourse repo
        └── .denver.yml       # Environment metadata
```

## Core Concepts

**Profile**: YAML template defining plugins, themes, settings, and seed data.
Lives in `~/.denver/profiles/<name>.yml`. Reusable blueprint.

**Environment**: Specific instance created from a profile. Lives in
`~/.denver/environments/<name>/`. Contains cloned discourse repo and plugins.

**Plugin Override**: Command-line flag to add or override plugin branches from
the profile. Format: `--plugin name:branch`

## Implementation Status

### Completed (20251109)
- Profile loading from YAML
- Environment creation with discourse cloning
- Branch selection for discourse core (`--branch` flag)
- Environment destruction

### Next Steps (Priority Order)
1. Plugin cloning based on profile
2. Plugin branch overrides (`--plugin` flag parsing and application)
3. Dev container config generation (unique ports/volumes per environment)
4. List command (show all environments)
5. Open command (launch VSCode devcontainer)

## Code Patterns

**Error Handling**: Use `fmt.Errorf` with `%w` for error wrapping. Return
errors up the stack, handle at command level.

**Command Structure**: Each cobra command in separate file in `cmd/`. Init
function registers with rootCmd.

**Paths**: Use `filepath.Join` for cross-platform paths. Get home dir with
`os.UserHomeDir()`.

**Git Operations**: Use `exec.Command` to shell out to git. Pipe stdout/stderr
to os.Stdout/os.Stderr for live output.

## Design Philosophy

**First Version Simplicity**:
- Just clone discourse (don't optimize disk usage yet)
- No git worktrees initially (can add later)
- Focus on solving the real problem (multiple environments)

**User Experience**:
- Clear error messages
- Live output for long operations (git clone)
- Success indicators (✓ checkmarks)

**Terminology**:
- "profile" not "base" (though base.yml is a profile name)
- "destroy" not "delete" or "remove"
- "environment" not "instance" or "workspace"

## Common Tasks

**Adding a New Command**:
1. Create `cmd/<command>.go`
2. Define cobra.Command with Use, Short, Long, Args, RunE
3. Add `rootCmd.AddCommand(<command>Cmd)` in init()
4. Implement the actual logic in a separate function

**Adding Profile Fields**:
1. Update structs in `internal/config/profile.go`
2. Add YAML tags
3. Update example profiles in `~/.denver/profiles/`

**Testing Locally**:
```bash
go build -o denver
./denver create test --profile base
./denver destroy test
```

## Known Limitations

- Only clones from GitHub (no local repo support yet)
- No validation of profile YAML structure
- No environment status checking
- No plugin dependency resolution
- No automatic port assignment (manual in devcontainer.json for now)

## Future Considerations

**Git Worktrees**: Could use worktrees to share .git objects between
environments. Saves disk space but adds complexity.

**Database Cloning**: Could clone databases from staging environments for
realistic testing.

**Plugin Dependencies**: Some plugins require other plugins. Could add
dependency resolution.

**Site Settings Automation**: Could automatically apply site settings from
profile after environment creation.

## Development Environment

Developed using Nix shell with Go 1.23.

```bash
cd ~/dev/denver
nix-shell
```
