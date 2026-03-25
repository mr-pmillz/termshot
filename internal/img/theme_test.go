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
	"image/color"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/mr-pmillz/termshot/internal/img"
)

var _ = Describe("Theme and color parsing", func() {
	Context("ParseHexColor", func() {
		It("should parse a valid 6-digit hex color", func() {
			c, err := ParseHexColor("#FF8800")
			Expect(err).ToNot(HaveOccurred())
			Expect(c).To(Equal(color.RGBA{R: 0xFF, G: 0x88, B: 0x00, A: 0xFF}))
		})

		It("should parse a valid 8-digit hex color with alpha", func() {
			c, err := ParseHexColor("#10101066")
			Expect(err).ToNot(HaveOccurred())
			Expect(c).To(Equal(color.RGBA{R: 0x10, G: 0x10, B: 0x10, A: 0x66}))
		})

		It("should handle lowercase hex digits", func() {
			c, err := ParseHexColor("#aabbcc")
			Expect(err).ToNot(HaveOccurred())
			Expect(c).To(Equal(color.RGBA{R: 0xAA, G: 0xBB, B: 0xCC, A: 0xFF}))
		})

		It("should handle input without # prefix", func() {
			c, err := ParseHexColor("FF0000")
			Expect(err).ToNot(HaveOccurred())
			Expect(c).To(Equal(color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}))
		})

		It("should reject invalid length", func() {
			_, err := ParseHexColor("#FFF")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("must be 6 or 8 hex digits"))
		})

		It("should reject invalid hex characters", func() {
			_, err := ParseHexColor("#GGHHII")
			Expect(err).To(HaveOccurred())
		})
	})
})
