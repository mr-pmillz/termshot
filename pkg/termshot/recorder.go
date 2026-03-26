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

package termshot

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Recorder captures written bytes and can render them as a termshot PNG.
// It implements io.Writer and is safe for concurrent use.
//
// Basic usage with defer:
//
//	func example() {
//	    rec := termshot.NewRecorder(termshot.WithColumns(80)).Tee(os.Stdout)
//	    defer rec.RenderToFile("output.png")
//	    fmt.Fprintln(rec, "Hello, world!")
//	}
//
// To handle the error from RenderToFile in a defer, use a named return:
//
//	func example() (err error) {
//	    rec := termshot.NewRecorder(termshot.WithColumns(80)).Tee(os.Stdout)
//	    defer func() { err = errors.Join(err, rec.RenderToFile("output.png")) }()
//	    fmt.Fprintln(rec, "Hello, world!")
//	    return nil
//	}
type Recorder struct {
	mu   sync.Mutex
	buf  bytes.Buffer
	tee  io.Writer
	opts []Option
}

// NewRecorder creates a Recorder that captures all written bytes for later
// rendering as a PNG. Options are passed through to [Render] when
// [Recorder.Render] or [Recorder.RenderToFile] is called. [WithQuiet] is
// applied automatically to suppress informational stderr output during
// rendering.
func NewRecorder(opts ...Option) *Recorder {
	return &Recorder{
		opts: append([]Option{WithQuiet()}, opts...),
	}
}

// Tee configures the Recorder to forward all writes to w in addition to
// buffering them. This is useful for keeping output visible in the terminal
// while recording. Tee returns the Recorder for chaining. It must be called
// before any writes.
func (r *Recorder) Tee(w io.Writer) *Recorder {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tee = w
	return r
}

// Write implements [io.Writer]. It appends p to the internal buffer and, if
// a tee writer is configured, also writes to it. Write is safe for concurrent
// use.
func (r *Recorder) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, err := r.buf.Write(p)
	if err != nil {
		return n, err
	}

	if r.tee != nil {
		// Write to tee but don't let tee errors lose buffered data.
		// The buffer write already succeeded.
		_, _ = r.tee.Write(p[:n])
	}

	return n, nil
}

// Bytes returns a copy of the captured bytes.
func (r *Recorder) Bytes() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]byte, r.buf.Len())
	copy(out, r.buf.Bytes())
	return out
}

// Render renders the captured content as a styled PNG image to w.
func (r *Recorder) Render(w io.Writer) error {
	data := r.Bytes()
	if len(data) == 0 {
		return fmt.Errorf("recorder has no content to render")
	}
	return Render(w, bytes.NewReader(data), r.opts...)
}

// RenderToFile renders the captured content as a PNG to the given file path.
// Only .png files are supported.
func (r *Recorder) RenderToFile(path string) error {
	if ext := filepath.Ext(path); ext != ".png" {
		return fmt.Errorf("file extension %q is not supported, only .png is supported", ext)
	}

	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return r.Render(f)
}

// Reset clears the captured buffer so the Recorder can be reused.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf.Reset()
}
