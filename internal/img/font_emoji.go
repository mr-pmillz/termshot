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
	_ "embed"
	"sync"

	imgfont "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

// Noto Emoji (SIL Open Font License) — monochrome emoji fallback font.
// Uses opentype (not freetype/truetype) because freetype's cmap parser
// cannot resolve supplementary-plane codepoints (U+10000+) like emoji.
// See internal/img/fonts/NotoEmoji/OFL.txt for license terms.

//go:embed fonts/NotoEmoji/NotoEmoji.ttf
var notoEmojiRegular []byte

var (
	notoEmojiOnce sync.Once
	notoEmojiFont *sfnt.Font
	notoEmojiErr  error
)

// emojiFallbackFont holds the parsed sfnt.Font for glyph-index lookups and
// the rendered font.Face for drawing. We need both because freetype/truetype
// faces always return ok=true from GlyphBounds (even for missing glyphs),
// so we use sfnt.GlyphIndex to reliably detect emoji coverage.
type emojiFallbackFont struct {
	font *sfnt.Font
	face imgfont.Face
}

// hasGlyph reports whether the emoji font contains a real glyph for r.
func (e *emojiFallbackFont) hasGlyph(r rune) bool {
	if e == nil || e.font == nil {
		return false
	}
	var buf sfnt.Buffer
	idx, err := e.font.GlyphIndex(&buf, r)
	return err == nil && idx != 0
}

func loadNotoEmojiFont() (*sfnt.Font, error) {
	notoEmojiOnce.Do(func() {
		notoEmojiFont, notoEmojiErr = opentype.Parse(notoEmojiRegular)
	})

	return notoEmojiFont, notoEmojiErr
}

func newEmojiFallback(size, dpi float64) *emojiFallbackFont {
	f, err := loadNotoEmojiFont()
	if err != nil {
		panic("failed to parse embedded Noto Emoji font: " + err.Error())
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size: size,
		DPI:  dpi,
	})
	if err != nil {
		panic("failed to create Noto Emoji font face: " + err.Error())
	}
	return &emojiFallbackFont{font: f, face: face}
}

func shouldUseEmojiFallback(text string, fallback *emojiFallbackFont) bool {
	if fallback == nil {
		return false
	}

	runes := []rune(text)
	if len(runes) != 1 {
		return false
	}

	return isEmojiRune(runes[0]) && fallback.hasGlyph(runes[0])
}

func isLikelyEmojiCluster(text string) bool {
	if text == "" {
		return false
	}

	hasEmojiBase := false
	hasKeycapBase := false
	for _, r := range text {
		switch {
		case isEmojiRune(r):
			hasEmojiBase = true
		case isKeycapBase(r):
			hasKeycapBase = true
		case r == '\u200D' || r == '\uFE0F' || r == '\u20E3':
			if r == '\u20E3' && !hasKeycapBase {
				return false
			}
			if r == '\u20E3' {
				hasEmojiBase = true
			}
		case r >= 0x1F3FB && r <= 0x1F3FF:
		default:
			return false
		}
	}

	return hasEmojiBase
}

func isKeycapBase(r rune) bool {
	return (r >= '0' && r <= '9') || r == '#' || r == '*'
}

func isEmojiRune(r rune) bool {
	switch {
	case r == 0x00A9 || r == 0x00AE:
		return true
	case r >= 0x203C && r <= 0x3299:
		return true
	case r >= 0x1F000 && r <= 0x1FAFF:
		return true
	default:
		return false
	}
}
