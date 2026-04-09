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
	"embed"
	"fmt"
	"image"
	"image/png"
	"strings"

	"golang.org/x/image/draw"
)

// Twemoji color emoji sprites (CC-BY 4.0).
// See internal/img/twemoji/LICENSE for license terms.
//
//go:embed twemoji/*.png
var twemojiFS embed.FS

// emojiSprite looks up a color emoji sprite for the given rune. Returns the
// decoded image and true if a sprite exists, or nil and false otherwise.
func emojiSprite(r rune) (image.Image, bool) {
	name := fmt.Sprintf("twemoji/%x.png", r)
	data, err := twemojiFS.ReadFile(name)
	if err != nil {
		return nil, false
	}
	img, err := png.Decode(strings.NewReader(string(data)))
	if err != nil {
		return nil, false
	}
	return img, true
}

// scaleImage resizes src to the given width and height using high-quality
// Catmull-Rom interpolation.
func scaleImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
