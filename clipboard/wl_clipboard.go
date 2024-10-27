//go:build linux

package clipboard

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func Check() bool {
	commands := []string{"wl-copy", "wl-paste"}
	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err != nil {
			panic(err)
		}
	}

	return true
}

func Watch(c *cli.Context, copiedBy string) chan Msg {
	ch := make(chan Msg)
	ticker := time.NewTicker(500 * time.Millisecond)
	skipInitial := true

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			format, mimeType, err := checkFormat()
			if err != nil {
				fmt.Printf("error: %v\n", err)
				continue
			}
			if format == FormatUnknown {
				continue
			}
			data, err := getData(mimeType)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				continue
			}

			msg := Msg{
				Format:   format,
				Payload:  hex.EncodeToString(data),
				CopiedBy: copiedBy,
				CopiedAt: time.Now(),
			}

			shouldSend := false
			if !LastClipboardContent.ContentEqual(msg) {
				if c.Bool("verbose") {
					if format == FormatText {
						fmt.Printf("copy: %s\n", string(data))
					}
					if format == FormatImage {
						fmt.Println("copy: [IMAGE]")
					}
				}

				shouldSend = true
			}
			LastClipboardContent = msg
			if shouldSend && !skipInitial {
				ch <- msg
			}

			skipInitial = false
		}
	}()

	return ch
}

func checkFormat() (format Format, mimeType string, err error) {
	cmd := exec.Command("wl-paste", "--list-types")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return FormatUnknown, "", fmt.Errorf("'wl-paste --list-types' error: %w", err)
	}

	mimeTypes := strings.Split(out.String(), "\n")
	hasText := false
	for _, mimeType := range mimeTypes {
		mimeType = strings.TrimSpace(mimeType)
		if mimeType == "text/plain" {
			hasText = true
		} else if strings.HasPrefix(mimeType, "image/") {
			return FormatImage, mimeType, nil
		}
	}
	if hasText {
		return FormatText, "text/plain", nil
	}

	return FormatUnknown, "", nil
}

func getData(mimeType string) ([]byte, error) {
	cmd := exec.Command("wl-paste", "--no-newline", "--type", mimeType)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("clipboard copy error: %w", err)
	}

	return out.Bytes(), nil
}

func Write(c *cli.Context, msg Msg, data []byte) error {
	if LastClipboardContent.ContentEqual(msg) {
		fmt.Printf("-> write: skipped (same payload)\n")
		return nil
	}
	LastClipboardContent = msg

	var cmd *exec.Cmd

	switch msg.Format {
	case FormatText:
		cmd = exec.Command("wl-copy", "--type", "text/plain")
	case FormatImage:
		contentType := http.DetectContentType(data)
		cmd = exec.Command("wl-copy", "--type", contentType)
	default:
		return fmt.Errorf("format error: %s", msg.Format)
	}

	cmd.Stdin = bytes.NewReader(data)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard write error: %w", err)
	}

	if c.Bool("verbose") {
		fmt.Printf("-> write: OK\n")
	}

	return nil
}
