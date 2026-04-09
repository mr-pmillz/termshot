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
	"strings"
	"testing"
)

func TestEmojiSpriteSupportsSequences(t *testing.T) {
	t.Parallel()

	samples := []string{"🇺🇸", "👍🏽", "👩‍🚀", "1️⃣", "❤️"}
	for _, sample := range samples {
		if _, ok := emojiSprite(sample); !ok {
			t.Fatalf("expected color emoji sprite for %q", sample)
		}
	}
}

func TestEmojiFallbackIgnoresPlainASCII(t *testing.T) {
	t.Parallel()

	fallback := newEmojiFallback(24, 144)
	if shouldUseEmojiFallback("1", fallback) {
		t.Fatal("expected ASCII digit to use primary font")
	}
	if shouldUseEmojiFallback("#", fallback) {
		t.Fatal("expected ASCII punctuation to use primary font")
	}
	if !shouldUseEmojiFallback("👍", fallback) {
		t.Fatal("expected emoji rune to use emoji fallback")
	}
}

func TestColumnsUsedCountsGraphemeWidth(t *testing.T) {
	t.Parallel()

	scaffold := NewImageCreator()
	if err := scaffold.AddContent(strings.NewReader("👍🏽")); err != nil {
		t.Fatalf("AddContent failed: %v", err)
	}

	if got := scaffold.ColumnsUsed(); got != 2 {
		t.Fatalf("ColumnsUsed() = %d, want 2", got)
	}
}

func TestLimitRenderedRowsCountsWrappedRows(t *testing.T) {
	t.Parallel()

	scaffold := NewImageCreator()
	scaffold.SetColumns(20)
	if err := scaffold.AddContent(strings.NewReader(strings.Repeat("x", 200))); err != nil {
		t.Fatalf("AddContent failed: %v", err)
	}

	before := scaffold.layout()
	if len(before.Rows) <= 3 {
		t.Fatalf("expected wrapped content to span more than 3 rows, got %d", len(before.Rows))
	}

	scaffold.LimitRenderedRows(3)
	after := scaffold.layout()
	if len(after.Rows) != 3 {
		t.Fatalf("expected 3 rendered rows after limiting, got %d", len(after.Rows))
	}
}

func TestLimitRenderedRowsExcludesCommandRows(t *testing.T) {
	t.Parallel()

	scaffold := NewImageCreator()
	scaffold.SetColumns(20)
	if err := scaffold.AddCommand("echo", "hi"); err != nil {
		t.Fatalf("AddCommand failed: %v", err)
	}
	if err := scaffold.AddContent(strings.NewReader(strings.Repeat("x", 200))); err != nil {
		t.Fatalf("AddContent failed: %v", err)
	}

	scaffold.LimitRenderedRows(2)
	layout := scaffold.layout()
	commandRows, _ := layout.commandBox()
	if commandRows != 1 {
		t.Fatalf("command rows = %d, want 1", commandRows)
	}
	if len(layout.Rows) != commandRows+2 {
		t.Fatalf("expected command row plus 2 limited content rows, got %d rows", len(layout.Rows))
	}
}
