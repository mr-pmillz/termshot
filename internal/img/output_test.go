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

package img_test

import (
	"bytes"
	"image/png"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
	. "github.com/mr-pmillz/termshot/internal/img"
)

var _ = Describe("Creating images", func() {
	Context("Use scaffold to create PNG file", func() {
		BeforeEach(func() {
			SetColorSettings(ON, ON)
		})

		It("should write a PNG stream based on provided input", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-foobar.png")))
		})

		It("should omit the window decorations when configured", func() {
			scaffold := NewImageCreator()
			scaffold.DrawDecorations(false)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-no-decoration.png")))
		})

		It("should omit the window shadow when configured", func() {
			scaffold := NewImageCreator()
			scaffold.DrawShadow(false)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-no-shadow.png")))
		})

		It("should clip the canvas when configured", func() {
			scaffold := NewImageCreator()
			scaffold.ClipCanvas(true)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-clip-canvas.png")))
		})

		It("should wrap the content when configured", func() {
			scaffold := NewImageCreator()
			scaffold.SetColumns(4)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-wrapping.png")))
		})

		It("should show the command when configured", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddCommand("echo", "foobar")).To(Succeed())
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).To(Succeed())
			Expect(scaffold).To(LookLike(testdata("expected-show-cmd.png")))
		})

		It("should apply margin correctly", func() {
			scaffold := NewImageCreator()
			scaffold.SetMargin(24)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-margin.png")))
		})

		It("should apply padding correctly", func() {
			scaffold := NewImageCreator()
			scaffold.SetPadding(60)

			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-padding.png")))
		})

		It("should line up block elements to cells", func() {
			baseString := `
				fffffffffff f f f f
				ffffffffff b b b b b
				fffffffffff f f f f
				ffffffffff b b b b b
				fffffffffff f f f f
				ffffffffff b b b b b
				fffffffffff f f f f
				ffffffffff b b b b b
			`

			// Allow nicer indentation for baseString
			baseString = strings.ReplaceAll(baseString, "\t", "")
			baseString = strings.TrimSpace(baseString)

			baseString = strings.ReplaceAll(baseString, "f", "█")
			baseString = strings.ReplaceAll(baseString, "b", "\x1b[31;107m┼\x1b[0m")

			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(bytes.NewBufferString(baseString))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-cells.png")))
		})

		It("should write a PNG stream based on provided input with ANSI sequences", func() {
			var buf bytes.Buffer
			_, _ = Fprintf(&buf, "Text with emphasis, like *bold*, _italic_, _*bold/italic*_ or ~underline~.\n\n")
			_, _ = Fprintf(&buf, "Colors:\n")
			_, _ = Fprintf(&buf, "\tRed{Red}\n")
			_, _ = Fprintf(&buf, "\tGreen{Green}\n")
			_, _ = Fprintf(&buf, "\tBlue{Blue}\n")
			_, _ = Fprintf(&buf, "\tMintCream{MintCream}\n")

			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(&buf)).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-ansi.png")))
		})
	})

	Context("Use scaffold with light theme", func() {
		BeforeEach(func() {
			SetColorSettings(ON, ON)
		})

		It("should write a PNG stream with light theme", func() {
			scaffold := NewImageCreator()
			scaffold.SetTheme(LightTheme)
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-foobar-light.png")))
		})

		It("should write a PNG stream with light theme and ANSI sequences", func() {
			var buf bytes.Buffer
			_, _ = Fprintf(&buf, "Text with emphasis, like *bold*, _italic_, _*bold/italic*_ or ~underline~.\n\n")
			_, _ = Fprintf(&buf, "Colors:\n")
			_, _ = Fprintf(&buf, "\tRed{Red}\n")
			_, _ = Fprintf(&buf, "\tGreen{Green}\n")
			_, _ = Fprintf(&buf, "\tBlue{Blue}\n")
			_, _ = Fprintf(&buf, "\tMintCream{MintCream}\n")

			scaffold := NewImageCreator()
			scaffold.SetTheme(LightTheme)
			Expect(scaffold.AddContent(&buf)).ToNot(HaveOccurred())
			Expect(scaffold).To(LookLike(testdata("expected-ansi-light.png")))
		})
	})

	Context("Use scaffold with custom color overrides", func() {
		BeforeEach(func() {
			SetColorSettings(ON, ON)
		})

		It("should render with a custom background color", func() {
			scaffold := NewImageCreator()
			scaffold.SetBackgroundColor("#FF0000")
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())

			var buf bytes.Buffer
			Expect(scaffold.WritePNG(&buf)).To(Succeed())
			Expect(buf.Len()).To(BeNumerically(">", 0))

			_, err := png.Decode(bytes.NewReader(buf.Bytes()))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should render with a custom foreground color", func() {
			scaffold := NewImageCreator()
			scaffold.SetForegroundColorHex("#00FF00")
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())

			var buf bytes.Buffer
			Expect(scaffold.WritePNG(&buf)).To(Succeed())
			Expect(buf.Len()).To(BeNumerically(">", 0))

			_, err := png.Decode(bytes.NewReader(buf.Bytes()))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Column reporting", func() {
		BeforeEach(func() {
			SetColorSettings(ON, ON)
		})

		It("should report the correct number of columns used", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())
			Expect(scaffold.ColumnsUsed()).To(Equal(6))
		})

		It("should report max columns across multiple lines", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader("short\nlonger line here\nmed"))).ToNot(HaveOccurred())
			Expect(scaffold.ColumnsUsed()).To(Equal(16))
		})

		It("should return zero for empty content", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.ColumnsUsed()).To(Equal(0))
		})
	})

	Context("Use scaffold with nerd font", func() {
		BeforeEach(func() {
			SetColorSettings(ON, ON)
		})

		It("should render a valid PNG with nerd font", func() {
			scaffold := NewImageCreator()
			scaffold.SetFont(FontNerd)
			Expect(scaffold.AddContent(strings.NewReader("foobar"))).ToNot(HaveOccurred())

			var buf bytes.Buffer
			Expect(scaffold.WritePNG(&buf)).To(Succeed())
			Expect(buf.Len()).To(BeNumerically(">", 0))

			_, err := png.Decode(bytes.NewReader(buf.Bytes()))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Use scaffold to create raw output file", func() {
		var buf bytes.Buffer

		BeforeEach(func() {
			SetColorSettings(ON, ON)
			buf.Reset()
		})

		It("should write an output file with the content as-is", func() {
			scaffold := NewImageCreator()
			Expect(scaffold.AddContent(strings.NewReader(Sprintf("MintCream{foobar}")))).To(Succeed())
			Expect(scaffold.WriteRaw(&buf)).To(Succeed())
			Expect(buf.String()).To(Equal("\x1b[38;2;245;255;250mfoobar\x1b[0m"))
		})
	})
})
