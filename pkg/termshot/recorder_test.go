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
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sync"

	"github.com/gonvenience/bunt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mr-pmillz/termshot/pkg/termshot"
)

var _ = Describe("Recorder", func() {
	BeforeEach(func() {
		bunt.SetColorSettings(bunt.ON, bunt.ON)
	})

	Context("basic write and render", func() {
		It("should capture written text and render a valid PNG", func() {
			rec := termshot.NewRecorder(termshot.WithColumns(80))
			_, _ = fmt.Fprintln(rec, "hello world")

			var buf bytes.Buffer
			Expect(rec.Render(&buf)).To(Succeed())
			Expect(buf.Len()).To(BeNumerically(">", 0))

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
			Expect(img.Bounds().Dy()).To(BeNumerically(">", 0))
		})

		It("should capture ANSI-colored text", func() {
			rec := termshot.NewRecorder(termshot.WithColumns(80))
			_, _ = fmt.Fprint(rec, "\x1b[32mgreen text\x1b[0m")

			var buf bytes.Buffer
			Expect(rec.Render(&buf)).To(Succeed())

			img, err := png.Decode(&buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should return error when rendering with no content", func() {
			rec := termshot.NewRecorder()
			var buf bytes.Buffer
			err := rec.Render(&buf)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no content"))
		})
	})

	Context("Tee", func() {
		It("should forward writes to the tee writer", func() {
			var teeBuf bytes.Buffer
			rec := termshot.NewRecorder(termshot.WithColumns(80)).Tee(&teeBuf)

			_, _ = fmt.Fprintln(rec, "hello tee")

			Expect(teeBuf.String()).To(Equal("hello tee\n"))

			// Should also be renderable
			var pngBuf bytes.Buffer
			Expect(rec.Render(&pngBuf)).To(Succeed())
			Expect(pngBuf.Len()).To(BeNumerically(">", 0))
		})
	})

	Context("Bytes", func() {
		It("should return a copy of captured bytes", func() {
			rec := termshot.NewRecorder()
			_, _ = fmt.Fprint(rec, "snapshot")

			data := rec.Bytes()
			Expect(string(data)).To(Equal("snapshot"))
		})
	})

	Context("Reset", func() {
		It("should clear the buffer so new content can be recorded", func() {
			rec := termshot.NewRecorder(termshot.WithColumns(80))
			_, _ = fmt.Fprint(rec, "first")
			rec.Reset()
			_, _ = fmt.Fprint(rec, "second")

			Expect(string(rec.Bytes())).To(Equal("second"))
		})
	})

	Context("RenderToFile", func() {
		It("should write a valid PNG to a file", func() {
			rec := termshot.NewRecorder(termshot.WithColumns(80))
			_, _ = fmt.Fprintln(rec, "file output")

			tmpDir := GinkgoT().TempDir()
			path := filepath.Join(tmpDir, "test.png")

			Expect(rec.RenderToFile(path)).To(Succeed())

			f, err := os.Open(path) // #nosec G304
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = f.Close() }()

			img, err := png.Decode(f)
			Expect(err).ToNot(HaveOccurred())
			Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		})

		It("should reject non-png file extension", func() {
			rec := termshot.NewRecorder()
			_, _ = fmt.Fprint(rec, "content")

			err := rec.RenderToFile("/tmp/test.jpg")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not supported"))
		})
	})

	Context("concurrent writes", func() {
		It("should be safe for concurrent use", func() {
			rec := termshot.NewRecorder(termshot.WithColumns(80))

			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(n int) {
					defer wg.Done()
					_, _ = fmt.Fprintf(rec, "goroutine %d\n", n)
				}(i)
			}
			wg.Wait()

			// All 10 goroutines should have written something
			Expect(len(rec.Bytes())).To(BeNumerically(">", 0))

			// Should still render successfully
			var buf bytes.Buffer
			Expect(rec.Render(&buf)).To(Succeed())
		})
	})

	Context("options passthrough", func() {
		It("should pass options through to Render", func() {
			narrowRec := termshot.NewRecorder(termshot.WithColumns(20))
			_, _ = fmt.Fprint(narrowRec, "abcdefghijklmnopqrstuvwxyz")
			var narrowBuf bytes.Buffer
			Expect(narrowRec.Render(&narrowBuf)).To(Succeed())

			wideRec := termshot.NewRecorder(termshot.WithColumns(80))
			_, _ = fmt.Fprint(wideRec, "abcdefghijklmnopqrstuvwxyz")
			var wideBuf bytes.Buffer
			Expect(wideRec.Render(&wideBuf)).To(Succeed())

			narrowImg, err := png.Decode(&narrowBuf)
			Expect(err).ToNot(HaveOccurred())
			wideImg, err := png.Decode(&wideBuf)
			Expect(err).ToNot(HaveOccurred())

			// Narrow columns should produce a narrower but taller image
			Expect(narrowImg.Bounds().Dx()).To(BeNumerically("<", wideImg.Bounds().Dx()))
		})

		It("should render with a highlighted command prepended to recorded content", func() {
			rec := termshot.NewRecorder(
				termshot.WithColumns(80),
				termshot.WithCommand("nmap", "-sV", "10.10.10.1"),
				termshot.WithHighlightCommand(true),
			)
			_, _ = fmt.Fprintln(rec, "22/tcp open ssh")

			var withCmdBuf bytes.Buffer
			Expect(rec.Render(&withCmdBuf)).To(Succeed())

			// Render the same content without command for comparison
			plainRec := termshot.NewRecorder(termshot.WithColumns(80))
			_, _ = fmt.Fprintln(plainRec, "22/tcp open ssh")

			var plainBuf bytes.Buffer
			Expect(plainRec.Render(&plainBuf)).To(Succeed())

			withCmdImg, err := png.Decode(&withCmdBuf)
			Expect(err).ToNot(HaveOccurred())
			plainImg, err := png.Decode(&plainBuf)
			Expect(err).ToNot(HaveOccurred())

			// With command prepended, image should be taller
			Expect(withCmdImg.Bounds().Dy()).To(BeNumerically(">", plainImg.Bounds().Dy()))
		})
	})
})
