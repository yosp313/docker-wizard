# Docker Wizard Technical Requirements and Design (TRD)

## Document info
- Product: Docker Wizard
- Type: Technical design document
- Version: 1.0
- Date: 2026-02-26
- Status: Current-state design baseline

## 1. Purpose
This document describes how Docker Wizard is implemented, how modules interact, and the technical decisions behind generation, preview, and write behavior.

It is intended for maintainers and contributors working on:
- CLI and TUI runtime behavior
- Dockerfile and compose generation
- Catalog configuration and extensibility
- File merge and backup mechanics

## 2. Scope
In scope:
- Entry points, mode routing, and runtime adapters
- Generator package structure and responsibilities
- Detection, template rendering, compose generation, validation, preview, write
- Determinism guarantees and merge semantics
- Testing and CI expectations

Out of scope:
- Product strategy and user goals (covered in `docs/prd.md`)
- Release process details beyond implementation dependencies
- Deployment orchestration and production infrastructure

## 3. High-level architecture

Execution layers:
1. Entry and mode dispatch
2. Runtime adapter (TUI, CLI interactive, batch)
3. Generator facade
4. Generator core packages
5. Config files and project filesystem

Primary code paths:
- `main.go` parses flags and validates mode/flag combinations.
- `internal/app/app.go` routes to the selected runtime mode.
- Runtime adapters call `internal/generator/generator.go` for core operations.
- Generator facade delegates to specialized packages for each concern.

Reference diagram: `docs/flow.md`

## 4. Runtime modes and control flow

### 4.1 Styled and plain TUI
- Entrypoint: `internal/tui/RunWizard`
- State model: `internal/tui/wizard.go`
- Views and styling: `internal/tui/view.go`, `internal/tui/styles.go`
- Command side effects: `internal/tui/commands.go`

Key behavior:
- Shared state machine supports both `styled` and `plain` render modes.
- Language detection occurs early and can be overridden by user choice.
- Service selection is category-based and progresses in fixed order.
- Review can transition to preview, then generate.
- Errors are represented as a dedicated step with retry/back options.
- Key handling is split by step-specific handlers to keep state transitions readable and testable.

### 4.2 CLI interactive
- Entrypoint: `internal/cli/RunInteractive`
- Prompt sequence mirrors the wizard intent without TUI rendering.
- Uses the same generator APIs as TUI to keep output behavior aligned.

### 4.3 Batch
- Entrypoint: `internal/cli/RunNonInteractive`
- Designed for CI/bootstrap automation.
- Validates options and defaults to dry run unless `--write` is set.
- Supports explicit language and service overrides.

## 5. Module design

### 5.1 Entry and app router
- `main.go`
  - Parses flags: mode, services, language, write, dry-run, version.
  - Rejects incompatible flag usage outside batch mode.
  - Normalizes service flag values (lowercase, trimmed, deduped).
- `internal/app/app.go`
  - Converts mode string to typed mode enum.
  - Dispatches to TUI, CLI interactive, or CLI batch runtime.

### 5.2 Generator facade
- `internal/generator/generator.go`
  - Thin API surface exported to runtime adapters.
  - Provides stable integration layer while core packages stay focused.
  - Re-exports core types to reduce adapter coupling.

### 5.3 Catalog and service model
- `internal/generator/catalog/catalog.go`
  - Loads `config/services.json` from project root, with executable-relative fallback.
  - Normalizes and validates service definitions.
  - Enforces unique IDs, valid categories, and existing dependency references.
- `internal/generator/catalog/services.go`
  - Defines `ServiceSpec` schema.

### 5.4 Language and version detection
- `internal/generator/dockerfile/detect.go`
  - Detects project type based on marker files and fixed priority order.
  - Detects language version hints where available.
  - Captures ecosystem-specific flags used by template rendering.

### 5.5 Dockerfile rendering
- `internal/generator/dockerfile/dockerfile.go`
  - Loads template lines from `config/dockerfiles.json`.
  - Renders line-by-line with `text/template` and strict key checking.
  - Applies defaults when version values are missing.
  - Encapsulates node package manager install/start command selection.

### 5.6 Compose generation
- `internal/generator/compose/compose.go`
  - Always includes baseline `app` service.
  - Validates selected IDs against catalog.
  - Expands required dependencies via `requires` traversal.
  - Renders deterministic output with sorted volumes and dependency lists.
  - Publishes host ports only for public services, otherwise uses `expose`.

### 5.7 Validation
- `internal/generator/validate/validate.go`
  - Produces user-facing warnings for:
    - Dependency inconsistencies
    - Host port collisions among public services
  - Includes baseline `app` service in port collision checks.

