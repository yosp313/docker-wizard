# Docker Wizard

Docker Wizard is a Go CLI/TUI tool that scaffolds a Docker development stack for your project. It detects the primary language in your working directory, generates a matching `Dockerfile`, and generates a `docker-compose.yml` based on the services you select.

## Features
- Step-by-step wizard UI with a progress bar header, status side panel, and animations
- Multiple run modes for local and automation workflows (`styled`, `plain`, `cli`, `batch`)
- Language + version detection with config-driven Dockerfile templates for Go, Node, Python, Ruby, PHP, Java, and .NET
- Category-based service selection (databases, queues, cache, analytics, proxies)
- Config-driven service catalog (edit `config/services.json`)
- Deterministic, reproducible compose output
- Safe file generation with user-priority merge mode (creates missing files and merges differing existing files)

## Quick start
```bash
# from this repo
go run .
```

## Install (curl)
```bash
curl -fsSL https://raw.githubusercontent.com/yosp313/docker-wizard/main/install.sh | sh
```

By default, the installer resolves and installs the latest GitHub Release tag.
Default install root is `~/.docker-wizard`.
To install an unreleased commit intentionally, set `REF=main`.

## CLI
```bash
docker-wizard --version
docker-wizard --mode styled
docker-wizard --mode plain
docker-wizard --mode cli
docker-wizard --mode batch --services mysql,redis --language go --dry-run
docker-wizard --mode batch --services all --write

# subcommands
docker-wizard add mysql redis kafka
docker-wizard add mysql --write
docker-wizard list
```

Run modes:
- `styled` (default): full TUI with Lip Gloss styling and animations
- `plain`: TUI flow with plain-text rendering for maximum terminal compatibility
- `cli`: line-by-line interactive prompts (non-TUI)
- `batch`: non-interactive automation mode driven by flags

Batch mode flags:
- `--services`: comma-separated service IDs (for example `mysql,redis`) or `all`
- `--language`: optional override (`go`, `node`, `python`, `ruby`, `php`, `java`, `dotnet`, `auto`)
- `--dry-run`: preview file status and warnings without writing (default behavior)
- `--write`: write generated files

### Subcommands

#### `docker-wizard add <service...>`
Incrementally add services to an existing `docker-compose.yml` without re-running the full wizard.

- Accepts one or more service IDs as positional arguments
- Dry-run by default — pass `--write` to apply changes
- Skips services already present in the compose file
- Auto-expands dependencies (e.g. `kafka` pulls in `zookeeper`)
- Generates compose-only output (no Dockerfile changes)
- Merges into existing compose using user-priority merge
- Creates a minimal compose if none exists

```bash
docker-wizard add mysql redis        # preview changes
docker-wizard add mysql redis --write # apply changes
```

#### `docker-wizard list`
Show available service IDs from the catalog, grouped by category.

```bash
docker-wizard list
```
## Usage flow
1. Start the wizard.
2. The tool detects your project language (you can override it).
3. Select services by category.
4. Review selections, warnings, and generated outputs.
5. Generate and run `docker compose up`.

## Key bindings
- `enter`: next/confirm
- `up`/`down`: move
- `space`: toggle service
- `b`: back
- `q`: quit
- `l`: choose language (detect step)
- `p`: preview (review step)
- `r`: retry (error step)
- `pgup`/`pgdown`/`home`/`end`: scroll preview

## Generated files
- `Dockerfile`
- `docker-compose.yml`
- `.dockerignore` (only if missing)

When generated files already exist:
- Matching files are left as-is
- Differing files are merged with user content taking priority
- Original differing files are backed up as `*.bak`
- The result screen reports `created`, `updated`, or `unchanged` per file

Compose merge behavior (user-priority):
- Existing keys/values are preserved; generated keys are added when missing
- `services.*.environment` merges by env key for list syntax (`- KEY=VALUE`), existing value wins
- `services.*.ports` merges by host port, existing host binding wins
- `services.*.depends_on`, `services.*.networks`, and `services.*.volumes` are existing-first set unions
- `environment` map syntax (`KEY: VALUE`) is not merged key-aware yet

## Releases
- Every push to `main` runs CI (`gofmt` check, `go vet`, `go test`, `go build`).
- If CI passes on `main`, the pipeline auto-tags the commit with the next patch (`vX.Y.Z`) and publishes a GitHub Release.
- Auto-tag uses a repository secret `RELEASE_PAT` (with repository contents write access) so tag pushes trigger the `Release` workflow.
- Tags use `vX.Y.Z` and start at `v0.1.0` if no tags exist.

## Configuration
- Service catalog lives in `config/services.json`.
- Dockerfile template catalog lives in `config/dockerfiles.json`.
- Edit `config/services.json` to add/remove services or change image tags, ports, and defaults.
- Edit `config/dockerfiles.json` to customize generated Dockerfiles per language.
- Services can declare categories, dependencies, and public exposure.
- See `docs/knowledge-base.md` for baseline conventions.

## Output conventions
- The compose file always includes an `app` service built from the local `Dockerfile`.
- Services are sorted for stable diffs.
- Volumes are declared when required by a service.
- Only services marked `public` in `config/services.json` publish host ports.
- Existing identical files are left unchanged.
- Existing differing files are merged and backed up as `*.bak`.
- Preview uses the same merge functions as write, so preview status/content matches write behavior.

## Dockerfile defaults
- Generated templates set `APP_START_CMD` and run `CMD ["sh", "-lc", "$APP_START_CMD"]`.
- Node uses `npm ci` when `package-lock.json` exists, otherwise `npm install`.
- Java and .NET templates use multi-stage builds by default.
- Templates are loaded from `config/dockerfiles.json`.

## Development
```bash
# build all packages
go build ./...

# run tests
go test ./...

# format
gofmt -w .
```

## Repository layout
- `main.go`: CLI/TUI entry point
- `internal/app`: application runtime
- `internal/cli`: interactive and non-interactive CLI flows
- `internal/generator`: language detection + compose/Dockerfile generation
- `internal/tui`: wizard UI
- `docs/knowledge-base.md`: product requirements and defaults
- `docs/prd.md`: product requirements document
- `docs/trd.md`: technical requirements and design document
- `docs/flow.md`: architecture diagram and module interaction map

## Documentation
- Knowledge base: `docs/knowledge-base.md`
- PRD: `docs/prd.md`
- TRD: `docs/trd.md`
- Architecture diagram: `docs/flow.md`
- Agent guidance: `AGENTS.md`
