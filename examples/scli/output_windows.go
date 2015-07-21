package main

import (
	"fmt"
	"time"

	"github.com/demisto/slack"
)

func format(msg *slack.Message) string {
	return fmt.Sprintf("%s %s [%s]: %s",
		time.Now().Format(time.Stamp), channelName(msg.Channel), findUser(msg.User).Name, msg.Text)
}
