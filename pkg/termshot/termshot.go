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

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mr-pmillz/termshot/internal/img"
	"golang.org/x/image/draw"
)

type config struct {
	columns     int
	margin      *int
	padding     *int
	decorations *bool
	shadow      *bool
	clipCanvas  *bool
	targetWidth int
	command     []string
	tmux        bool
	tmuxPane    string
}

// Render reads ANSI-styled terminal text from r and writes a styled PNG
// image to w. Use Option values to configure the output appearance and size.
func Render(w io.Writer, r io.Reader, opts ...Option) error {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}

	scaffold := img.NewImageCreator()

	if cfg.columns > 0 {
		scaffold.SetColumns(cfg.columns)
	}

	if cfg.margin != nil {
		if *cfg.margin < 0 {
			return fmt.Errorf("margin must be zero or greater: not %d", *cfg.margin)
		}
		scaffold.SetMargin(float64(*cfg.margin))
	}

	if cfg.padding != nil {
		if *cfg.padding < 0 {
			return fmt.Errorf("padding must be zero or greater: not %d", *cfg.padding)
		}
		scaffold.SetPadding(float64(*cfg.padding))
	}

	if cfg.decorations != nil {
		scaffold.DrawDecorations(*cfg.decorations)
	}

	if cfg.shadow != nil {
		scaffold.DrawShadow(*cfg.shadow)
	}

	if cfg.clipCanvas != nil {
		scaffold.ClipCanvas(*cfg.clipCanvas)
	}

	if len(cfg.command) > 0 {
		if err := scaffold.AddCommand(cfg.command...); err != nil {
			return fmt.Errorf("failed to add command: %w", err)
		}
	}

	// Determine content source: tmux pane or provided reader
	var content io.Reader
	if cfg.tmux {
		if os.Getenv("TMUX") == "" {
			return fmt.Errorf("not inside a tmux session")
		}

		data, err := captureTmux(cfg.tmuxPane)
		if err != nil {
			return err
		}
		content = bytes.NewReader(data)

		// Auto-detect pane width for column wrapping if not set
		if cfg.columns == 0 {
			if width, err := paneWidth(); err == nil {
				scaffold.SetColumns(width)
			}
		}
	} else {
		content = r
	}

	if err := scaffold.AddContent(content); err != nil {
		return fmt.Errorf("failed to add content: %w", err)
	}

	if cfg.targetWidth > 0 {
		return renderScaled(w, &scaffold, cfg.targetWidth)
	}

	return scaffold.WritePNG(w)
}

func captureTmux(target string) ([]byte, error) {
	args := []string{"capture-pane", "-e", "-p"}
	if target != "" {
		args = append(args, "-t", target)
	}

	out, err := exec.Command("tmux", args...).Output() // #nosec G204
	if err != nil {
		return nil, fmt.Errorf("failed to capture tmux pane: %w", err)
	}

	return out, nil
}

func paneWidth() (int, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_width}").Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get tmux pane width: %w", err)
	}

	width, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse tmux pane width: %w", err)
	}

	return width, nil
}

func renderScaled(w io.Writer, scaffold *img.Scaffold, targetWidth int) error {
	src, err := scaffold.Image()
	if err != nil {
		return fmt.Errorf("failed to render image: %w", err)
	}

	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth == 0 {
		return fmt.Errorf("rendered image has zero width")
	}

	targetHeight := int(float64(srcHeight) * float64(targetWidth) / float64(srcWidth))
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	return png.Encode(w, dst)
}
