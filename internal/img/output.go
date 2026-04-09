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

package img

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"strings"

	"github.com/esimov/stackblur-go"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/gonvenience/bunt"
	"github.com/gonvenience/font"
	"github.com/gonvenience/term"
	imgfont "golang.org/x/image/font"
)

const (
	red    = "#ED655A"
	yellow = "#E1C04C"
	green  = "#71BD47"
)

const (
	defaultFontSize = 12
	defaultFontDPI  = 144
)

// commandIndicator is the string to be used to indicate the command in the screenshot
var commandIndicator = func() string {
	if val, ok := os.LookupEnv("TS_COMMAND_INDICATOR"); ok {
		return val
	}

	return "➜"
}()

// FontFamily selects which embedded font to use for rendering.
type FontFamily int

const (
	// FontHack uses the default Hack font (embedded via gonvenience/font).
	FontHack FontFamily = iota
	// FontNerd uses the ZedMono Nerd Font for broader glyph/icon support.
	FontNerd
)

type Scaffold struct {
	content bunt.String

	factor float64

	columns int

	theme           Theme
	bgColorOverride *string
	fgColorOverride *color.Color

	clipCanvas bool

	drawDecorations bool
	drawShadow      bool

	shadowRadius  uint8
	shadowOffsetX float64
	shadowOffsetY float64

	padding float64
	margin  float64

	regular       imgfont.Face
	bold          imgfont.Face
	italic        imgfont.Face
	boldItalic    imgfont.Face
	emojiFallback *emojiFallbackFont

	tabSpaces int

	highlightCommand bool
	highlightColor   string
	highlightTight   bool
	commandRanges    []contentRange
}

func NewImageCreator() Scaffold {
	f := 2.0

	fontFaceOptions := &truetype.Options{
		Size: f * defaultFontSize,
		DPI:  defaultFontDPI,
	}

	return Scaffold{
		theme: DarkTheme,

		factor: f,

		margin:  f * 48,
		padding: f * 24,

		drawDecorations: true,
		drawShadow:      true,

		shadowRadius:  uint8(math.Min(f*16, 255)),
		shadowOffsetX: f * 16,
		shadowOffsetY: f * 16,

		regular:       font.Hack.Regular(fontFaceOptions),
		bold:          font.Hack.Bold(fontFaceOptions),
		italic:        font.Hack.Italic(fontFaceOptions),
		boldItalic:    font.Hack.BoldItalic(fontFaceOptions),
		emojiFallback: newEmojiFallback(f*defaultFontSize, defaultFontDPI),

		tabSpaces: 2,
	}
}

func (s *Scaffold) SetFontFaceRegular(face imgfont.Face) { s.regular = face }

func (s *Scaffold) SetFontFaceBold(face imgfont.Face) { s.bold = face }

func (s *Scaffold) SetFontFaceItalic(face imgfont.Face) { s.italic = face }

func (s *Scaffold) SetFontFaceBoldItalic(face imgfont.Face) { s.boldItalic = face }

// SetTheme sets the color theme for rendering.
func (s *Scaffold) SetTheme(t Theme) { s.theme = t }

// SetBackgroundColor overrides the window background color from the active theme.
func (s *Scaffold) SetBackgroundColor(hex string) { s.bgColorOverride = &hex }

// SetForegroundColorHex overrides the default text color from the active theme.
func (s *Scaffold) SetForegroundColorHex(hex string) {
	c, err := ParseHexColor(hex)
	if err != nil {
		return
	}
	s.fgColorOverride = &[]color.Color{c}[0]
}

