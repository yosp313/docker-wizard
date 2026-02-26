# Docker Wizard Architecture Diagram

```mermaid
flowchart LR
    subgraph ENTRY [Entry and orchestration]
      MAIN["main.go parse args version output"]
      APP["internal/app/app.go dispatch by mode"]
    end

    subgraph MODES [Runtime interfaces]
      TUI["internal/tui/* styled and plain wizard"]
      CLII["internal/cli/interactive.go prompt driven flow"]
      CLIB["internal/cli/noninteractive.go batch automation flow"]
    end

    subgraph FACADE [Generator facade]
      GEN["internal/generator/generator.go stable API for UI and CLI"]
    end

    subgraph CORE [Generator core packages]
      CAT["internal/generator/catalog/* load and normalize service catalog"]
      DET["internal/generator/dockerfile/detect.go language and version detection"]
      DFT["internal/generator/dockerfile/dockerfile.go render Dockerfile templates"]
      CMP["internal/generator/compose/compose.go build deterministic compose"]
      VAL["internal/generator/validate/validate.go selection and collision warnings"]
      PRE["internal/generator/preview/preview.go new same different exists status"]
      WRT["internal/generator/write/write.go create merge backup files"]
    end

    subgraph CONFIG [Configuration sources]
      SVC["config/services.json"]
      DFC["config/dockerfiles.json"]
    end

    subgraph FILES [Project filesystem]
      ROOT["Current project root"]
      MARK["Language markers go.mod package.json pyproject.toml and others"]
      OUT1["docker-compose.yml"]
      OUT2["Dockerfile"]
      OUT3[".dockerignore"]
      BAK["*.bak backup files"]
    end

    MAIN --> APP
    APP --> TUI
    APP --> CLII
    APP --> CLIB

    TUI --> GEN
    CLII --> GEN
    CLIB --> GEN

    GEN --> CAT
    GEN --> DET
    GEN --> DFT
    GEN --> CMP
    GEN --> VAL
    GEN --> PRE
    GEN --> WRT

    CAT --> SVC
    DFT --> DFC
    DET --> MARK
    PRE --> ROOT

    CMP --> CAT
    VAL --> CAT
    VAL --> CMP

    WRT --> OUT1
    WRT --> OUT2
    WRT --> OUT3
    WRT --> BAK
```

## Module responsibilities
- `main.go`: parses CLI flags, validates option combinations, prints version, then calls app runtime.
- `internal/app/app.go`: central mode router (`styled`, `plain`, `cli`, `batch`) and runtime entrypoint.
- `internal/tui/*`: Bubble Tea state machine and views for the step-by-step wizard flow.
- `internal/tui/wizard.go`: step-specific key handlers keep transitions readable while preserving one state model.
- `internal/cli/interactive.go`: prompt-based interactive CLI path using the same generator APIs.
- `internal/cli/noninteractive.go`: batch mode orchestration for CI/script workflows (`--services`, `--language`, `--dry-run`, `--write`).
- `internal/generator/generator.go`: facade layer that exposes generator operations to UI/CLI callers.
- `internal/generator/catalog/*`: reads and validates service definitions from `config/services.json`.
- `internal/generator/dockerfile/*`: detects language/version from project files and renders templates from `config/dockerfiles.json`.
- `internal/generator/compose/compose.go`: builds deterministic compose output and expands required service dependencies.
- `internal/generator/validate/validate.go`: computes warning messages (dependency issues and host port collisions).
- `internal/generator/preview/preview.go`: computes pre-write file status (`new`, `same`, `different`, `exists`) using the same merge functions used by write.
- `internal/generator/write/write.go`: writes managed files, performs user-priority merge (existing values win), and creates `.bak` backups.
