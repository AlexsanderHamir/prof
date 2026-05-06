<p align="center">
  <img src="assets/prof_logo_v0.0.1.png" alt="Prof" width="120" height="120" />
</p>

# Prof

Profiling eats time. Learning tools, running commands, hunting down output. Prof automates the whole process: one command collects everything and organizes it for analysis.

[![GoDoc](https://godoc.org/github.com/AlexsanderHamir/prof?status.svg)](https://godoc.org/github.com/AlexsanderHamir/prof)
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/github/v/tag/AlexsanderHamir/prof?sort=semver)](https://github.com/AlexsanderHamir/prof/releases)
![Go Version](https://img.shields.io/badge/Go-1.24.3%2B-blue)

[Documentation](https://alexsanderhamir.github.io/prof/)

## Start here

From your module root (`go.mod`):

```bash
prof ui
```

<p align="center">
  <img src="assets/prof_ui_example.gif" alt="prof ui: terminal and graphical UI" />
</p>

Menus first; **`prof auto`**, **`prof track`**, and flags are for scripts and CI. Examples, flags, and layout: **[documentation site](https://alexsanderhamir.github.io/prof/)** ([Quickstart](https://alexsanderhamir.github.io/prof/quickstart/), [Collect](https://alexsanderhamir.github.io/prof/collect/), [Compare](https://alexsanderhamir.github.io/prof/compare/)).

## Install

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

## Repo links

- [Contributing](CONTRIBUTING.md) · [Codebase design](CODEBASE_DESIGN.md) · [Code of conduct](CODE_OF_CONDUCT.md)

## Requirements

Go 1.24.3+, optional [Graphviz](https://graphviz.org/) for PNG call graphs, `go.mod` at project root.

## License

[MIT](LICENSE)