// SetFont configures the font family used for rendering. Recalculates font
// faces using the current factor and DPI settings.
func (s *Scaffold) SetFont(family FontFamily) {
	fontFaceOptions := &truetype.Options{
		Size: s.factor * defaultFontSize,
		DPI:  defaultFontDPI,
	}

	switch family {
	case FontNerd:
		s.regular = zedMonoFace(zedMonoRegular, fontFaceOptions)
		s.bold = zedMonoFace(zedMonoBold, fontFaceOptions)
		s.italic = zedMonoFace(zedMonoItalic, fontFaceOptions)
		s.boldItalic = zedMonoFace(zedMonoBoldItalic, fontFaceOptions)
	default:
		s.regular = font.Hack.Regular(fontFaceOptions)
		s.bold = font.Hack.Bold(fontFaceOptions)
		s.italic = font.Hack.Italic(fontFaceOptions)
		s.boldItalic = font.Hack.BoldItalic(fontFaceOptions)
	}

	s.emojiFallback = newEmojiFallback(s.factor*defaultFontSize, defaultFontDPI)
}

func (s *Scaffold) effectiveBGColor() string {
	if s.bgColorOverride != nil {
		return *s.bgColorOverride
	}
	return s.theme.BackgroundColor
}

func (s *Scaffold) effectiveFGColor() color.Color {
	if s.fgColorOverride != nil {
		return *s.fgColorOverride
	}
	return s.theme.DefaultForegroundColor
}

func (s *Scaffold) effectiveShadowColor() string {
	return s.theme.ShadowBaseColor
}

func (s *Scaffold) effectiveBorderColor() string {
	return s.theme.BorderColor
}

func (s *Scaffold) SetColumns(columns int) { s.columns = columns }

func (s *Scaffold) SetMargin(margin float64) { s.margin = margin * s.factor }

func (s *Scaffold) SetPadding(padding float64) { s.padding = padding * s.factor }

func (s *Scaffold) DrawDecorations(value bool) { s.drawDecorations = value }

func (s *Scaffold) DrawShadow(value bool) { s.drawShadow = value }

func (s *Scaffold) ClipCanvas(value bool) { s.clipCanvas = value }

func (s *Scaffold) GetFixedColumns() int {
	if s.columns != 0 {
		return s.columns
	}

	columns, _ := term.GetTerminalSize()
	return columns
}

func (s *Scaffold) AddCommand(args ...string) error {
	before := len(s.content)
	err := s.AddContent(strings.NewReader(
		bunt.Sprintf("Lime{%s} DimGray{%s}\n",
			commandIndicator,
			strings.Join(args, " "),
		),
	))
	if err != nil {
		return err
	}

	s.commandRanges = append(s.commandRanges, contentRange{
		start: before,
		end:   len(s.content),
	})

	return nil
}

// HighlightCommand enables drawing a colored box around the command line(s).
// Use with AddCommand / --show-cmd. Default color is red (#FF0000).
func (s *Scaffold) HighlightCommand(enabled bool) { s.highlightCommand = enabled }

// SetHighlightColor overrides the highlight box color (default #FF0000).
func (s *Scaffold) SetHighlightColor(hex string) { s.highlightColor = hex }

// HighlightTight makes the highlight box fit tightly around the command text
// instead of spanning the full content width.
func (s *Scaffold) HighlightTight(enabled bool) { s.highlightTight = enabled }

func (s *Scaffold) AddContent(in io.Reader) error {
	parsed, err := bunt.ParseStream(in)
	if err != nil {
		return fmt.Errorf("failed to parse input stream: %w", err)
	}

	s.content = append(s.content, (*parsed)...)

	return nil
}

func (s *Scaffold) measureContent(layout contentLayout) (width float64, height float64) {
	cellWidth, cellHeight := s.cellSize()

	if s.columns > 0 {
		width = float64(s.GetFixedColumns()) * cellWidth
	} else {
		width = float64(layout.MaxCols) * cellWidth
	}

	height = float64(len(layout.Rows)) * cellHeight

	return width, height
}

