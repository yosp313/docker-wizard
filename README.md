# Docker Wizard

Docker Wizard is a Go TUI that scaffolds a Docker development stack for your project. It detects the primary language in your working directory, generates a matching `Dockerfile`, and generates a `docker-compose.yml` based on the services you select.

## Features
- Step-by-step wizard UI with a polished header/footer and animations
- Language + version detection with Dockerfile templates for Go, Node, Python, Ruby, PHP, Java, and .NET
- Category-based service selection (databases, queues, cache, analytics, proxies)
- Config-driven service catalog (edit `config/services.json`)
- Deterministic, reproducible compose output
- Safe file generation that avoids overwriting existing files

## Quick start
```bash
# from this repo
go run .
```

## Install (curl)
```bash
curl -fsSL https://raw.githubusercontent.com/yosp313/docker-wizard/main/install.sh | sh
```

## CLI
```bash
docker-wizard --version
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

## Releases
- Every push to `main` auto-tags a new patch release and publishes a GitHub Release.
- Tags use `vX.Y.Z` and start at `v0.1.0` if no tags exist.

## Configuration
- Service catalog lives in `config/services.json`.
- Edit this file to add/remove services or change image tags, ports, and defaults.
- Services can declare categories, dependencies, and public exposure.
- See `docs/knowledge-base.md` for baseline conventions.

## Output conventions
- The compose file always includes an `app` service built from the local `Dockerfile`.
- Services are sorted for stable diffs.
- Volumes are declared when required by a service.
- Only services marked `public` in `config/services.json` publish host ports.
- Existing files are never overwritten.

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
- `internal/generator`: language detection + compose/Dockerfile generation
- `internal/tui`: wizard UI
- `docs/knowledge-base.md`: product requirements and defaults

## Documentation
- Knowledge base: `docs/knowledge-base.md`
- Agent guidance: `AGENTS.md`
