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
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/gonvenience/bunt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mr-pmillz/termshot/pkg/termshot"
)

var _ = Describe("CaptureSession", Serial, func() {
	BeforeEach(func() {
		bunt.SetColorSettings(bunt.ON, bunt.ON)
	})

	It("should capture fmt.Println output and render a valid PNG", func() {
		tmpDir := GinkgoT().TempDir()
		path := filepath.Join(tmpDir, "capture.png")

		capture, err := termshot.StartCapture(path, termshot.WithColumns(80))
		Expect(err).ToNot(HaveOccurred())

		fmt.Println("captured via StartCapture")

		Expect(capture.Done()).To(Succeed())

		// Verify the PNG file was created
		f, err := os.Open(path) // #nosec G304
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = f.Close() }()

		img, err := png.Decode(f)
		Expect(err).ToNot(HaveOccurred())
		Expect(img.Bounds().Dx()).To(BeNumerically(">", 0))
		Expect(img.Bounds().Dy()).To(BeNumerically(">", 0))
	})

	It("should capture stderr output", func() {
		tmpDir := GinkgoT().TempDir()
		path := filepath.Join(tmpDir, "stderr.png")

		capture, err := termshot.StartCapture(path, termshot.WithColumns(80))
		Expect(err).ToNot(HaveOccurred())

		fmt.Fprintf(os.Stderr, "stderr output\n")

		Expect(capture.Done()).To(Succeed())

		// Verify captured content includes stderr
		data := capture.Recorder().Bytes()
		Expect(string(data)).To(ContainSubstring("stderr output"))
	})

	It("should restore stdout and stderr after Done", func() {
		tmpDir := GinkgoT().TempDir()
		path := filepath.Join(tmpDir, "restore.png")

		capture, err := termshot.StartCapture(path, termshot.WithColumns(80))
		Expect(err).ToNot(HaveOccurred())

		_, _ = fmt.Println("during capture")

		Expect(capture.Done()).To(Succeed())

		// After Done, writing to stdout should work normally (not hang or error).
		// If fds weren't restored, this would write to the closed pipe.
		n, err := fmt.Fprintln(os.Stdout, "after restore")
		Expect(err).ToNot(HaveOccurred())
		Expect(n).To(BeNumerically(">", 0))
	})

	It("should be safe to call Done multiple times", func() {
		tmpDir := GinkgoT().TempDir()
		path := filepath.Join(tmpDir, "double.png")

		capture, err := termshot.StartCapture(path, termshot.WithColumns(80))
		Expect(err).ToNot(HaveOccurred())

		fmt.Println("once")

		Expect(capture.Done()).To(Succeed())
		Expect(capture.Done()).To(Succeed()) // second call is a no-op
	})

	It("should provide access to the underlying Recorder", func() {
		tmpDir := GinkgoT().TempDir()
		path := filepath.Join(tmpDir, "recorder.png")

		capture, err := termshot.StartCapture(path, termshot.WithColumns(80))
		Expect(err).ToNot(HaveOccurred())

		fmt.Println("recorder access test")

		rec := capture.Recorder()
		Expect(rec).ToNot(BeNil())

		Expect(capture.Done()).To(Succeed())

		data := rec.Bytes()
		Expect(string(data)).To(ContainSubstring("recorder access test"))
	})
})
