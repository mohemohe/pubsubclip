package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mohemohe/pubsubclip/clipboard"
	"github.com/redis/go-redis/v9"
	// "github.com/rs/xid"
	"github.com/urfave/cli/v2"
)

var rdb *redis.Client
var lastMsg clipboard.Msg = clipboard.Msg{}
var copiedBy string
var Version string

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:  "version",
		Usage: "show version",
	}
	app := &cli.App{
		Name:      "pubsubclip",
		Usage:     "publish / subscribe clipboard via redis",
		Version:   Version,
		Copyright: "© 2024 mohemohe",
	}
	app.Commands = []*cli.Command{
		{
			Name:  "watch",
			Usage: "watch clipboard",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "addr",
					Usage:    "address of redis server",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "username",
					Usage: "username of redis server",
				},
				&cli.StringFlag{
					Name:  "password",
					Usage: "password of redis server",
				},
				&cli.StringFlag{
					Name:        "channel",
					Usage:       "channel name of redis pubsub",
					DefaultText: "pubsubclip",
					Value:       "pubsubclip",
				},
				&cli.BoolFlag{
					Name: "verbose",
				},
			},
			Action: run,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	_ = clipboard.Check()

	rdb = redis.NewClient(&redis.Options{
		Addr:     c.String("addr"),
		Username: c.String("username"),
		Password: c.String("password"),
	})
	cmd := rdb.Ping(context.TODO())
	if _, err := cmd.Result(); err != nil {
		panic(err)
	}

	// copiedBy = xid.New().String()
	if hostname, err := os.Hostname(); err == nil {
		copiedBy = hostname
	} else {
		panic(err)
	}

	fmt.Printf("start pubsubclip: %s\n", copiedBy)

	go watchClip(c)
	watchRedis(c)

	return nil
}

func watchClip(c *cli.Context) {
	ch := clipboard.Watch(c, copiedBy)
	msg := clipboard.Msg{}
	for {
		msg = <-ch

		if lastMsg.Format == clipboard.FormatImage && time.Now().Before(lastMsg.CopiedAt.Add(3*time.Second)) {
			fmt.Printf("-> publish: skipped (too early)\n")
			lastMsg = msg
			continue
		}

		if lastMsg.Format != msg.Format || lastMsg.Payload != msg.Payload {
			lastMsg = msg
			if b, err := json.Marshal(msg); err == nil {
				rdb.Publish(context.TODO(), c.String("channel"), string(b))
				fmt.Printf("-> publish: OK\n")
			} else {
				fmt.Printf("json marshal error: %v\n", err)
			}
		} else if c.Bool("verbose") {
			fmt.Printf("-> publish: skipped (same payload)\n")
		}
	}
}

func watchRedis(c *cli.Context) {
	s := rdb.Subscribe(context.TODO(), c.String("channel"))
	ch := s.Channel()
	msg := clipboard.Msg{}
	for rMsg := range ch {
		if err := json.Unmarshal([]byte(rMsg.Payload), &msg); err == nil {
			if msg.CopiedBy == copiedBy {
				if c.Bool("verbose") {
					fmt.Printf("paste: from %s, skipped (copied by myself)\n", msg.CopiedBy)
				}
				continue
			}

			if msg.Format == clipboard.FormatImage && lastMsg.Format == clipboard.FormatImage && msg.CopiedAt.Before(lastMsg.CopiedAt.Add(3*time.Second)) {
				fmt.Printf("paste: [IMAGE] from %s\n", msg.CopiedBy)
				fmt.Printf("-> write: skipped (too early)\n")
				continue
			}

			if lastMsg.Format != msg.Format || lastMsg.Payload != msg.Payload {
				lastMsg = msg
				if decoded, err := hex.DecodeString(msg.Payload); err == nil {
					if c.Bool("verbose") {
						if msg.Format == clipboard.FormatText {
							fmt.Printf("paste: %s from %s\n", string(decoded), msg.CopiedBy)
						} else if msg.Format == clipboard.FormatImage {
							fmt.Printf("paste: [IMAGE] from %s\n", msg.CopiedBy)
						}
					}
					clipboard.Write(msg.Format, decoded)
				} else {
					fmt.Printf("error: %v\n", err)
				}
			} else if c.Bool("verbose") {
				fmt.Printf("-> write: skipped (same payload)\n")
			}
		} else {
			fmt.Printf("json unmarshal error: %v\n", err)
		}
	}
}
