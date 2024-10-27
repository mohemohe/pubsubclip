//go:build windows || darwin

package clipboard

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
)

func Check() bool {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	return true
}

func Watch(c *cli.Context, copiedBy string) chan Msg {
	ch1 := clipboard.Watch(context.TODO(), clipboard.FmtText)
	ch2 := clipboard.Watch(context.TODO(), clipboard.FmtImage)

	ch3 := make(chan Msg)

	go func() {
		var msg = Msg{}
		for {
			select {
			case b, _ := <-ch1:
				if c.Bool("verbose") {
					fmt.Printf("copy: %s\n", string(b))
				}

				msg = Msg{
					Format:   FormatText,
					Payload:  hex.EncodeToString(b),
					CopiedBy: copiedBy,
					CopiedAt: time.Now(),
				}
			case b, _ := <-ch2:
				if c.Bool("verbose") {
					fmt.Println("copy: [IMAGE]")
				}
				msg = Msg{
					Format:   FormatImage,
					Payload:  hex.EncodeToString(b),
					CopiedBy: copiedBy,
					CopiedAt: time.Now(),
				}
			}

			shuldSend := !LastClipboardContent.ContentEqual(msg)
			LastClipboardContent = msg
			if shuldSend {
				ch3 <- msg
			} else if c.Bool("verbose") {
				fmt.Printf("-> copy: skipped (same payload)\n")
			}
		}
	}()
	return ch3
}

func Write(c *cli.Context, msg Msg, data []byte) error {
	if msg.ContentEqual(LastClipboardContent) {
		fmt.Printf("-> write: skipped (same payload)\n")
		return nil
	}
	LastClipboardContent = msg

	switch msg.Format {
	case FormatText:
		clipboard.Write(clipboard.FmtText, data)
	case FormatImage:
		clipboard.Write(clipboard.FmtImage, data)
	default:
		return fmt.Errorf("format error: %s", msg.Format)
	}

	if c.Bool("verbose") {
		fmt.Printf("-> write: OK\n")
	}

	return nil
}
