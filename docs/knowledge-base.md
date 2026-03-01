# Knowledge Base

## Purpose
- Product: TUI scaffold generator for Docker images and docker-compose
- Goal: let users select services and auto-generate a `docker-compose.yml` plus a project-specific `Dockerfile`
- Primary users: developers bootstrapping local dev stacks

## Current scope
- Service selection UI for compose generation
- Generate `docker-compose.yml` with the chosen services
- Detect the main project language
- Generate a matching `Dockerfile`
- Add the generated app service into `docker-compose.yml`
- Deterministic output and merge write mode for existing files

## Technology stack
- Language: Go
- TUI: Bubble Tea, Bubbles, Huh, Lip Gloss, Harmonica

## Engineering principles
- KISS, SOLID, DRY
- Go best practices
- Keep file organization simple and readable

## Defaults and conventions
### Compose output
- Compose version: 3.9
- Base services include an `app` service with `build.context: .` and `dockerfile: Dockerfile`
- Default app port mapping: `8080:8080`
- Ports are only published for services marked `public` in `config/services.json`
- Internal services use `expose` instead of host port publishing
- Services share the `app-net` network
- Volumes: declared for mysql, postgres, and caddy when selected

### Service catalog
- Defaults live in `config/services.json` and can be edited there
- Services declare categories, dependencies, and public exposure

### Dockerfile catalog
- Dockerfile templates live in `config/dockerfiles.json` and can be edited there
- Templates are selected by detected language and rendered with detected/default versions

### Service catalog additions
- MongoDB (database)
- Memcached (cache)
- Plausible (analytics) with bundled Clickhouse and Postgres services

### Dockerfile templates
- Go: multi-stage build from `golang:1.25-alpine` to `alpine:3.20`
- Node: `node:20-alpine` with npm/yarn/pnpm auto-detection (`npm ci` when `package-lock.json` exists)
- Python: `python:3.12-slim`, optional requirements install
- Ruby: `ruby:3.3-alpine` with bundler
- PHP: `php:8.3-fpm-alpine`
- Java: multi-stage build (Maven/Gradle builder to `eclipse-temurin:21-jre` runtime)
- .NET: multi-stage build (`dotnet/sdk` builder to `dotnet/aspnet` runtime)
- Fallback: `alpine:3.20`
- All templates set `APP_START_CMD` and run through `sh -lc` for command override support
- Template defaults are stored in `config/dockerfiles.json`

### File write behavior
- New files are created
- Existing identical files are marked unchanged
- Existing differing files are merged and marked updated
- Existing differing file contents are backed up to `*.bak`
- `.dockerignore` is created only when missing
- Merge is user-priority: existing values are preserved and generated values are additive
- Compose list merge rules:
  - `services.*.environment` merges by env key for list form (`KEY=VALUE`), existing value wins
  - `services.*.ports` merges by host port, existing host binding wins
  - `services.*.depends_on`, `services.*.networks`, `services.*.volumes` are existing-first set unions
  - Environment map form (`KEY: VALUE`) is not key-aware merged yet
- Preview uses the same merge functions as write for parity

## Detection rules
### Language detection priority
1. Go: `go.mod`
2. Node: `package.json` (detects package manager via lock file)
3. Python: `requirements.txt`, `pyproject.toml`, or `Pipfile`
4. Ruby: `Gemfile`
5. PHP: `composer.json` or `.php-version`
6. Java: `pom.xml`, `build.gradle`, `build.gradle.kts`
7. .NET: `*.csproj`
8. Fallback: unknown -> generic Dockerfile

### Version detection
- Go: `go.mod` `go` directive
- Node: `.nvmrc`, `.node-version`, or `package.json` `engines.node`
- Python: `.python-version`, `runtime.txt`, or `pyproject.toml` `requires-python`
- Ruby: `.ruby-version`
- PHP: `.php-version` or `composer.json` `config.platform.php`
- Java: `pom.xml` (`maven.compiler.release/source` or `java.version`), or Gradle toolchain/source compatibility
- .NET: `global.json` SDK version

## UX flow
### Wizard flow
- Welcome
- Detect language
- Optional language override
- Databases (optional)
- Message queues (optional)
- Cache (optional)
- Analytics (optional)
- Webservers / Proxies (optional)
- Review and generate
- Result

### Styled TUI layout
- Header: app title, subtitle, project name, and a progress bar showing current step out of total steps
- Side panel (visible when terminal width >= 100 columns): step number, stage name, language, service count, warnings, blockers, and a contextual tip
- Footer: context-sensitive key binding hints
- The header and side panel have no overlapping information; the header shows branding and progress, while the side panel shows session status

### Run modes
- Styled mode (`--mode styled`, default): full TUI styling
- Plain mode (`--mode plain`): same TUI flow with plain text rendering for terminal compatibility
- CLI interactive mode (`--mode cli`): prompt-driven non-TUI flow
- Batch mode (`--mode batch`): non-interactive flag-driven flow for CI/bootstrap usage

### Batch mode flags
- `--services`: comma-separated service IDs or `all`
- `--language`: optional language override (`go`, `node`, `python`, `ruby`, `php`, `java`, `dotnet`, `auto`)
- `--dry-run`: preview only (default when `--write` is not set)
- `--write`: write generated files

## Planned / TBD
- Extensibility: how to add a new service or language template
- Testing: unit tests, fixture projects, and snapshot tests
- Non-goals: no deployment or cloud orchestration
- Distribution: releases, install method, and supported OS/terminal constraints

## Subcommands

### `docker-wizard add <service...>` — incremental service addition
- Adds services to an existing `docker-compose.yml` without re-running the full wizard
- Accepts one or more service IDs as positional arguments
- Dry-run by default; pass `--write` to apply changes
- Skips services already present in the compose file (with notice)
- Auto-expands dependencies (e.g. kafka pulls in zookeeper) with notice
- Generates compose-only fragments (no app service, no Dockerfile changes)
- Merges into existing compose using the same user-priority merge infrastructure
- Creates a minimal compose file if none exists (no app service)
- Networks always use `app-net`

### `docker-wizard list` — show available services
- Lists all selectable services from the catalog grouped by category
- Uses category order and labels from `internal/utils/services.go`
