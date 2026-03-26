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

### Light mode and custom colors

Use `--light` for a light-themed screenshot, or override individual colors:

```sh
# Light mode
termshot --light -- ls -la

# Custom background and foreground
termshot --bg-color="#002B36" --fg-color="#839496" -- ls -la

# Nerd Font for icon/glyph support
termshot --nerd-font -- ls -la
```

### Pentesting: highlight the command with a red box

Use `--highlight-cmd` with `--show-cmd` to draw a red box around the command line. This is standard practice in penetration testing reports:

```sh
# Live capture with highlighted command
termshot --light -c --highlight-cmd -- nmap -sV 10.10.10.1

# After-the-fact: screenshot saved output without re-running the command
nmap -sV 10.10.10.1 > /tmp/nmap-output.txt 2>&1
termshot --light -c --highlight-cmd \
  --raw-read /tmp/nmap-output.txt \
  -f nmap-screenshot.png \
  -- nmap -sV 10.10.10.1
```

In the after-the-fact workflow, `--raw-read` provides the content and the args after `--` are used only as display text — the command is **not** re-executed.

Use `--highlight-tight` to fit the box snugly around the command text instead of spanning the full width:

```sh
# Tight box — ends where the command text ends
termshot --light -c --highlight-cmd --highlight-tight -- nmap -sV 10.10.10.1

# Full-width box (default)
termshot --light -c --highlight-cmd -- nmap -sV 10.10.10.1
```

Use `--highlight-color` to override the box color:

```sh
termshot -c --highlight-cmd --highlight-color="#FFA500" -- nuclei -t cves/ -u target
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

### Theme & Colors

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--light` | `-l` | `false` | Light color theme |
| `--bg-color` | | | Override background color (hex, e.g. `#FFFFFF`) |
| `--fg-color` | | | Override foreground/text color (hex, e.g. `#333333`) |
| `--nerd-font` | | `false` | Use ZedMono Nerd Font (broader glyph/icon support) |
| `--highlight-cmd` | | `false` | Draw a box around the command (use with `-c`) |
| `--highlight-tight` | | `false` | Fit the highlight box tightly around the command text |
| `--highlight-color` | | `#FF0000` | Override highlight box color |

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
| `--tmux-lines` | | `0` | Capture last N lines from tmux scrollback (0 = visible pane only) |

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

### Run a command and screenshot with highlighted prompt

`RenderCommand` runs a command, captures its output, and renders a screenshot in one call. The command is executed once via `sh -c` and is never re-run. This is the easiest way to integrate termshot into Go programs like pentesting tools or CI/CD pipelines:

```go
package main

import (
    "os"

    "github.com/mr-pmillz/termshot/pkg/termshot"
)

func main() {
    f, _ := os.Create("nmap-screenshot.png")
    defer f.Close()

    termshot.RenderCommand(f, "nmap -sV 10.10.10.1",
        termshot.WithLightMode(),
        termshot.WithHighlightCommand(true),  // red box around command
        termshot.WithHighlightTight(true),    // box fits the command text
        termshot.WithColumns(120),
    )
}
```

### Screenshot saved output after the fact

When you've already captured command output (e.g. from a pipeline), use `Render` with `WithCommand` to add the command display text without re-executing it:

```go
// output was captured earlier, command is NOT re-run
output, _ := os.ReadFile("/tmp/nuclei-results.txt")

f, _ := os.Create("nuclei-screenshot.png")
defer f.Close()

termshot.Render(f, bytes.NewReader(output),
    termshot.WithCommand("nuclei", "-t", "cves/", "-u", "target"),
    termshot.WithHighlightCommand(true),
    termshot.WithLightMode(),
    termshot.WithColumns(120),
)
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

### Recorder — capture output via io.Writer

`Recorder` is an `io.Writer` that buffers everything written to it and renders a PNG on demand. Use `Tee` to keep output visible in the terminal while recording. This is useful for logging pipelines, test harnesses, or any code that produces output through an `io.Writer`:

```go
package main

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/mr-pmillz/termshot/pkg/termshot"
)

func main() {
    rec := termshot.NewRecorder(
        termshot.WithColumns(120),
        termshot.WithLightMode(),
    ).Tee(os.Stdout) // also print to terminal

    // Route command output through the recorder
    cmd := exec.Command("go", "test", "./...")
    cmd.Stdout = rec
    cmd.Stderr = rec
    _ = cmd.Run()

    // Render everything that was written
    _ = rec.RenderToFile("test-results.png")
}
```

`Recorder` is goroutine-safe, so multiple writers can use it concurrently. Use `Reset()` to clear the buffer and reuse it.

### CaptureSession — automatic stdout/stderr capture with defer

`StartCapture` redirects `os.Stdout` and `os.Stderr` at the file-descriptor level, so **all** output is captured automatically — including output from third-party libraries, child processes, and `fmt.Println`. Output is tee'd to the original terminal so it remains visible during capture. Call `Done` via `defer` to restore the original file descriptors and render the PNG:

```go
package main

import (
    "fmt"
    "os"

    "github.com/mr-pmillz/termshot/pkg/termshot"
)

func main() {
    capture, err := termshot.StartCapture("output.png",
        termshot.WithColumns(80),
        termshot.WithLightMode(),
    )
    if err != nil {
        panic(err)
    }
    defer capture.Done()

    fmt.Println("This is automatically captured!")
    fmt.Fprintf(os.Stderr, "Errors are captured too.\n")
    // When main returns, Done() renders everything to output.png
}
```

`CaptureSession` is supported on Linux and macOS (uses `dup2` syscall). It is **not** goroutine-safe — other goroutines writing to stdout/stderr will also be captured. For goroutine-safe capture, use `Recorder` instead.

You can also use `CaptureSession` inside any function with `defer` to screenshot just that function's output:

```go
func runScan(target string) error {
    capture, err := termshot.StartCapture(
        fmt.Sprintf("scan-%s.png", target),
        termshot.WithColumns(120),
        termshot.WithHighlightCommand(true),
    )
    if err != nil {
        return err
    }
    defer capture.Done()

    // Everything printed in this function is captured
    cmd := exec.Command("nmap", "-sV", target)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

### Capture tmux scrollback

Use `WithTmuxLines(n)` to capture the last N lines from a tmux pane's scrollback buffer instead of just the visible area:

```go
f, _ := os.Create("scrollback.png")
defer f.Close()

termshot.Render(f, nil,
    termshot.WithTmuxLines(50),   // last 50 lines (implies WithTmux)
    termshot.WithColumns(120),
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
| `WithHighlightCommand(bool)` | Draw a colored box around the command |
| `WithHighlightTight(bool)` | Fit the highlight box tightly around the command text |
| `WithHighlightColor(hex)` | Override highlight box color (default: `#FF0000`) |
| `WithLightMode()` | Light color theme |
| `WithBackgroundColor(hex)` | Override background color |
| `WithForegroundColor(hex)` | Override text color |
| `WithNerdFont()` | Use ZedMono Nerd Font |
| `WithTmux()` | Capture current tmux pane (reader is ignored) |
| `WithTmuxPane(target)` | Capture specific tmux pane |
| `WithTmuxLines(n)` | Capture last N lines from tmux scrollback (implies `WithTmux`) |
| `WithQuiet()` | Suppress informational stderr messages during rendering |

> _Note:_ This project is work in progress. Although a lot of ANSI sequences can be parsed, there are commands that create output which cannot be parsed correctly yet. Commands that reset the cursor position are known to create issues.
