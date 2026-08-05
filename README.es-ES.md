

<p align="center">
  <img src="assets/prof_logo_v0.0.1.png" alt="Prof" width="120" height="120" />
</p>

# Prof

El profiling consume tiempo. Aprender herramientas, ejecutar comandos, buscar la salida. Prof automatiza todo el proceso: un solo comando recopila todo y lo organiza para su análisis.

[![GoDoc](https://godoc.org/github.com/AlexsanderHamir/prof?status.svg)](https://godoc.org/github.com/AlexsanderHamir/prof)
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/github/v/tag/AlexsanderHamir/prof?sort=semver)](https://github.com/AlexsanderHamir/prof/releases)
![Go Version](https://img.shields.io/badge/Go-1.24.3%2B-blue)

[Sitio de documentación](https://alexsanderhamir.github.io/prof/)

## Comienza aquí

Desde la raíz de tu módulo (`go.mod`):

```bash
prof ui
```

<p align="center">
  <img src="assets/prof_ui_example.gif" alt="prof ui: interfaz de terminal y gráfica" />
</p>

## Instalación

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

## Enlaces del repositorio

- [Contribuir](CONTRIBUTING.md) · [Diseño de la base de código](CODEBASE_DESIGN.md) · [Código de conducta](CODE_OF_CONDUCT.md)

## Requisitos

Go 1.24.3+, [Graphviz](https://graphviz.org/) opcional para gráficos de llamadas en PNG, `go.mod` en la raíz del proyecto.

## Licencia

[MIT](LICENSE)