func (s *Scaffold) image() (image.Image, error) {
	var f = func(value float64) float64 { return s.factor * value }

	var (
		corner   = f(6)
		radius   = f(9)
		distance = f(25)
	)

	layout := s.layout()
	contentWidth, contentHeight := s.measureContent(layout)

	// Make sure the output window is big enough in case no content or very few
	// content will be rendered
	contentWidth = math.Max(contentWidth, 3*distance+3*radius)

	marginX, marginY := s.margin, s.margin
	paddingX, paddingY := s.padding, s.padding

	xOffset := marginX
	yOffset := marginY

	var titleOffset float64
	if s.drawDecorations {
		titleOffset = f(40)
	}

	width := contentWidth + 2*marginX + 2*paddingX
	height := contentHeight + 2*marginY + 2*paddingY + titleOffset

	dc := gg.NewContext(int(width), int(height))

	// Optional: Apply blurred rounded rectangle to mimic the window shadow
	//
	if s.drawShadow {
		xOffset -= s.shadowOffsetX / 2
		yOffset -= s.shadowOffsetY / 2

		bc := gg.NewContext(int(width), int(height))
		bc.DrawRoundedRectangle(xOffset+s.shadowOffsetX, yOffset+s.shadowOffsetY, width-2*marginX, height-2*marginY, corner)
		bc.SetHexColor(s.effectiveShadowColor())
		bc.Fill()

		src := bc.Image()
		dst := image.NewNRGBA(src.Bounds())
		if err := stackblur.Process(dst, src, uint32(s.shadowRadius)); err != nil {
			return nil, err
		}

		dc.DrawImage(dst, 0, 0)
	}

	// Draw rounded rectangle with outline to produce impression of a window
	//
	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(s.effectiveBGColor())
	dc.Fill()

	dc.DrawRoundedRectangle(xOffset, yOffset, width-2*marginX, height-2*marginY, corner)
	dc.SetHexColor(s.effectiveBorderColor())
	dc.SetLineWidth(f(1))
	dc.Stroke()

	// Optional: Draw window decorations (i.e. three buttons) to produce the
	// impression of an actional window
	//
	if s.drawDecorations {
		for i, color := range []string{red, yellow, green} {
			dc.DrawCircle(xOffset+paddingX+float64(i)*distance+f(4), yOffset+paddingY+f(4), radius)
			dc.SetHexColor(color)
			dc.Fill()
		}
	}

	// Apply the actual text into the prepared content area of the window
	var xBase, yBase = xOffset + paddingX, yOffset + paddingY + titleOffset
	w, h := s.cellSize()
	for rowIndex, row := range layout.Rows {
		x := xBase
		y := yBase + float64(rowIndex)*h
		for _, segment := range row.Segments {
			if segment.Kind == segmentTab {
				x += w * float64(segment.Width)
				continue
			}

			segmentWidth := float64(segment.Width)

			switch segment.Settings & 0x02 { //nolint:gocritic
			case 2:
				dc.SetRGB255(
					int((segment.Settings>>32)&0xFF), // #nosec G115
					int((segment.Settings>>40)&0xFF), // #nosec G115
					int((segment.Settings>>48)&0xFF), // #nosec G115
				)

				dc.DrawRectangle(x, y, w*segmentWidth, h)
				dc.Fill()
			}

			switch segment.Settings & 0x01 {
			case 1:
				dc.SetRGB255(
					int((segment.Settings>>8)&0xFF),  // #nosec G115
					int((segment.Settings>>16)&0xFF), // #nosec G115
					int((segment.Settings>>24)&0xFF), // #nosec G115
				)

			default:
				dc.SetColor(s.effectiveFGColor())
			}

			text := renderText(segment.Text)
			if sprite, ok := emojiSprite(text); ok {
				drawWidth := int(w * segmentWidth)
				if drawWidth < 1 {
					drawWidth = 1
				}

				scaled := scaleImage(text, sprite, drawWidth, int(h))
				dc.DrawImage(scaled, int(x), int(y))
			} else {
				activeFace := s.textFace(segment.Settings, text)
				dc.SetFontFace(activeFace)
				ascent := float64(activeFace.Metrics().Ascent) / (1 << 6)
				dc.DrawString(text, x, y+ascent)
			}

			if segment.Settings&0x1C == 16 {
				dc.DrawLine(x, y+h, x+w*segmentWidth, y+h)
				dc.SetLineWidth(f(1))
				dc.Stroke()
			}

			x += w * segmentWidth
		}
	}

	// Optional: Draw a highlight box around the command line(s)
	if s.highlightCommand {
		commandRows, commandMaxCols := layout.commandBox()
		if commandRows > 0 {
			boxColor := s.highlightColor
			if boxColor == "" {
				boxColor = "#FF0000"
			}
			boxPad := f(4)
			boxX := xOffset + paddingX - boxPad
			boxY := yOffset + paddingY + titleOffset - boxPad
			boxW := contentWidth + 2*boxPad
			if s.highlightTight && commandMaxCols > 0 {
				boxW = float64(commandMaxCols)*w + 2*boxPad
			}
			boxH := float64(commandRows)*h + 2*boxPad

			dc.SetHexColor(boxColor)
			dc.SetLineWidth(f(2))
			dc.DrawRoundedRectangle(boxX, boxY, boxW, boxH, f(3))
			dc.Stroke()
		}
	}

	return dc.Image(), nil
}

