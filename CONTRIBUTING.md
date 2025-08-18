# Contributing

We welcome contributions of all kinds — bug fixes, new features, tests, and documentation improvements. Whether you're a seasoned Go developer or just getting started, we appreciate your interest in making `prof` better!

Before getting started, please check the [issues](https://github.com/AlexsanderHamir/prof/issues) to avoid duplicated effort or to find something to work on.

## 🏗️ Codebase Design

Before diving into development, we recommend familiarizing yourself with the project's architecture and design principles. This will help you understand how your contributions fit into the overall system.

**📖 [Codebase Design Documentation](docs/codebase_design.md)** - Comprehensive overview of the project's architecture, package structure, data flow, and design decisions.

**Key Design Principles:**

- **Layered Architecture**: Clean separation between CLI, engine, and internal packages
- **Single Responsibility**: Each package has a focused, well-defined purpose
- **Interface Contracts**: Stable APIs between major components
- **Configuration-Driven**: Flexible behavior through JSON configuration files

**Quick Architecture Overview:**

```
CLI (User Interface) → Engine (Business Logic) → Parser (Data Processing) → Internal (Utilities)
```

### 📁 Understanding the Codebase Structure

The project follows Go's standard project layout with clear package responsibilities:

- **`cmd/prof/`** - Application entry point and main function
- **`cli/`** - Command-line interface using Cobra framework
- **`engine/`** - Core business logic (benchmark, collector, tracker)
- **`parser/`** - Profile data parsing and processing
- **`internal/`** - Protected utilities (config, args, shared)
- **`tests/`** - Integration and end-to-end tests

**💡 Pro Tip**: Start with the [codebase design documentation](docs/codebase_design.md) to understand the relationships between these packages and how data flows through the system.

## 🔧 Quick Start

**Requirements:** Go 1.24.3+, Git

```bash
# Clone the repository
git clone https://github.com/AlexsanderHamir/prof.git
cd prof
go mod tidy

# Run tests
go test -v ./...

# Check for linter issues (first-time install if needed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# Build the binary for local testing
go build -o prof ./cmd/prof
```

## 📋 Contribution Guidelines

### 1. ✅ Run Tests and Linters Locally

Before submitting a pull request, make sure your changes pass:

- Unit tests:

  ```bash
  go test ./...
  ```

- Linting (using `golangci-lint`):

  ```bash
  golangci-lint run
  ```

GitHub Actions will run these checks again, but local checks save time and ensure faster feedback.

### 2. 🎯 Follow Code Style Guidelines

- Write idiomatic Go.
- Keep functions small and focused.
- Favor clarity over cleverness.
- Add comments for exported functions and any complex logic.

### 3. 📦 Commit Practices

- Use atomic commits — one logical change per commit.
- Start commit messages with a verb (e.g., `Fix`, `Add`, `Refactor`, `Document`).
- Avoid mixing unrelated changes like formatting, renaming, and new logic in the same commit.

### 4. 🧪 Add Tests

All non-trivial features and bug fixes should include tests that validate the behavior. If you're unsure how to test a change, open a draft PR or ask in the issue thread.

### 5. 📝 Document User-Facing Changes

If your change affects the:

- **CLI**
- **Configuration**
- **Output format**

…please update the corresponding documentation:

- `README.md`
- CLI help text (`--help`)
- Code comments or examples

### 6. 📬 Open a Pull Request

When your code is ready:

- Open a PR with a descriptive title and summary.
- Reference any relevant issue numbers (e.g., `Closes #12`).
- Mark the PR as a draft if it's not ready for review yet.

### 7. 💬 Collaborate Through Feedback

We review pull requests to ensure consistency, maintainability, and direction. Reviews are collaborative — don't hesitate to ask questions or propose alternatives. We're here to help you land the change.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
