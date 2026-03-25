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

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// isTmuxSession returns true if the current process is running inside a tmux session.
func isTmuxSession() bool {
	return os.Getenv("TMUX") != ""
}

// captureTmuxPane runs tmux capture-pane and returns the ANSI-colored content
// of the specified pane. If target is empty, the current pane is captured.
func captureTmuxPane(target string) ([]byte, error) {
	args := []string{"capture-pane", "-e", "-p"}
	if target != "" {
		args = append(args, "-t", target)
	}

	out, err := exec.Command("tmux", args...).Output() // #nosec G204
	if err != nil {
		return nil, fmt.Errorf("failed to capture tmux pane: %w", err)
	}

	return out, nil
}

// tmuxPaneWidth returns the column width of the current tmux pane.
func tmuxPaneWidth() (int, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_width}").Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get tmux pane width: %w", err)
	}

	width, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse tmux pane width: %w", err)
	}

	return width, nil
}
