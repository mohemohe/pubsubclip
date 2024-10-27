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
	msgChan := make(chan Msg)
	ticker := time.NewTicker(500 * time.Millisecond)

	go func() {
		defer ticker.Stop()

		var lastClipboardContent []byte

		for range ticker.C {
			format, err := checkFormat()
			if err != nil {
				fmt.Printf("error: %v\n", err)
				continue
			}
			data, err := getData()
			if err != nil {
				fmt.Printf("error: %v\n", err)
				continue
			}

			if lastClipboardContent == nil {
				lastClipboardContent = data
				continue
			}

			if bytes.Compare(lastClipboardContent, data) != 0 {
				msg := Msg{
					Format:   format,
					Payload:  hex.EncodeToString(data),
					CopiedBy: copiedBy,
					CopiedAt: time.Now(),
				}
				lastClipboardContent = data

				msgChan <- msg
			}
		}
	}()

	return msgChan
}

func checkFormat() (format Format, err error) {
	cmd := exec.Command("wl-paste", "--list-types")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return FormatUnknown, fmt.Errorf("'wl-paste --list-types' error: %w", err)
	}

	mimeTypes := strings.Split(out.String(), "\n")
	for _, mimeType := range mimeTypes {
		mimeType = strings.TrimSpace(mimeType)
		if mimeType == "text/plain" {
			return FormatText, nil
		} else if strings.HasPrefix(mimeType, "image/") {
			return FormatImage, nil
		}
	}

	return FormatUnknown, nil
}

func getData() ([]byte, error) {
	cmd := exec.Command("wl-paste", "--no-newline")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("clipboard image copy error: %w", err)
	}

	return out.Bytes(), nil
}

func Write(format Format, data []byte) error {
	var cmd *exec.Cmd

	switch format {
	case FormatText:
		cmd = exec.Command("wl-copy", "--type", "text/plain")
	case FormatImage:
		contentType := http.DetectContentType(data)
		cmd = exec.Command("wl-copy", "--type", contentType)
	default:
		return fmt.Errorf("format error: %d", format)
	}

	cmd.Stdin = bytes.NewReader(data)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard write error: %w", err)
	}

	return nil
}
