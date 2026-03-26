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
	"errors"
	"io"
	"os"
)

// CaptureSession redirects os.Stdout and os.Stderr at the file-descriptor
// level, capturing all output (including from libraries that write directly to
// stdout/stderr) and rendering it as a PNG when [CaptureSession.Done] is
// called.
//
// Output is tee'd to the original stdout so it remains visible in the
// terminal during capture.
//
// CaptureSession is NOT safe for concurrent use. Other goroutines that write
// to os.Stdout or os.Stderr will also be captured. Always call Done via defer
// immediately after StartCapture:
//
//	func example() {
//	    capture, err := termshot.StartCapture("output.png", termshot.WithColumns(80))
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer capture.Done()
//
//	    fmt.Println("This is automatically captured!")
//	}
type CaptureSession struct {
	rec        *Recorder
	outputPath string

	savedOut *os.File // dup'd original stdout fd
	savedErr *os.File // dup'd original stderr fd
	pipeR    *os.File
	pipeW    *os.File
	copyDone chan struct{}
	copyErr  error
	done     bool
}

// StartCapture begins capturing all output written to os.Stdout and
// os.Stderr. The captured content will be rendered as a PNG to outputPath
// when [CaptureSession.Done] is called. Options are forwarded to [Render].
//
// This function is only supported on unix systems (Linux and macOS).
func StartCapture(outputPath string, opts ...Option) (*CaptureSession, error) {
	return startCapture(outputPath, opts...)
}

// Done restores the original stdout and stderr, waits for the copy goroutine
// to drain the pipe, and renders the captured output as a PNG to the path
// provided to [StartCapture]. It is safe to call Done multiple times; only
// the first call has any effect.
func (c *CaptureSession) Done() error {
	if c.done {
		return nil
	}
	c.done = true

	var errs []error

	// Restore file descriptors via dup2. os.Stdout and os.Stderr still
	// wrap fds 1 and 2, but after dup2 those fds point back to the
	// original terminal — no Go-level pointer swap needed.
	if err := dupFd(c.savedOut, 1); err != nil {
		errs = append(errs, err)
	}
	if err := dupFd(c.savedErr, 2); err != nil {
		errs = append(errs, err)
	}

	// Close pipe write end to signal EOF to the copy goroutine.
	// At this point fds 1 and 2 have been restored, so pipeW is the
	// last reference to the pipe write end.
	_ = c.pipeW.Close()

	// Wait for copy to finish
	<-c.copyDone
	_ = c.pipeR.Close()

	if c.copyErr != nil {
		errs = append(errs, c.copyErr)
	}

	// Clean up saved fds
	_ = c.savedOut.Close()
	_ = c.savedErr.Close()

	// Render captured content
	if err := c.rec.RenderToFile(c.outputPath); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Recorder returns the underlying Recorder, which can be used to inspect
// captured bytes or render to a different destination.
func (c *CaptureSession) Recorder() *Recorder {
	return c.rec
}

// startCopyLoop reads from pipeR and writes to the recorder. It runs in a
// goroutine and signals completion via the copyDone channel.
func (c *CaptureSession) startCopyLoop() {
	go func() {
		defer close(c.copyDone)
		_, c.copyErr = io.Copy(c.rec, c.pipeR)
	}()
}
