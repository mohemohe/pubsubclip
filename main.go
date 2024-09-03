package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"golang.design/x/clipboard"
)

type (
	Msg struct {
		Format  clipboard.Format
		Payload string
	}
)

var rdb *redis.Client
var lastMsg Msg = Msg{}
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
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     c.String("addr"),
		Username: c.String("username"),
		Password: c.String("password"),
	})
	cmd := rdb.Ping(context.TODO())
	if _, err := cmd.Result(); err != nil {
		panic(err)
	}

	go watchClip(c)
	watchRedis(c)

	return nil
}

func watchClip(c *cli.Context) {
	ch1 := clipboard.Watch(context.TODO(), clipboard.FmtText)
	ch2 := clipboard.Watch(context.TODO(), clipboard.FmtImage)
	var msg = Msg{}
	for {
		select {
		case b, _ := <-ch1:
			if c.Bool("verbose") {
				fmt.Printf("copy: %s\n", string(b))
			}

			msg = Msg{
				clipboard.FmtText,
				hex.EncodeToString(b),
			}
		case b, _ := <-ch2:
			if c.Bool("verbose") {
				fmt.Println("copy: [IMAGE]")
			}
			msg = Msg{
				clipboard.FmtImage,
				hex.EncodeToString(b),
			}
		}

		if lastMsg.Format != msg.Format || lastMsg.Payload != msg.Payload {
			lastMsg = msg
			if b, err := json.Marshal(msg); err == nil {
				rdb.Publish(context.TODO(), c.String("channel"), string(b))
			} else {
				fmt.Printf("json marshal error: %v\n", err)
			}
		} else if c.Bool("verbose") {
			fmt.Printf("-> copy: skipped\n")
		}
	}
}

func watchRedis(c *cli.Context) {
	s := rdb.Subscribe(context.TODO(), c.String("channel"))
	ch := s.Channel()
	msg := Msg{}
	for rMsg := range ch {
		if err := json.Unmarshal([]byte(rMsg.Payload), &msg); err == nil {
			if lastMsg.Format != msg.Format || lastMsg.Payload != msg.Payload {
				lastMsg = msg
				if decoded, err := hex.DecodeString(msg.Payload); err == nil {
					if c.Bool("verbose") {
						if msg.Format == clipboard.FmtText {
							fmt.Printf("paste: %s\n", string(decoded))
						} else if msg.Format == clipboard.FmtImage {
							fmt.Println("paste: [IMAGE]")
						}
					}
					clipboard.Write(msg.Format, decoded)
				} else {
					fmt.Printf("error: %v\n", err)
				}
			} else if c.Bool("verbose") {
				fmt.Printf("-> paste: skipped\n")
			}
		} else {
			fmt.Printf("json unmarshal error: %v\n", err)
		}
	}
}
