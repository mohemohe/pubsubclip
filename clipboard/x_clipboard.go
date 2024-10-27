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
			ch3 <- msg
		}
	}()
	return ch3
}

func Write(format Format, data []byte) error {
	if format == FormatText {
		clipboard.Write(clipboard.FmtText, data)
	} else if format == FormatImage {
		clipboard.Write(clipboard.FmtImage, data)
	} else {
		return fmt.Errorf("format error: %d", format)
	}
	return nil
}
