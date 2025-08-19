# Contributing

Thanks for your interest in improving **`prof`**! We welcome contributions of all kinds — bug fixes, features, tests, and documentation.

👉 Before starting, check [issues](https://github.com/AlexsanderHamir/prof/issues) to avoid duplication or to pick something up.

## 🏗️ Codebase Overview

- **Architecture:** Layered design with clear separation:

  ```
  CLI → Engine → Parser → Internal
  ```

- **Key principles:**

  - Single responsibility per package
  - Stable interfaces between components
  - Config-driven behavior (JSON)

- **Structure:**

  - `cmd/prof/` – Entry point
  - `cli/` – Cobra-based CLI

  - `engine/` – Core logic

    - `benchmark/` – Running and managing benchmarks
    - `collector/` – Gathering profiling data
    - `tracker/` – Tracking runs, comparisons, and state

  - `parser/` – Profile data parsing and processing

  - `internal/` – Shared utilities

    - `config/` – Config file handling
    - `args/` – Argument parsing
    - `shared/` – Common helpers

  - `tests/` – Integration and E2E tests

📖 For details, see [codebase design docs](CODEBASE_DESIGN.md).

## 🔧 Quick Start

**Requirements:** Go 1.24.3+, Git

```bash
git clone https://github.com/AlexsanderHamir/prof.git
cd prof
go mod tidy

# Run tests
go test ./...

# Lint (install first if needed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# Build binary
go build -o prof ./cmd/prof
```

## 📋 Contribution Guidelines

1. **✅ Tests & Linting**
   Run before pushing:

   ```bash
   go test ./...
   golangci-lint run
   ```

2. **🎯 Code Style**

   - Idiomatic Go
   - Small, focused functions
   - Clear > clever
   - Comments for exported functions & tricky logic

3. **📦 Commits**

   - One logical change per commit
   - Use verbs (`Add`, `Fix`, `Refactor`, `Docs`)
   - Don’t mix unrelated changes

4. **🧪 Tests**

   - Required for non-trivial changes
   - Unsure? Open a draft PR for feedback

5. **📝 Documentation**
   Update when changes affect:

   - CLI
   - Config
   - Output

6. **📬 Pull Requests**

   - Descriptive title & summary
   - Reference issues (e.g., `Closes #12`)
   - Draft if not ready

7. **💬 Reviews**
   Feedback is collaborative — ask questions, suggest alternatives, and we’ll help guide contributions.