### 5.8 Preview and write
- `internal/generator/preview/preview.go`
  - Compares merged target content with existing files.
  - Produces per-file status (`new`, `same`, `different`, `exists`).
- `internal/generator/write/write.go`
  - Writes files via temp-file and atomic rename pattern.
  - Merge behavior:
    - Compose: user-priority YAML merge preserving existing keys/values and adding generated keys when missing.
    - Compose list semantics:
      - `services.*.environment` list entries merge by env key (`KEY=VALUE`), existing value wins.
      - `services.*.ports` merge by host port, existing host binding wins.
      - `services.*.depends_on`, `services.*.networks`, and `services.*.volumes` use existing-first set union.
      - `environment` map form (`KEY: VALUE`) is not key-aware merged in current version.
    - Dockerfile: preserve existing content; inject `ENV APP_START_CMD` and `CMD` only when needed.
  - Creates backups (`.bak`) for differing existing files.
  - Creates `.dockerignore` only when missing.

## 6. Data contracts

Core runtime contracts:
- `app.Options`
  - Mode + automation settings passed from `main.go` to app router.
- `cli.NonInteractiveOptions`
  - Batch options consumed by non-interactive runtime.
- `dockerfile.LanguageDetails`
  - Detection result carrying language type, feature flags, versions, and project metadata.
- `catalog.ServiceSpec`
  - Service definition schema from catalog.
- `compose.ComposeSelection`
  - Ordered service IDs selected for generation.
- `preview.Preview`
  - Pre-write status and generated content per managed file.
- `write.Output`
  - Final write statuses and backup paths.

## 7. Determinism and idempotency

Deterministic behaviors:
- Catalog-driven service ordering by `order` then `label`.
- Compose output ordering for services, dependencies, ports, and volumes.
- Stable YAML encoding for merged compose documents.

Idempotent behaviors:
- Re-running with same inputs should produce unchanged statuses.
- Existing identical files remain untouched.
- Existing differing files are updated in a controlled merge path with backups.
- Preview and write share merge logic, so preview status/content aligns with actual write outcomes.

## 8. Error handling model
- Validation and runtime functions return errors instead of panicking.
- Error context is wrapped at boundaries (for example, file read and merge failures).
- TUI exposes error step with retry and back navigation.
- CLI and batch modes print actionable failure messages and exit non-zero.

## 9. Security and safety considerations
- No secret storage in code paths.
- Safe-write strategy prevents partial write corruption.
- Backup creation mitigates accidental configuration loss.
- Batch mode requires explicit `--write` for mutating operations.

## 10. Performance considerations
- Designed for local repositories with small to moderate config sizes.
- Catalog load and generation are lightweight and synchronous.
- TUI keeps UI update loop responsive by isolating side effects in commands.

## 11. Testing strategy

Current automated coverage includes:
- Argument and mode parsing tests (`main_test.go`, `internal/app/app_test.go`)
- CLI helper behavior (`internal/cli/*_test.go`)
- TUI step behavior and preview interactions (`internal/tui/wizard_test.go`)
- Dockerfile template rendering behavior (`internal/generator/dockerfile/dockerfile_test.go`)
- Write and merge behavior (`internal/generator/write/write_test.go`)

Recommended expansion:
- Fixture-based integration tests for end-to-end generation across languages.
- Snapshot tests for compose and Dockerfile outputs.
- Additional merge edge cases for complex existing compose files.

## 12. CI and quality gates
- CI workflow validates:
  - `gofmt` formatting
  - `go vet`
  - `go test ./...`
  - `go build ./...`
- Successful `main` pipeline drives auto-tag and release workflows.

## 13. Extensibility guidelines

Add a new service:
1. Add service definition to `config/services.json`.
2. Provide valid category and dependency references.
3. Add/adjust tests for selection, warnings, and compose output.

Add a new language template:
1. Extend language detection in `internal/generator/dockerfile/detect.go` if needed.
2. Add template in `config/dockerfiles.json`.
3. Extend template data and defaults in `internal/generator/dockerfile/dockerfile.go`.
4. Add rendering tests in `internal/generator/dockerfile/dockerfile_test.go`.

## 14. Known tradeoffs
- Compose merge prefers preserving existing user config over strict regeneration purity.
- Warning-based validation does not block all runtime misconfigurations.
- Simplicity-first design avoids heavy abstraction layers, which keeps code direct but less plugin-oriented.

## 15. Related docs
- Product requirements: `docs/prd.md`
- Architecture diagram: `docs/flow.md`
- Product defaults and decisions: `docs/knowledge-base.md`
