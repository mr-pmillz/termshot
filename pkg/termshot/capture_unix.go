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

//go:build unix

package termshot

import (
	"fmt"
	"os"
	"syscall"
)

func startCapture(outputPath string, opts ...Option) (*CaptureSession, error) {
	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	// Dup the original stdout and stderr fds so we can restore them later
	// and tee output to the original terminal.
	savedOutFd, err := syscall.Dup(1) // #nosec G115
	if err != nil {
		_ = pipeR.Close()
		_ = pipeW.Close()
		return nil, fmt.Errorf("failed to dup stdout: %w", err)
	}
	savedOut := os.NewFile(uintptr(savedOutFd), "saved-stdout") // #nosec G115

	savedErrFd, err := syscall.Dup(2) // #nosec G115
	if err != nil {
		_ = pipeR.Close()
		_ = pipeW.Close()
		_ = savedOut.Close()
		return nil, fmt.Errorf("failed to dup stderr: %w", err)
	}
	savedErr := os.NewFile(uintptr(savedErrFd), "saved-stderr") // #nosec G115

	// Redirect fd 1 and fd 2 to the pipe write end.
	// os.Stdout and os.Stderr still wrap fds 1 and 2 respectively,
	// so all writes automatically go to the pipe now.
	if err := syscall.Dup2(int(pipeW.Fd()), 1); err != nil { // #nosec G115
		_ = pipeR.Close()
		_ = pipeW.Close()
		_ = savedOut.Close()
		_ = savedErr.Close()
		return nil, fmt.Errorf("failed to redirect stdout: %w", err)
	}
	if err := syscall.Dup2(int(pipeW.Fd()), 2); err != nil { // #nosec G115
		// Attempt to restore stdout before failing
		_ = syscall.Dup2(savedOutFd, 1) // #nosec G115
		_ = pipeR.Close()
		_ = pipeW.Close()
		_ = savedOut.Close()
		_ = savedErr.Close()
		return nil, fmt.Errorf("failed to redirect stderr: %w", err)
	}

	// Create recorder that tees to the saved (original) stdout
	rec := NewRecorder(opts...)
	rec.Tee(savedOut)

	c := &CaptureSession{
		rec:        rec,
		outputPath: outputPath,
		savedOut:   savedOut,
		savedErr:   savedErr,
		pipeR:      pipeR,
		pipeW:      pipeW,
		copyDone:   make(chan struct{}),
	}

	c.startCopyLoop()
	return c, nil
}

// dupFd restores a file descriptor by dup2'ing the saved fd onto targetFd.
func dupFd(saved *os.File, targetFd int) error {
	if err := syscall.Dup2(int(saved.Fd()), targetFd); err != nil { // #nosec G115
		return fmt.Errorf("failed to restore fd %d: %w", targetFd, err)
	}
	return nil
}
