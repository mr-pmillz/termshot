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

package termshot_test

import (
	"bytes"
	"image/png"
	"os"
	"strings"

	"github.com/gonvenience/bunt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mr-pmillz/termshot/pkg/termshot"
)

var _ = Describe("Termshot Library", func() {
	BeforeEach(func() {
		bunt.SetColorSettings(bunt.ON, bunt.ON)
	})

	Context("Render with defaults", func() {
		It("should produce a valid PNG from plain text", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("hello world"))
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
			Expect(img.Bounds().Dy()).To(BeNumerically(">", 0))
		})

		It("should produce a valid PNG from ANSI text", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("\x1b[32mgreen text\x1b[0m"))
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should produce a valid PNG from multiline text", func() {
			var buf bytes.Buffer
			input := "line one\nline two\nline three"
			err := termshot.Render(&buf, strings.NewReader(input))
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
			Expect(img.Bounds().Dy()).To(BeNumerically(">", 0))
		})
	})

	Context("WithTargetWidth", func() {
		It("should scale the output to the specified pixel width", func() {
			var buf bytes.Buffer
			targetWidth := 672
			err := termshot.Render(&buf, strings.NewReader("hello world"),
				termshot.WithTargetWidth(targetWidth),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(Equal(targetWidth))
		})

		It("should preserve aspect ratio when scaling", func() {
			input := "line one\nline two\nline three"

			var origBuf bytes.Buffer
			err := termshot.Render(&origBuf, strings.NewReader(input),
				termshot.WithColumns(40),
			)
			Expect(err).ToNot(HaveOccurred())
			origImg, err := png.Decode(&origBuf)
			Expect(err).ToNot(HaveOccurred())
			origRatio := float64(origImg.Bounds().Dy()) / float64(origImg.Bounds().Dx())

			var scaledBuf bytes.Buffer
			targetWidth := 500
			err = termshot.Render(&scaledBuf, strings.NewReader(input),
				termshot.WithColumns(40),
				termshot.WithTargetWidth(targetWidth),
			)
			Expect(err).ToNot(HaveOccurred())
			scaledImg, err := png.Decode(&scaledBuf)
			Expect(err).ToNot(HaveOccurred())
			scaledRatio := float64(scaledImg.Bounds().Dy()) / float64(scaledImg.Bounds().Dx())

			Expect(scaledRatio).To(BeNumerically("~", origRatio, 0.02))
		})
	})

	Context("WithTargetWidthInches", func() {
		It("should compute correct pixel width for 96 DPI", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("hello"),
				termshot.WithTargetWidthInches(7.0, 96),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(Equal(672))
		})

		It("should compute correct pixel width for 150 DPI", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("hello"),
				termshot.WithTargetWidthInches(7.0, 150),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(Equal(1050))
		})
	})

	Context("WithColumns", func() {
		It("should wrap content at the specified column count", func() {
			var narrowBuf, wideBuf bytes.Buffer

			err := termshot.Render(&narrowBuf, strings.NewReader("abcdefghijklmnop"),
				termshot.WithColumns(8),
			)
			Expect(err).ToNot(HaveOccurred())

			err = termshot.Render(&wideBuf, strings.NewReader("abcdefghijklmnop"),
				termshot.WithColumns(80),
			)
			Expect(err).ToNot(HaveOccurred())

			narrowImg, err := png.Decode(&narrowBuf)
			Expect(err).ToNot(HaveOccurred())
			wideImg, err := png.Decode(&wideBuf)
			Expect(err).ToNot(HaveOccurred())

			Expect(narrowImg.Bounds().Dx()).To(BeNumerically("<", wideImg.Bounds().Dx()))
			Expect(narrowImg.Bounds().Dy()).To(BeNumerically(">", wideImg.Bounds().Dy()))
		})
	})

	Context("Options", func() {
		It("should render without decorations", func() {
			var withDeco, withoutDeco bytes.Buffer

			err := termshot.Render(&withDeco, strings.NewReader("test"))
			Expect(err).ToNot(HaveOccurred())

			err = termshot.Render(&withoutDeco, strings.NewReader("test"),
				termshot.WithDecorations(false),
			)
			Expect(err).ToNot(HaveOccurred())

			imgWith, err := png.Decode(&withDeco)
			Expect(err).ToNot(HaveOccurred())
			imgWithout, err := png.Decode(&withoutDeco)
			Expect(err).ToNot(HaveOccurred())
			Expect(imgWithout.Bounds().Dy()).To(BeNumerically("<", imgWith.Bounds().Dy()))
		})

		It("should render without shadow", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("test"),
				termshot.WithShadow(false),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should render with clip canvas", func() {
			var clippedBuf, unclippedBuf bytes.Buffer

			err := termshot.Render(&unclippedBuf, strings.NewReader("test"))
			Expect(err).ToNot(HaveOccurred())

			err = termshot.Render(&clippedBuf, strings.NewReader("test"),
				termshot.WithClipCanvas(true),
			)
			Expect(err).ToNot(HaveOccurred())

			clippedImg, err := png.Decode(&clippedBuf)
			Expect(err).ToNot(HaveOccurred())
			unclippedImg, err := png.Decode(&unclippedBuf)
			Expect(err).ToNot(HaveOccurred())

			// Clipped should be smaller or equal in both dimensions
			Expect(clippedImg.Bounds().Dx()).To(BeNumerically("<=", unclippedImg.Bounds().Dx()))
			Expect(clippedImg.Bounds().Dy()).To(BeNumerically("<=", unclippedImg.Bounds().Dy()))
		})

		It("should reject negative margin", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("test"),
				termshot.WithMargin(-5),
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("margin must be zero or greater"))
		})

		It("should reject negative padding", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("test"),
				termshot.WithPadding(-5),
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("padding must be zero or greater"))
		})

		It("should apply custom margin", func() {
			var defaultBuf, smallBuf bytes.Buffer

			err := termshot.Render(&defaultBuf, strings.NewReader("test"))
			Expect(err).ToNot(HaveOccurred())

			err = termshot.Render(&smallBuf, strings.NewReader("test"),
				termshot.WithMargin(0),
			)
			Expect(err).ToNot(HaveOccurred())

			defaultImg, err := png.Decode(&defaultBuf)
			Expect(err).ToNot(HaveOccurred())
			smallImg, err := png.Decode(&smallBuf)
			Expect(err).ToNot(HaveOccurred())

			Expect(smallImg.Bounds().Dx()).To(BeNumerically("<", defaultImg.Bounds().Dx()))
		})

		It("should apply custom padding", func() {
			var defaultBuf, largeBuf bytes.Buffer

			err := termshot.Render(&defaultBuf, strings.NewReader("test"))
			Expect(err).ToNot(HaveOccurred())

			err = termshot.Render(&largeBuf, strings.NewReader("test"),
				termshot.WithPadding(60),
			)
			Expect(err).ToNot(HaveOccurred())

			defaultImg, err := png.Decode(&defaultBuf)
			Expect(err).ToNot(HaveOccurred())
			largeImg, err := png.Decode(&largeBuf)
			Expect(err).ToNot(HaveOccurred())

			Expect(largeImg.Bounds().Dx()).To(BeNumerically(">", defaultImg.Bounds().Dx()))
		})

		It("should prepend command when WithCommand is used", func() {
			var withCmd, withoutCmd bytes.Buffer

			err := termshot.Render(&withoutCmd, strings.NewReader("output"))
			Expect(err).ToNot(HaveOccurred())

			err = termshot.Render(&withCmd, strings.NewReader("output"),
				termshot.WithCommand("echo", "hello"),
			)
			Expect(err).ToNot(HaveOccurred())

			imgWithout, err := png.Decode(&withoutCmd)
			Expect(err).ToNot(HaveOccurred())
			imgWith, err := png.Decode(&withCmd)
			Expect(err).ToNot(HaveOccurred())

			// With command prepended, image should be taller
			Expect(imgWith.Bounds().Dy()).To(BeNumerically(">", imgWithout.Bounds().Dy()))
		})
	})

	Context("WithTmux", func() {
		It("should fail when not in a tmux session", func() {
			// Temporarily unset TMUX to test the guard
			orig := os.Getenv("TMUX")
			Expect(os.Unsetenv("TMUX")).To(Succeed())
			defer func() { Expect(os.Setenv("TMUX", orig)).To(Succeed()) }()

			var buf bytes.Buffer
			err := termshot.Render(&buf, nil, termshot.WithTmux())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not inside a tmux session"))
		})

		It("should capture the current tmux pane when in tmux", func() {
			if os.Getenv("TMUX") == "" {
				Skip("not running inside tmux")
			}

			var buf bytes.Buffer
			err := termshot.Render(&buf, nil, termshot.WithTmux())
			Expect(err).ToNot(HaveOccurred())
			Expect(buf.Len()).To(BeNumerically(">", 0))

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should capture a specific tmux pane", func() {
			if os.Getenv("TMUX") == "" {
				Skip("not running inside tmux")
			}

			pane := os.Getenv("TMUX_PANE")
			if pane == "" {
				Skip("TMUX_PANE not set")
			}

			var buf bytes.Buffer
			err := termshot.Render(&buf, nil, termshot.WithTmuxPane(pane))
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should support target width with tmux capture", func() {
			if os.Getenv("TMUX") == "" {
				Skip("not running inside tmux")
			}

			var buf bytes.Buffer
			err := termshot.Render(&buf, nil,
				termshot.WithTmux(),
				termshot.WithTargetWidthInches(7.0, 96),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(Equal(672))
		})
	})

	Context("Theme and color options", func() {
		It("should render with light mode", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("light mode test"),
				termshot.WithLightMode(),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should render with custom background color", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("custom bg"),
				termshot.WithBackgroundColor("#FFFFFF"),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should render with custom foreground color", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("custom fg"),
				termshot.WithForegroundColor("#000000"),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should render with nerd font", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("nerd font test"),
				termshot.WithNerdFont(),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should combine light mode with custom colors", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("combined"),
				termshot.WithLightMode(),
				termshot.WithBackgroundColor("#FFFFF0"),
				termshot.WithForegroundColor("#333333"),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})
	})

	Context("Word document use case", func() {
		It("should render at 96 DPI for screen display in Word", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("CI pipeline output\nStep 1: Build\nStep 2: Test\nStep 3: Deploy"),
				termshot.WithColumns(80),
				termshot.WithTargetWidthInches(7.0, 96),
				termshot.WithDecorations(false),
				termshot.WithShadow(false),
				termshot.WithMargin(0),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(Equal(672))
		})

		It("should render at 150 DPI for print quality in Word", func() {
			var buf bytes.Buffer
			err := termshot.Render(&buf, strings.NewReader("CI pipeline output\nStep 1: Build\nStep 2: Test\nStep 3: Deploy"),
				termshot.WithColumns(80),
				termshot.WithTargetWidthInches(7.0, 150),
				termshot.WithDecorations(false),
				termshot.WithShadow(false),
				termshot.WithMargin(0),
			)
			Expect(err).ToNot(HaveOccurred())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(Equal(1050))
		})
	})
})