func (layout contentLayout) commandBox() (rows int, maxCols int) {
	for _, row := range layout.Rows {
		if !row.HasCommand {
			if rows > 0 {
				break
			}
			continue
		}

		rows++
		if row.Width > maxCols {
			maxCols = row.Width
		}
	}

	return rows, maxCols
}

func renderText(text string) string {
	switch text {
	case "✗", "ˣ":
		return "×"
	default:
		return text
	}
}

func (s *Scaffold) textFace(settings uint64, text string) imgfont.Face {
	baseFace := s.regular
	switch settings & 0x1C {
	case 4:
		baseFace = s.bold
	case 8:
		baseFace = s.italic
	case 12:
		baseFace = s.boldItalic
	}

	if shouldUseEmojiFallback(text, s.emojiFallback) {
		return s.emojiFallback.face
	}

	return baseFace
}

func (s *Scaffold) cellSize() (float64, float64) {
	bounds, _, ok := s.regular.GlyphBounds('█')
	if !ok {
		panic("An internal font does not support the critical character █")
	}

	// Round down to force cells to have consistent pixel sizes
	w := float64((bounds.Max.X - bounds.Min.X).Floor())
	h := float64((bounds.Max.Y - bounds.Min.Y).Floor())
	return w, h
}

// Image renders and returns the scaffold content as an image.Image.
func (s *Scaffold) Image() (image.Image, error) {
	return s.image()
}

// Write writes the scaffold content as PNG into the provided writer
//
// Deprecated: Use [Scaffold.WritePNG] instead.
func (s *Scaffold) Write(w io.Writer) error {
	return s.WritePNG(w)
}

// WritePNG writes the scaffold content as PNG into the provided writer
func (s *Scaffold) WritePNG(w io.Writer) error {
	img, err := s.image()
	if err != nil {
		return err
	}

	// Optional: Clip image to minimum size by removing all surrounding transparent pixels
	//
	if s.clipCanvas {
		if imgRGBA, ok := img.(*image.RGBA); ok {
			var minX, minY = math.MaxInt, math.MaxInt
			var maxX, maxY = 0, 0

			var bounds = imgRGBA.Bounds()
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					r, g, b, a := imgRGBA.At(x, y).RGBA()
					isTransparent := r == 0 && g == 0 && b == 0 && a == 0

					if !isTransparent {
						if x < minX {
							minX = x
						}

						if y < minY {
							minY = y
						}

						if x > maxX {
							maxX = x
						}

						if y > maxY {
							maxY = y
						}
					}
				}
			}

			img = imgRGBA.SubImage(image.Rect(minX, minY, maxX, maxY))
		}
	}

	return png.Encode(w, img)
}

// ColumnsUsed returns the maximum rendered columns used across all rows in the
// current content. Call after AddContent.
func (s *Scaffold) ColumnsUsed() int {
	return s.layout().MaxCols
}

// WriteRaw writes the scaffold content as-is into the provided writer
func (s *Scaffold) WriteRaw(w io.Writer) error {
	_, err := w.Write([]byte(s.content.String()))
	return err
}
