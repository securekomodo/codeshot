// Package clipboard puts a PNG file on the system clipboard by shelling
// out to the platform's clipboard tool.
package clipboard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Copier copies a PNG file to the clipboard.
type Copier struct {
	Name string
	run  func(path string) error
}

// CopyPNG places the PNG at path on the clipboard.
func (c Copier) CopyPNG(path string) error {
	if err := c.run(path); err != nil {
		return fmt.Errorf("%s: %w", c.Name, err)
	}
	return nil
}

// Detect picks the clipboard tool for this platform.
func Detect() (Copier, error) {
	switch runtime.GOOS {
	case "darwin":
		return Copier{"osascript", func(path string) error {
			esc := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(path)
			script := fmt.Sprintf(`set the clipboard to (read (POSIX file "%s") as «class PNGf»)`, esc)
			return run("osascript", "-e", script)
		}}, nil
	case "linux":
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			if _, err := exec.LookPath("wl-copy"); err == nil {
				return Copier{"wl-copy", func(path string) error {
					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()
					cmd := exec.Command("wl-copy", "--type", "image/png")
					cmd.Stdin = f
					out, err := cmd.CombinedOutput()
					if err != nil {
						return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
					}
					return nil
				}}, nil
			}
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return Copier{"xclip", func(path string) error {
				return run("xclip", "-selection", "clipboard", "-t", "image/png", "-i", path)
			}}, nil
		}
		return Copier{}, errors.New("no clipboard tool found: install wl-clipboard (Wayland) or xclip (X11)")
	case "windows":
		return Copier{"powershell", func(path string) error {
			esc := strings.ReplaceAll(path, "'", "''")
			script := "Add-Type -AssemblyName System.Windows.Forms,System.Drawing; " +
				"[System.Windows.Forms.Clipboard]::SetImage([System.Drawing.Image]::FromFile('" + esc + "'))"
			return run("powershell", "-NoProfile", "-STA", "-Command", script)
		}}, nil
	}
	return Copier{}, fmt.Errorf("clipboard copy is not supported on %s", runtime.GOOS)
}

func run(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
