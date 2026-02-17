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
- Volumes: declared for mysql, postgres, and caddy when selected

### Service defaults
- mysql: `mysql:8.0`, port `3306:3306`, env `MYSQL_ROOT_PASSWORD=example`, volume `mysql-data`
- postgres: `postgres:16`, port `5432:5432`, env `POSTGRES_PASSWORD=example`, volume `postgres-data`
- redis: `redis:7-alpine`, port `6379:6379`
- analytics: `metabase/metabase:latest`, port `3000:3000`
- nginx: `nginx:alpine`, port `80:80`
- traefik: `traefik:v2.11`, ports `80:80`, `8080:8080`, command `--api.insecure=true`, `--providers.docker=true`
- caddy: `caddy:2`, ports `80:80`, `443:443`, volume `caddy-data`
- rabbitmq: `rabbitmq:3-management`, ports `5672:5672`, `15672:15672`
- kafka: `bitnami/kafka:3.7`, port `9092:9092`, depends on zookeeper
- zookeeper (auto-added when kafka is selected): `bitnami/zookeeper:3.9`, port `2181:2181`, env `ALLOW_ANONYMOUS_LOGIN=yes`

### Dockerfile templates
- Go: multi-stage build from `golang:1.25-alpine` to `alpine:3.20`
- Node: `node:20-alpine` with npm/yarn/pnpm auto-detection
- Python: `python:3.12-slim`, optional requirements install
- Ruby: `ruby:3.3-alpine` with bundler
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
5. Java: `pom.xml`, `build.gradle`, `build.gradle.kts`
6. .NET: `*.csproj`
7. Fallback: unknown -> generic Dockerfile

## UX flow (TBD)
- Step order and navigation (next/back/cancel)
- Validation and error messaging
- Preview and confirmation screens

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
