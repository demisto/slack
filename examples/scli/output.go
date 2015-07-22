// +build !windows

package main

import (
	"fmt"
	"time"

	"github.com/demisto/slack"
)

var colors = map[string]string{
	"black":   "\x1b[30m",
	"red":     "\x1b[31m",
	"green":   "\x1b[32m",
	"yellow":  "\x1b[33m",
	"blue":    "\x1b[34m",
	"magenta": "\x1b[35m",
	"cyan":    "\x1b[36m",
	"white":   "\x1b[37m",
	"reset":   "\x1b[39;49m",
}

func safeUser(user *slack.User) *slack.User {
	if user == nil {
		return &slack.User{}
	}
	return user
}

func format(msg *slack.Message) string {
	return fmt.Sprintf("%s%s %s%s %s[%s]: %s%s%s",
		colors[Options.Colors["date"]], time.Now().Format(time.Stamp),
		colors[Options.Colors["channel"]], channelName(msg.Channel),
		colors[Options.Colors["user"]], safeUser(findUser(msg.User)).Name,
		colors[Options.Colors["text"]], msg.Text,
		colors["reset"])
}
