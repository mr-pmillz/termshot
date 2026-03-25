# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is termshot

A Go CLI tool that executes terminal commands in a pseudo-terminal, captures output (including ANSI escape codes), and renders it as a styled PNG image resembling a macOS terminal window.

## Build & Test Commands

```bash
# Build
go build ./cmd/termshot/

# Run all tests (Ginkgo with race detection, randomization, coverage)
make test

# Run tests for a single package
go run github.com/onsi/ginkgo/v2/ginkgo run ./internal/img/
go run github.com/onsi/ginkgo/v2/ginkgo run ./internal/ptexec/

# Lint
golangci-lint run

# Clean
make clean
```

## Architecture

Three internal packages, each with a single responsibility:

- **`internal/cmd`** — Cobra CLI setup and orchestration. Parses flags, coordinates the pipeline: execute command -> capture output -> render image. `root_darwin.go` adds macOS-only `--clipboard` flag via build tags.
- **`internal/ptexec`** — Runs commands in a pseudo-terminal using `creack/pty`. Builder pattern API (`New().Cols(80).Command("ls").Run()`). Handles PTY sizing, `SIGWINCH` signals, and CI environment detection.
- **`internal/img`** — Renders ANSI-colored terminal output to PNG. `Scaffold` struct uses builder pattern for configuration. Pipeline: parse ANSI via `gonvenience/bunt` -> measure with font metrics -> draw shadow (`stackblur`) -> draw window frame + decorations (`fogleman/gg`) -> render each colored character with Hack font at 2x DPI.

**Data flow:** Command args -> ptexec (or `--raw-read` file) -> raw bytes with ANSI codes -> `Scaffold.AddContent()` parses into `bunt.ColoredRune[]` -> `Scaffold.WritePNG()` renders image.

## Testing

Uses **Ginkgo v2 + Gomega**. Tests are collocated with source files. Image tests in `internal/img` compare rendered PNG output byte-for-byte against reference images in `test/data/`. The custom `LookLike()` Gomega matcher handles this comparison.

## Key Dependencies

- `gonvenience/bunt` — ANSI escape code parsing into colored runes. Color and style data is bit-packed in `ColoredRune.Settings` (foreground RGB bits 8-24, background RGB bits 32-48, font style bits via `& 0x1C`).
- `fogleman/gg` — 2D graphics rendering for the terminal window image.
- `esimov/stackblur-go` — Gaussian blur for window shadow effect.
- `gonvenience/font` — Provides embedded Hack font (regular, bold, italic, bold-italic variants).

## Conventions

- Version is injected at build time via `-ldflags -X github.com/mr-pmillz/termshot/internal/cmd.version=...`
- `CGO_ENABLED=0` for static binaries
- The `TS_COMMAND_INDICATOR` env var overrides the default `➜` prompt indicator in screenshots
- GoReleaser handles cross-compilation (linux/darwin, amd64/arm64) and Homebrew tap publishing
- Linting uses golangci-lint with `gocritic` and `gosec` enabled
