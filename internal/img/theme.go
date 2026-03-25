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
	"image/color"
	"strings"

	"github.com/gonvenience/bunt"
)

// Theme holds the configurable colors for the terminal window rendering.
type Theme struct {
	BackgroundColor        string      // window fill hex color
	BorderColor            string      // window border stroke hex color
	ShadowBaseColor        string      // shadow hex color (with alpha, e.g. #10101066)
	DefaultForegroundColor color.Color // fallback text color
}

// DarkTheme is the default dark terminal theme matching the original hardcoded values.
var DarkTheme = Theme{
	BackgroundColor:        "#151515",
	BorderColor:            "#404040",
	ShadowBaseColor:        "#10101066",
	DefaultForegroundColor: bunt.LightGray,
}

// LightTheme provides a light terminal appearance with dark text on a light background.
var LightTheme = Theme{
	BackgroundColor:        "#F5F5F5",
	BorderColor:            "#C0C0C0",
	ShadowBaseColor:        "#00000022",
	DefaultForegroundColor: color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF},
}

// ParseHexColor converts a CSS-style hex color string (#RRGGBB or #RRGGBBAA)
// to a color.RGBA value.
func ParseHexColor(hex string) (color.RGBA, error) {
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b, a uint8
	a = 0xFF

	switch len(hex) {
	case 6:
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("invalid hex color %q: %w", "#"+hex, err)
		}

	case 8:
		_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
		if err != nil {
			return color.RGBA{}, fmt.Errorf("invalid hex color %q: %w", "#"+hex, err)
		}

	default:
		return color.RGBA{}, fmt.Errorf("invalid hex color %q: must be 6 or 8 hex digits", "#"+hex)
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}
