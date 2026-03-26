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

package termshot

// Option configures the Render function.
type Option func(*config)

// WithColumns sets a fixed column count for text wrapping. When set, lines
// longer than this value are wrapped. When unset, the terminal width is
// detected automatically (defaulting to 80 in non-terminal environments
// such as CI/CD pipelines).
func WithColumns(n int) Option {
	return func(c *config) { c.columns = n }
}

// WithTargetWidth sets the desired output image width in pixels. The image
// is rendered at full internal resolution, then scaled to this width while
// preserving the aspect ratio.
func WithTargetWidth(pixels int) Option {
	return func(c *config) { c.targetWidth = pixels }
}

// WithTargetWidthInches computes the target width from physical dimensions.
// For example, WithTargetWidthInches(7.0, 150) produces an image 1050 pixels
// wide, suitable for a 7-inch column in a Word document printed at 150 DPI.
func WithTargetWidthInches(inches float64, dpi float64) Option {
	return func(c *config) { c.targetWidth = int(inches * dpi) }
}

// WithMargin sets the margin around the terminal window in logical pixels
// (default: 48).
func WithMargin(value int) Option {
	return func(c *config) { c.margin = &value }
}

// WithPadding sets the padding inside the terminal window in logical pixels
// (default: 24).
func WithPadding(value int) Option {
	return func(c *config) { c.padding = &value }
}

// WithDecorations controls whether window decorations (traffic light buttons)
// are drawn (default: true).
func WithDecorations(enabled bool) Option {
	return func(c *config) { c.decorations = &enabled }
}

// WithShadow controls whether the window shadow is drawn (default: true).
func WithShadow(enabled bool) Option {
	return func(c *config) { c.shadow = &enabled }
}

// WithClipCanvas removes transparent margins from the output image
// (default: false).
func WithClipCanvas(enabled bool) Option {
	return func(c *config) { c.clipCanvas = &enabled }
}

// WithCommand prepends a styled command prompt line to the output content.
func WithCommand(args ...string) Option {
	return func(c *config) { c.command = args }
}

// WithTmux captures the current tmux pane content. When set, the reader
// argument to Render is ignored and the content is read from the tmux pane.
// The pane width is automatically detected for column wrapping.
func WithTmux() Option {
	return func(c *config) { c.tmux = true }
}

// WithTmuxPane captures a specific tmux pane by target identifier
// (e.g. "%1", "session:window.pane"). Like WithTmux, the reader argument
// to Render is ignored.
func WithTmuxPane(target string) Option {
	return func(c *config) { c.tmux = true; c.tmuxPane = target }
}

// WithLightMode enables the light color theme (light background, dark text).
func WithLightMode() Option {
	return func(c *config) { c.light = true }
}

// WithBackgroundColor overrides the window background color. Value must be
// a CSS-style hex color (e.g. "#FFFFFF"). Takes precedence over theme.
func WithBackgroundColor(hex string) Option {
	return func(c *config) { c.bgColor = &hex }
}

// WithForegroundColor overrides the default text color. Value must be
// a CSS-style hex color (e.g. "#1E1E1E"). Takes precedence over theme.
func WithForegroundColor(hex string) Option {
	return func(c *config) { c.fgColor = &hex }
}

// WithNerdFont uses ZedMono Nerd Font for broader glyph and icon support
// instead of the default Hack font.
func WithNerdFont() Option {
	return func(c *config) { c.nerdFont = true }
}

// WithHighlightCommand draws a colored box around the command line(s) added
// via WithCommand. Useful for pentesting reports. Default color is red (#FF0000).
func WithHighlightCommand(enabled bool) Option {
	return func(c *config) { c.highlightCommand = &enabled }
}

// WithHighlightColor overrides the highlight box color (default #FF0000).
func WithHighlightColor(hex string) Option {
	return func(c *config) { c.highlightColor = &hex }
}

// WithHighlightTight fits the highlight box tightly around the command text
// instead of spanning the full content width.
func WithHighlightTight(enabled bool) Option {
	return func(c *config) { c.highlightTight = &enabled }
}

// WithQuiet suppresses informational messages (such as the column count)
// that are normally written to stderr during rendering.
func WithQuiet() Option {
	return func(c *config) { c.quiet = true }
}

// WithTmuxLines limits tmux capture to the last n lines from the pane's
// scrollback buffer. Implies WithTmux(). When n is 0 or negative, the
// entire scrollback history is captured. When unset, only the visible
// pane content is captured (the default tmux behavior).
func WithTmuxLines(n int) Option {
	return func(c *config) {
		c.tmux = true
		c.tmuxLines = &n
	}
}
