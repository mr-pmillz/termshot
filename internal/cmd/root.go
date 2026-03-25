// Copyright © 2020 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/neat"

	"github.com/mr-pmillz/termshot/internal/img"
	"github.com/mr-pmillz/termshot/internal/ptexec"

	"github.com/spf13/cobra"
)

// version string will be injected by automation
var version string

// saveToClipboard function will be implemented by OS specific code
var saveToClipboard func(img.Scaffold) error

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s [%s flags] [--] command [command flags] [command arguments] [...]", executableName(), executableName()),
	Short: "Creates a screenshot of terminal command output",
	Long: `Executes the provided command as-is with all flags and arguments in a pseudo
terminal and captures the generated output. The result is printed as it was
produced. Additionally, an image will be rendered in a lookalike terminal
window including all terminal colors and text decorations.
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion, err := cmd.Flags().GetBool("version"); showVersion && err == nil {
			if len(version) == 0 {
				version = "(development)"
			}

			// #nosec G104
			// nolint:all
			bunt.Printf("Lime{*%s*} version DimGray{%s}\n",
				executableName(),
				version,
			)

			return nil
		}

		rawRead, _ := cmd.Flags().GetString("raw-read")
		rawWrite, _ := cmd.Flags().GetString("raw-write")
		tmuxCapture, _ := cmd.Flags().GetBool("tmux")
		tmuxPane, _ := cmd.Flags().GetString("tmux-pane")

		if tmuxCapture || tmuxPane != "" {
			if !isTmuxSession() {
				return fmt.Errorf("not inside a tmux session")
			}
		}

		if len(args) == 0 && rawRead == "" && !tmuxCapture && tmuxPane == "" {
			return cmd.Usage()
		}

		var scaffold = img.NewImageCreator()
		var buf bytes.Buffer
		var pt = ptexec.New()

		// Apply theme and color overrides (theme first, then individual overrides)
		if light, _ := cmd.Flags().GetBool("light"); light {
			scaffold.SetTheme(img.LightTheme)
		}
		if bgColor, _ := cmd.Flags().GetString("bg-color"); bgColor != "" {
			scaffold.SetBackgroundColor(bgColor)
		}
		if fgColor, _ := cmd.Flags().GetString("fg-color"); fgColor != "" {
			scaffold.SetForegroundColorHex(fgColor)
		}
		if nerdFont, _ := cmd.Flags().GetBool("nerd-font"); nerdFont {
			scaffold.SetFont(img.FontNerd)
		}
		if val, _ := cmd.Flags().GetBool("highlight-cmd"); val {
			scaffold.HighlightCommand(true)
		}
		if hex, _ := cmd.Flags().GetString("highlight-color"); hex != "" {
			scaffold.SetHighlightColor(hex)
		}
		if val, _ := cmd.Flags().GetBool("highlight-tight"); val {
			scaffold.HighlightTight(true)
		}

		// Initialise scaffold with a column sizing so that the
		// content can be wrapped accordingly
		//
		if columns, err := cmd.Flags().GetInt("columns"); err == nil && columns > 0 {
			scaffold.SetColumns(columns)
			pt.Cols(uint16(columns))
		}

		// Change margin if requested
		if margin, err := cmd.Flags().GetInt("margin"); err == nil {
			if margin < 0 {
				return fmt.Errorf("margin must be zero or greater: not %d", margin)
			}
			scaffold.SetMargin(float64(margin))
		}

		// Change padding if requested
		if padding, err := cmd.Flags().GetInt("padding"); err == nil {
			if padding < 0 {
				return fmt.Errorf("padding must be zero or greater: not %d", padding)
			}
			scaffold.SetPadding(float64(padding))
		}

		// Disable window shadow if requested
		//
		if val, err := cmd.Flags().GetBool("no-shadow"); err == nil {
			scaffold.DrawShadow(!val)
		}

		// Disable window decorations (buttons) if requested
		//
		if val, err := cmd.Flags().GetBool("no-decoration"); err == nil {
			scaffold.DrawDecorations(!val)
		}

		// Configure that canvas is clipped at the end
		//
		if val, err := cmd.Flags().GetBool("clip-canvas"); err == nil {
			scaffold.ClipCanvas(val)
		}

		// Optional: Prepend command line arguments to output content.
		// When combined with --raw-read, the args after -- are used only
		// as display text (the command is not re-executed).
		//
		if includeCommand, err := cmd.Flags().GetBool("show-cmd"); err == nil && includeCommand {
			if err := scaffold.AddCommand(args...); err != nil {
				return err
			}
		}

		// Get the actual content for the screenshot
		//
		if tmuxCapture || tmuxPane != "" {
			// Capture content from a tmux pane
			content, err := captureTmuxPane(tmuxPane)
			if err != nil {
				return err
			}
			buf.Write(content)

			// Auto-detect pane width for column wrapping if not set
			if columns, err := cmd.Flags().GetInt("columns"); err == nil && columns == 0 {
				if width, err := tmuxPaneWidth(); err == nil {
					scaffold.SetColumns(width)
				}
			}

		} else if rawRead == "" {
			// Run the provided command in a pseudo terminal and capture
			// the output to be later rendered into the screenshot
			bytes, err := pt.Command(args[0], args[1:]...).Run()
			if err != nil {
				return fmt.Errorf("failed to run command in pseudo terminal: %w", err)
			}
			buf.Write(bytes)

		} else {
			// Read the content from an existing file instead of
			// executing a command to read its output
			bytes, err := readFile(rawRead)
			if err != nil {
				return fmt.Errorf("failed to read contents: %w", err)
			}
			buf.Write(bytes)
		}

		// Allow manual override of command output content
		//
		if edit, err := cmd.Flags().GetBool("edit"); err == nil && edit && rawRead == "" {
			tmpFile, tmpErr := os.CreateTemp("", executableName())
			if tmpErr != nil {
				return tmpErr
			}

			defer func() { _ = os.Remove(tmpFile.Name()) }()

			if err := os.WriteFile(tmpFile.Name(), buf.Bytes(), os.FileMode(0644)); err != nil {
				return err
			}

			editor := os.Getenv("EDITOR")
			if len(editor) == 0 {
				editor = "vi"
			}

			if _, err := ptexec.New().Command(editor, tmpFile.Name()).Run(); err != nil {
				return err
			}

			bytes, tmpErr := os.ReadFile(tmpFile.Name())
			if tmpErr != nil {
				return tmpErr
			}

			buf.Reset()
			buf.Write(bytes)
		}

		// Add the captured output to the scaffold
		//
		if err := scaffold.AddContent(&buf); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Number of columns used: %d. Use '--columns' to impose it.\n", scaffold.ColumnsUsed())

		// Optional: Save content as-is to a file
		//
		if rawWrite != "" {
			var output *os.File
			var err error
			switch rawWrite {
			case "-":
				output = os.Stdout

			default:
				output, err = os.Create(filepath.Clean(rawWrite))
				if err != nil {
					return fmt.Errorf("failed to create file: %w", err)
				}

				defer func() { _ = output.Close() }()
			}

			return scaffold.WriteRaw(output)
		}

		// Optional: Save image to clipboard
		//
		if toClipboard, err := cmd.Flags().GetBool("clipboard"); err == nil && toClipboard {
			return saveToClipboard(scaffold)
		}

		// Save image to file
		//
		filename, err := cmd.Flags().GetString("filename")
		if filename == "" || err != nil {
			fmt.Fprintf(os.Stderr, "failed to read filename from command-line, defaulting to out.png")
			filename = "out.png"
		}

		if extension := filepath.Ext(filename); extension != ".png" {
			return fmt.Errorf("file extension %q of filename %q is not supported, only png is supported", extension, filename)
		}

		file, err := os.Create(filepath.Clean(filename))
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		defer func() { _ = file.Close() }()
		return scaffold.WritePNG(file)
	},
}

// Execute is the main entry point into the CLI code
func Execute() {
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, e error) error {
		return fmt.Errorf("unknown %s flag %w",
			executableName(),
			fmt.Errorf("issue with %v\n\nIn order to differentiate between program flags and command flags,\nuse '--' before the command so that all flags before the separator\nbelong to %s, while all others are used for the command.\n\n%s", e, executableName(), c.UsageString()),
		)
	})

	if err := rootCmd.Execute(); err != nil {
		var headline, content string

		type wrappedError interface {
			Error() string
			Unwrap() error
		}

		switch err := err.(type) {
		case wrappedError:
			headline = strings.SplitN(err.Error(), ":", 2)[0]
			content = err.Unwrap().Error()

		default:
			headline = "Error occurred"
			content = err.Error()
		}

		fmt.Fprint(os.Stderr, neat.ContentBox(
			headline,
			content,
			neat.HeadlineColor(bunt.OrangeRed),
			neat.ContentColor(bunt.LightCoral),
			neat.NoLineWrap(),
		))

		os.Exit(1)
	}
}

func executableName() string {
	if executable, err := os.Executable(); err == nil {
		return filepath.Clean(filepath.Base(executable))
	}

	return "termshot"
}

func readFile(name string) ([]byte, error) {
	switch name {
	case "-":
		return io.ReadAll(os.Stdin)

	default:
		return os.ReadFile(filepath.Clean(name))
	}
}

func init() {
	rootCmd.Flags().SortFlags = false

	// flags to control content
	rootCmd.Flags().BoolP("edit", "e", false, "edit content before creating screenshot")

	// flags to control look
	rootCmd.Flags().BoolP("show-cmd", "c", false, "include command in screenshot")
	rootCmd.Flags().IntP("columns", "C", 0, "force fixed number of columns in screenshot")
	rootCmd.Flags().IntP("margin", "m", 48, "set margin around the window")
	rootCmd.Flags().IntP("padding", "p", 24, "set padding around the content inside window")
	rootCmd.Flags().Bool("no-decoration", false, "do not draw window decorations")
	rootCmd.Flags().Bool("no-shadow", false, "do not draw window shadow")
	rootCmd.Flags().BoolP("clip-canvas", "s", false, "clip canvas to visible image area (no margin)")

	// flags to control theme and colors
	rootCmd.Flags().BoolP("light", "l", false, "use light color theme")
	rootCmd.Flags().String("bg-color", "", "override background color (hex, e.g. #FFFFFF)")
	rootCmd.Flags().String("fg-color", "", "override foreground/text color (hex, e.g. #333333)")
	rootCmd.Flags().Bool("nerd-font", false, "use ZedMono Nerd Font (broader glyph/icon support)")
	rootCmd.Flags().Bool("highlight-cmd", false, "draw a box around the command line (use with --show-cmd)")
	rootCmd.Flags().String("highlight-color", "", "override highlight box color (hex, default #FF0000)")
	rootCmd.Flags().Bool("highlight-tight", false, "fit the highlight box tightly around the command text")

	// flags for output related settings
	rootCmd.Flags().StringP("filename", "f", "out.png", "filename of the screenshot")

	// flags for raw output processing
	rootCmd.Flags().String("raw-write", "", "write raw output to file instead of creating a screenshot")
	rootCmd.Flags().String("raw-read", "", "read raw input from file instead of executing a command")

	// flags for tmux integration
	rootCmd.Flags().Bool("tmux", false, "capture the current tmux pane")
	rootCmd.Flags().String("tmux-pane", "", "capture a specific tmux pane by target (e.g. %1, session:window.pane)")

	// internals
	rootCmd.Flags().BoolP("version", "v", false, "show version")
}
