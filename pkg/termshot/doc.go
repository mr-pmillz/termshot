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

// Package termshot renders ANSI-styled terminal text as PNG images.
//
// It produces images that resemble a terminal window, complete with optional
// window decorations, shadow, and full ANSI color and text style support.
//
// # Basic usage
//
//	var buf bytes.Buffer
//	err := termshot.Render(&buf, strings.NewReader("\x1b[32mhello\x1b[0m"))
//
// For Word documents at 150 DPI with 7-inch content width:
//
//	err := termshot.Render(f, reader,
//	    termshot.WithColumns(80),
//	    termshot.WithTargetWidthInches(7.0, 150),
//	)
//
// # Recorder — explicit io.Writer capture
//
// [Recorder] is an [io.Writer] that buffers everything written to it, then
// renders the content as a PNG. It is goroutine-safe and composable with
// loggers, exec.Cmd.Stdout, or any other writer:
//
//	rec := termshot.NewRecorder(termshot.WithColumns(80)).Tee(os.Stdout)
//	defer rec.RenderToFile("output.png")
//	fmt.Fprintln(rec, "Hello, world!")
//
// # CaptureSession — automatic stdout/stderr capture
//
// [StartCapture] redirects os.Stdout and os.Stderr at the file-descriptor
// level so that all output (including from third-party libraries) is
// captured automatically. Output is tee'd to the original terminal:
//
//	capture, err := termshot.StartCapture("output.png", termshot.WithColumns(80))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer capture.Done()
//	fmt.Println("This is automatically captured!")
//
// CaptureSession is only supported on Unix systems (Linux and macOS).
package termshot
