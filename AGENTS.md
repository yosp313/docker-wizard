# AGENTS

This file guides agentic coding tools operating in this repository.

## Repository snapshot
- Module: `docker-wizard`
- Go version: `1.25` (see `go.mod`)
- Entry point: `main.go`
- App package: `internal/app`
- TUI libraries (planned): bubbletea, bubbles, huh, lipgloss, harmonica

## Workspace layout
- `main.go` wires the CLI/TUI entry point.
- `internal/app` contains the core runtime logic.
- `docs/knowledge-base.md` captures requirements and decisions.
- Keep new packages in `internal/...` unless shared externally.

## Dependency management
- Prefer the standard library when possible.
- Add new dependencies only when needed and documented.
- Run `go mod tidy` after adding or removing dependencies.
- Avoid transitive dependency bloat for small helpers.


## Build, lint, test
No dedicated task runner or Makefile is present.
Use standard Go tooling.

### Build
- Build all packages: `go build ./...`
- Build main binary: `go build .`
- Run locally: `go run .`

### Lint / static checks
- Format check/fix: `gofmt -w .`
- Vet: `go vet ./...`
- If additional linters are added later, document them here.

### Tests
- Run all tests: `go test ./...`
- Run a single package: `go test ./internal/app`
- Run a single test: `go test ./internal/app -run '^TestName$' -count=1`
- Run a single subtest: `go test ./internal/app -run '^TestName$/Subtest$' -count=1`
- Verbose output: `go test ./... -v`

## Code style guidelines
Follow Go conventions and keep code simple and readable.

### Formatting
- Use `gofmt` for all Go files.
- Keep line length reasonable; prefer clear naming over line breaks.
- Avoid unnecessary alignment or tabular formatting.

### Imports
- Use standard Go import grouping: stdlib, then third-party, then local.
- Keep import blocks minimal; avoid unused imports.
- Prefer explicit imports over dot or blank imports, unless required.


### Package layout
- Keep packages small and cohesive.
- Prefer `internal/...` for app-only packages.
- Avoid deep nesting; two levels is usually enough.


### Naming
- Use idiomatic Go names: short, clear, lowerCamel for locals.
- Exported identifiers must have clear, concise names.
- Avoid stutter with package names (e.g., `app.App` is fine, `app.AppApp` is not).
- Use `NewX` for constructors when needed.


### Types
- Prefer concrete types; use interfaces only when they add clarity or testability.
- Keep structs small; group related fields.
- Use type aliases sparingly; prefer new types when behavior differs.

### Interfaces
- Define interfaces at the point of use, not the point of implementation.
- Keep interfaces minimal; avoid "fat" interfaces.

### Error handling
- Return errors, do not panic for expected failures.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Handle errors immediately; avoid silent ignores.
- Keep error messages lowercase and without punctuation.

### Input/output
- Avoid printing from deep helpers; return errors and data instead.
- Keep filesystem access behind focused functions for testability.
- Prefer deterministic output ordering for generated files.

### Functions
- Keep functions small and focused.
- Prefer early returns to reduce nesting.
- Avoid side effects in helpers unless clearly documented.

### Logging / output
- Use structured output where practical.
- Avoid global loggers unless needed by the TUI.
- Keep user-facing messages concise and actionable.

### CLI/TUI behavior
- Keep prompts clear and scoped to a single decision.
- Provide sane defaults and allow back navigation when possible.
- Avoid blocking operations on the main UI loop.

### Testing style
- Use table-driven tests where multiple cases exist.
- Keep tests deterministic; avoid time-based flakiness.
- Name tests as `TestXxx` and subtests with clear labels.


## TUI-specific guidance
- Keep model/state transitions pure and explicit.
- Separate UI rendering from business logic.
- Prefer small components with clear message handling.
- Avoid global state; pass dependencies into models.

## Compose/Dockerfile generation
- Generate deterministic, stable output for reproducible diffs.
- Sort services, volumes, and networks consistently.
- Keep templates minimal and easy to extend.
- Avoid hardcoding platform-specific paths.
- Validate user selections before writing files.

## Documentation
- Update `docs/knowledge-base.md` when requirements change.
- Keep README concise; link to deeper docs when needed.

## Git and workflow
- Do not amend commits unless explicitly requested.
- Avoid force pushes and history rewrites.
- Keep commits scoped to a single logical change.


## Repository rules
- No Cursor rules found in `.cursor/rules/` or `.cursorrules`.
- No Copilot rules found in `.github/copilot-instructions.md`.

## Safety and hygiene
- Do not rewrite history unless explicitly requested.
- Do not remove or revert unrelated changes.
- Do not add secrets or credentials to the repo.

## When in doubt
- Prefer the simplest viable implementation.
- Follow Go best practices and standard library patterns.
- Ask a single focused question only if blocked.
