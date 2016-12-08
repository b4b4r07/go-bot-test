package slack

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

// Slack represets a Slack bot.
type Slack struct {
	token  string // slack token
	rtm    *slack.RTM
	api    *slack.Client
	botUID string

	// singleton channel name
	channel string
}

// NewClient creates a new slack bot.
func NewClient(token string) *Slack {
	if len(token) < 1 {
		panic("can't seem to start myself")
	}
	api := slack.New(token)

	s := &Slack{token: token, api: api, rtm: api.NewRTM()}
	go s.rtm.ManageConnection()

	return s
}

func (s *Slack) wasMentioned(msg string) bool {
	if len(msg) < 1 {
		return false
	}
	return strings.Contains(msg, s.botUID)
}

// expect some byte and write to slack
func (s *Slack) Write(o []byte) (n int, err error) {
	outBuf := bytes.Buffer{}
	outBuf.Write(o)

	params := slack.NewPostMessageParameters()
	params.Username = "supslack"
	params.AsUser = true

	params.Attachments = []slack.Attachment{
		{
			Text:       fmt.Sprintf("%s", outBuf.String()),
			MarkdownIn: []string{"text"},
		},
	}

	s.api.PostMessage(s.channel, "", params)
	return len(o), nil
}

// Start waits for Slack events.
func (s *Slack) Start() {
Loop:
	for {
		select {
		case msg := <-s.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				log.Println("slackbot: hello dave.")
			case *slack.ConnectedEvent:
				log.Println("slackbot: I'm online dave.")
				for _, ch := range ev.Info.Channels {
					log.Printf("slackbot: joined channel %s\n", ch.Name)
					s.rtm.SendMessage(
						s.rtm.NewOutgoingMessage(
							ch.Name,
							"Never send a human to do a machine's job.",
						),
					)
				}
				s.botUID = fmt.Sprintf("<@%s>: ", ev.Info.User.ID)
			case *slack.MessageEvent:
				s.channel = ev.Msg.Channel
				r, _ := regexp.Compile(`(bot hash)\s`)
				if r.MatchString(ev.Text) {
					log.Printf("slackbot: joined channel %s\n", ev.Text)
					s.Write([]byte(ev.Text))
				}
			case *slack.InvalidAuthEvent:
				log.Println("I seem to be disconnected, can't let you do that.")
				break Loop
			}
		}
	}
}
