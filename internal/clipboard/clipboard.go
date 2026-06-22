// Package clipboard provides cross-platform clipboard read/write by shelling
// out to the native clipboard utilities. No cgo, no third-party deps.
package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// ErrNoClipboard is returned when no usable clipboard utility is found.
var ErrNoClipboard = fmt.Errorf("no clipboard utility found on this system")

// cmdSpec is a candidate command plus its arguments.
type cmdSpec struct {
	name string
	args []string
}

// readCandidates returns the ordered list of read commands to try for the OS.
func readCandidates() []cmdSpec {
	switch runtime.GOOS {
	case "darwin":
		return []cmdSpec{{"pbpaste", nil}}
	case "windows":
		return []cmdSpec{{"powershell", []string{"-NoProfile", "-Command", "Get-Clipboard"}}}
	default: // linux, *bsd
		// Prefer Wayland, then X11. -o/-out = paste/output.
		return []cmdSpec{
			{"wl-paste", []string{"--no-newline"}},
			{"xclip", []string{"-selection", "clipboard", "-o"}},
			{"xsel", []string{"--clipboard", "--output"}},
		}
	}
}

// writeCandidates returns the ordered list of write commands to try for the OS.
func writeCandidates() []cmdSpec {
	switch runtime.GOOS {
	case "darwin":
		return []cmdSpec{{"pbcopy", nil}}
	case "windows":
		return []cmdSpec{{"clip", nil}}
	default:
		return []cmdSpec{
			{"wl-copy", nil},
			{"xclip", []string{"-selection", "clipboard"}},
			{"xsel", []string{"--clipboard", "--input"}},
		}
	}
}

func lookable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Read returns the current clipboard contents as a string.
func Read() (string, error) {
	for _, c := range readCandidates() {
		if !lookable(c.name) {
			continue
		}
		cmd := exec.Command(c.name, c.args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("%s failed: %w", c.name, err)
		}
		return out.String(), nil
	}
	return "", ErrNoClipboard
}

// Write replaces the clipboard contents with s.
func Write(s string) error {
	for _, c := range writeCandidates() {
		if !lookable(c.name) {
			continue
		}
		cmd := exec.Command(c.name, c.args...)
		cmd.Stdin = bytes.NewBufferString(s)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s failed: %w", c.name, err)
		}
		return nil
	}
	return ErrNoClipboard
}
