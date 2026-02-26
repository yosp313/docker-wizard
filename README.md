# Docker Wizard

Docker Wizard is a Go CLI/TUI tool that scaffolds a Docker development stack for your project. It detects the primary language in your working directory, generates a matching `Dockerfile`, and generates a `docker-compose.yml` based on the services you select.

## Features
- Step-by-step wizard UI with a polished header/footer and animations
- Multiple run modes for local and automation workflows (`styled`, `plain`, `cli`, `batch`)
- Language + version detection with config-driven Dockerfile templates for Go, Node, Python, Ruby, PHP, Java, and .NET
- Category-based service selection (databases, queues, cache, analytics, proxies)
- Config-driven service catalog (edit `config/services.json`)
- Deterministic, reproducible compose output
- Safe file generation with merge mode (creates missing files and merges differing existing files)

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
- Differing files are merged with generated output
- Original differing files are backed up as `*.bak`
- The result screen reports `created`, `updated`, or `unchanged` per file

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
- `docs/flow.md`: architecture diagram and module interaction map
- `docs/mindmap.md`: flow-oriented process map

## Documentation
- Knowledge base: `docs/knowledge-base.md`
- PRD: `docs/prd.md`
- Architecture diagram: `docs/flow.md`
- Flow map: `docs/mindmap.md`
- Agent guidance: `AGENTS.md`
