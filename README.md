# Docker Wizard

Docker Wizard is a Go TUI that scaffolds a Docker development stack for your project. It detects the primary language in your working directory, generates a matching `Dockerfile`, and builds a `docker-compose.yml` based on the services you select.

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

The wizard generates:
- `Dockerfile`
- `docker-compose.yml`

## Usage flow
1. Start the wizard.
2. The tool detects your project language (you can override it).
3. Select services by category (databases, queues, cache, analytics, proxies).
4. Review selections, warnings, and generated outputs.
5. Generate and run `docker compose up`.

## Output conventions
- The compose file always includes an `app` service built from the local `Dockerfile`.
- Services are sorted for stable diffs.
- Volumes are declared when required by a service.
- Only services marked `public` in `config/services.json` publish host ports.

## Service defaults
- Service catalog lives in `config/services.json`.
- Edit this file to add/remove services or change image tags, ports, and defaults.
- Services can declare categories, dependencies, and public exposure.
- See `docs/knowledge-base.md` for the current baseline conventions.

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
