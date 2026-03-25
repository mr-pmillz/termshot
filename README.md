# termshot

[![License](https://img.shields.io/github/license/mr-pmillz/termshot.svg)](https://github.com/mr-pmillz/termshot/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mr-pmillz/termshot)](https://goreportcard.com/report/github.com/mr-pmillz/termshot)
[![Tests](https://github.com/mr-pmillz/termshot/workflows/Tests/badge.svg)](https://github.com/mr-pmillz/termshot/actions?query=workflow%3A%22Tests%22)
[![Go Reference](https://pkg.go.dev/badge/github.com/mr-pmillz/termshot.svg)](https://pkg.go.dev/github.com/mr-pmillz/termshot)
[![Release](https://img.shields.io/github/release/mr-pmillz/termshot.svg)](https://github.com/mr-pmillz/termshot/releases/latest)

Generate beautiful screenshots of your terminal, from your terminal. Supports running commands, capturing existing output, tmux pane capture, and use as a Go library.

## Installation

```sh
go install github.com/mr-pmillz/termshot/cmd/termshot@latest
```

Pre-compiled binaries for Darwin and Linux are available on the [Releases](https://github.com/mr-pmillz/termshot/releases/) page.

## CLI Usage

### Screenshot a command

Prefix any command with `termshot` to capture its output as a PNG:

```sh
termshot ls -a
```

This generates `out.png` in the current directory.

Use `--` to separate termshot flags from command flags, and wrap piped commands in quotes:

```sh
termshot --show-cmd -- "ls -1 | grep go"
```

### Screenshot existing output

Use `--raw-read` to screenshot content from a file or stdin without running a command. This is useful in CI/CD pipelines where the output has already been generated:

```sh
# From a file
termshot --raw-read build-output.log -f build-screenshot.png

# From stdin
cat output.txt | termshot --raw-read - -f screenshot.png

# Pipe any command's output
my-build-script 2>&1 | termshot --raw-read - --columns 120 -f ci-output.png
```

### Screenshot a tmux pane

When running inside tmux, use `--tmux` to capture the current pane with its colors and formatting intact:

```sh
# Capture the current pane
termshot --tmux -f pane-screenshot.png

# Capture a specific pane by target
termshot --tmux-pane %1 -f other-pane.png

# Combine with styling options
termshot --tmux --no-decoration --no-shadow -f clean-capture.png
```

The pane width is auto-detected for correct column wrapping. You can override it with `--columns`.

You can also pipe tmux capture output manually:

```sh
tmux capture-pane -e -p | termshot --raw-read - -f screenshot.png
tmux capture-pane -e -p -t %2 | termshot --raw-read - --columns 120 -f pane2.png
```

## CLI Flags

### Appearance

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--show-cmd` | `-c` | `false` | Include the command in the screenshot |
| `--columns` | `-C` | auto | Fixed column count for line wrapping |
| `--margin` | `-m` | `48` | Space around the window |
| `--padding` | `-p` | `24` | Space inside the window |
| `--no-decoration` | | `false` | Hide window buttons |
| `--no-shadow` | | `false` | Hide window shadow |
| `--clip-canvas` | `-s` | `false` | Remove transparent margins |

### Output

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--filename` | `-f` | `out.png` | Output file path |
| `--clipboard` | `-b` | `false` | Copy to clipboard (macOS only) |

### Input

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--raw-read` | | | Read from file or stdin (`-`) instead of running a command |
| `--raw-write` | | | Write raw text to file instead of creating a screenshot |
| `--tmux` | | `false` | Capture the current tmux pane |
| `--tmux-pane` | | | Capture a specific tmux pane by target (e.g. `%1`) |

### Other

| Flag | Short | Description |
|------|-------|-------------|
| `--edit` | `-e` | Edit content in `$EDITOR` before generating screenshot |
| `--version` | `-v` | Print version |

## Go Library

The `pkg/termshot` package lets you render terminal screenshots from Go code. This is useful for generating documentation images in CI/CD pipelines, test reports, or any workflow where you need to convert ANSI terminal output to PNG.

```sh
go get github.com/mr-pmillz/termshot
```

### Basic usage

```go
package main

import (
    "os"
    "strings"

    "github.com/mr-pmillz/termshot/pkg/termshot"
)

func main() {
    f, _ := os.Create("screenshot.png")
    defer f.Close()

    input := strings.NewReader("\x1b[32m$ go test ./...\x1b[0m\nok  mypackage 0.42s")
    termshot.Render(f, input, termshot.WithColumns(80))
}
```

### Sized for Word documents

Generate images that fit a 7-inch content width in Microsoft Word:

```go
f, _ := os.Create("report-output.png")
defer f.Close()

termshot.Render(f, reader,
    termshot.WithColumns(80),
    termshot.WithTargetWidthInches(7.0, 150),  // 1050px at 150 DPI
    termshot.WithDecorations(false),
    termshot.WithShadow(false),
    termshot.WithMargin(0),
)
```

### Capture a tmux pane

```go
f, _ := os.Create("tmux-capture.png")
defer f.Close()

termshot.Render(f, nil,
    termshot.WithTmux(),
    termshot.WithTargetWidth(1200),
    termshot.WithDecorations(false),
)
```

### Library options

| Option | Description |
|--------|-------------|
| `WithColumns(n)` | Fixed column count for line wrapping |
| `WithTargetWidth(pixels)` | Scale output to exact pixel width |
| `WithTargetWidthInches(inches, dpi)` | Scale to physical size (e.g. `7.0, 150`) |
| `WithMargin(n)` | Margin around window (default: 48) |
| `WithPadding(n)` | Padding inside window (default: 24) |
| `WithDecorations(bool)` | Window buttons (default: true) |
| `WithShadow(bool)` | Window shadow (default: true) |
| `WithClipCanvas(bool)` | Remove transparent edges (default: false) |
| `WithCommand(args...)` | Prepend styled command prompt |
| `WithTmux()` | Capture current tmux pane (reader is ignored) |
| `WithTmuxPane(target)` | Capture specific tmux pane |

> _Note:_ This project is work in progress. Although a lot of ANSI sequences can be parsed, there are commands that create output which cannot be parsed correctly yet. Commands that reset the cursor position are known to create issues.
