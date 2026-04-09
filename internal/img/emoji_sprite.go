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
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/png"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/draw"
)

// Twemoji color emoji sprites (CC-BY 4.0).
// See internal/img/twemoji/LICENSE for license terms.
//
//go:embed twemoji/*.png
var twemojiFS embed.FS

type emojiSpriteEntry struct {
	image image.Image
	ok    bool
}

var (
	emojiSpriteCache sync.Map
	emojiScaleCache  sync.Map
)

// emojiSprite looks up a color emoji sprite for the given grapheme cluster. Returns the
// decoded image and true if a sprite exists, or nil and false otherwise.
func emojiSprite(text string) (image.Image, bool) {
	if !isLikelyEmojiCluster(text) {
		return nil, false
	}

	for _, key := range emojiSpriteKeys(text) {
		if sprite, ok := loadEmojiSprite(key); ok {
			return sprite, true
		}
	}

	return nil, false
}

func loadEmojiSprite(key string) (image.Image, bool) {
	if cached, ok := emojiSpriteCache.Load(key); ok {
		entry := cached.(emojiSpriteEntry)
		return entry.image, entry.ok
	}

	data, err := twemojiFS.ReadFile(fmt.Sprintf("twemoji/%s.png", key))
	if err != nil {
		emojiSpriteCache.Store(key, emojiSpriteEntry{ok: false})
		return nil, false
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		emojiSpriteCache.Store(key, emojiSpriteEntry{ok: false})
		return nil, false
	}

	emojiSpriteCache.Store(key, emojiSpriteEntry{image: img, ok: true})
	return img, true
}

func emojiSpriteKeys(text string) []string {
	if text == "" {
		return nil
	}

	exact := emojiSpriteKey(text, false)
	withoutVS16 := emojiSpriteKey(text, true)
	if exact == withoutVS16 {
		return []string{exact}
	}

	return []string{exact, withoutVS16}
}

func emojiSpriteKey(text string, stripVS16 bool) string {
	parts := make([]string, 0, len(text))
	for _, r := range text {
		if stripVS16 && r == '\uFE0F' {
			continue
		}

		parts = append(parts, strconv.FormatInt(int64(r), 16))
	}

	return strings.Join(parts, "-")
}

// scaleImage resizes src to the given width and height using high-quality
// Catmull-Rom interpolation.
func scaleImage(key string, src image.Image, width, height int) image.Image {
	cacheKey := fmt.Sprintf("%s/%dx%d", key, width, height)
	if cached, ok := emojiScaleCache.Load(cacheKey); ok {
		return cached.(image.Image)
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	emojiScaleCache.Store(cacheKey, dst)
	return dst
}
