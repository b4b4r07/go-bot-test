package main

import (
	"os"

	"github.com/b4b4r07/go-bot-test/slack"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		panic("slack token must be set")
	}

	s := slack.NewClient(token)
	s.Start()
}
