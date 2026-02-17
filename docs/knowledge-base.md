# Knowledge Base

## Project overview
- Product: TUI scaffold generator for Docker images and docker-compose
- Goal: let users select services and auto-generate a `docker-compose.yml` plus a project-specific `Dockerfile`
- Primary users: developers bootstrapping local dev stacks

## Technology stack
- Language: Go
- TUI: Bubble Tea, Bubbles, Huh, Lip Gloss, Harmonica

## Core features
- Service selection UI for compose generation
- Generate `docker-compose.yml` with the chosen services
- Detect the main project language
- Generate a matching `Dockerfile`
- Add the generated app service into `docker-compose.yml`

## Initial service catalog
- MySQL
- PostgreSQL
- Redis
- Analytics tool
- Nginx
- Traefik
- Caddy
- RabbitMQ
- Kafka

## Engineering principles
- KISS, SOLID, DRY
- Go best practices
- Keep file organization simple and readable

## Defaults and conventions
### Compose structure
- Compose version: 3.9
- Base services include an `app` service with `build.context: .` and `dockerfile: Dockerfile`
- Default app port mapping: `8080:8080`
- Ports are only published for services marked `public` in `config/services.json`
- Internal services use `expose` instead of host port publishing
- Services share the `app-net` network
- Volumes: declared for mysql, postgres, and caddy when selected

### Service defaults
Defaults are defined in `config/services.json` and can be edited there.

### Service catalog additions
- MongoDB (database)
- Memcached (cache)
- Plausible (analytics) with bundled Clickhouse and Postgres services

### Dockerfile templates
- Go: multi-stage build from `golang:1.25-alpine` to `alpine:3.20`
- Node: `node:20-alpine` with npm/yarn/pnpm auto-detection
- Python: `python:3.12-slim`, optional requirements install
- Ruby: `ruby:3.3-alpine` with bundler
- PHP: `php:8.3-fpm-alpine`
- Java: `eclipse-temurin:21-jre`
- .NET: `mcr.microsoft.com/dotnet/aspnet:8.0`
- Fallback: `alpine:3.20`

## Detection rules
### Language detection
Priority order:
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

## UX flow (TBD)
### Wizard flow
- Welcome
- Detect language
- Databases (optional)
- Message queues (optional)
- Cache (optional)
- Analytics (optional)
- Webservers / Proxies (optional)
- Review and generate
- Result

## Extensibility (TBD)
- How to add a new service
- How to add a new language template

## Testing (TBD)
- Unit tests for generation logic
- Fixture projects for language detection
- Snapshot tests for compose/Dockerfile output

## Non-goals (TBD)
- No deployment or cloud orchestration
- No container runtime management beyond compose generation

## Distribution (TBD)
- Release artifacts and install method
- Supported OS/terminal constraints
