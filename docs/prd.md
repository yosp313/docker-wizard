# Docker Wizard Product Requirements Document (PRD)

## Document info
- Product: Docker Wizard
- Version: 1.0
- Date: 2026-02-26
- Status: Baseline PRD aligned to current implementation

## 1. Overview
Docker Wizard is a Go-based tool that scaffolds local Docker development environments. It detects the primary language of the project in the current directory, generates a matching `Dockerfile`, and generates a `docker-compose.yml` from selected supporting services.

The product supports both interactive and automated workflows through four run modes:
- `styled`: full TUI with styling and animations
- `plain`: TUI flow with plain text rendering
- `cli`: line-by-line interactive prompts
- `batch`: non-interactive, flag-driven generation

## 2. Problem statement
Developers repeatedly spend time building and maintaining Docker setup files across projects. This process is often inconsistent, error-prone, and hard to standardize across teams.

Key pain points:
- Repeated manual setup of common infra services (databases, queues, cache, proxies)
- Inconsistent Dockerfile patterns across language ecosystems
- Risk of overwriting existing local Docker configs
- Lack of deterministic output for clean code reviews

## 3. Goals
- Reduce time to first local containerized run.
- Generate sane defaults for app + optional infra services.
- Keep output deterministic and reproducible across runs.
- Preserve user work in existing files using merge + backup behavior.
- Support both interactive local usage and CI/bootstrap automation.

## 4. Non-goals
- Production deployment orchestration.
- Cloud infrastructure provisioning.
- Kubernetes or Helm generation.
- Deep app runtime tuning beyond scaffold defaults.

## 5. Target users
- Developers bootstrapping local development stacks.
- Teams standardizing local Docker conventions.
- Platform/DevEx engineers automating project setup.

## 6. User stories
- As a developer, I want language detection so I can avoid writing a Dockerfile from scratch.
- As a developer, I want to choose supporting services by category so I can build my stack quickly.
- As a user with existing files, I want safe merge behavior and backups so my changes are not lost.
- As an automation engineer, I want a batch mode so setup can run in scripts and CI.
- As a maintainer, I want catalog-driven services and templates so defaults can be updated without code changes.

## 7. Functional requirements

### 7.1 Language detection and Dockerfile generation
- Detect language in this priority order:
  1. Go: `go.mod`
  2. Node: `package.json`
  3. Python: `requirements.txt`, `pyproject.toml`, or `Pipfile`
  4. Ruby: `Gemfile`
  5. PHP: `composer.json` or `.php-version`
  6. Java: `pom.xml`, `build.gradle`, `build.gradle.kts`
  7. .NET: `*.csproj`
  8. Fallback: unknown
- Detect versions from ecosystem files where available.
- Generate Dockerfile templates for `go`, `node`, `python`, `ruby`, `php`, `java`, `dotnet`, and `unknown`.
- Allow language override in interactive and batch modes.

### 7.2 Service selection and compose generation
- Load services from `config/services.json`.
- Present selectable services by category:
  - database
  - message-queue
  - cache
  - analytics
  - proxy
- Always include an `app` service in compose output.
- Expand required services automatically based on `requires`.
- Keep `depends_on` references only for selected/expanded services.
- Publish host ports only for `public: true` services.
- Use `expose` for internal services.
- Include named volumes and network declarations when needed.

### 7.3 Validation and preview
- Provide warnings for:
  - dependency inconsistencies
  - host port collisions among public services
- Provide managed-file preview statuses:
  - `new`
  - `same`
  - `different`
  - `exists` (for `.dockerignore` when already present)

### 7.4 Write behavior and safety
- Create missing files.
- Mark identical existing files as unchanged.
- Merge differing existing files.
- Save backups of differing files as `*.bak`.
- Create `.dockerignore` only if missing.

### 7.5 Modes and interface
- Support run modes: `styled`, `plain`, `cli`, `batch`.
- In batch mode, support flags:
  - `--services`
  - `--language`
  - `--dry-run`
  - `--write`
- Default batch behavior should be preview-only unless `--write` is set.

## 8. Non-functional requirements
- Deterministic output ordering for stable diffs.
- Clear, concise user-facing output in all modes.
- Safe-by-default write flow with non-destructive merge behavior.
- Cross-terminal compatibility via styled/plain rendering options.
- Maintainability through modular generator packages and config-driven catalogs.

## 9. Success metrics
- Time to generate first valid scaffold (from command start).
- Percentage of reruns producing unchanged output for unchanged inputs.
- Percentage of runs completing without manual edits to generated files.
- Adoption of batch mode in scripts/automation.
- Reduction in onboarding/setup friction for new projects.

## 10. Constraints and assumptions
- Compose file version baseline is `3.9`.
- Default app port mapping is `8080:8080`.
- Service/template catalogs are the source of truth for defaults.
- Existing local files may contain user customizations and must be preserved.

## 11. Risks
- Merge behavior intentionally preserves existing content, which can retain outdated config.
- Warning-only validation (for example, port collisions) may still allow runtime failures.
- Catalog growth can increase maintenance complexity without governance.

## 12. Release and quality gates
- CI must pass formatting, vet, tests, and build checks.
- Successful pushes to `main` trigger auto-tag and release workflows.

## 13. Future opportunities
- Stronger fixture/snapshot testing across representative project types.
- Clear contributor docs for adding services and Dockerfile templates.
- More user customization options for ports and app start commands.
- Expanded install/distribution guidance across OS and shell environments.
