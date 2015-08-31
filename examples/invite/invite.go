package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/demisto/slack"
)

var (
	token           = flag.String("token", "", "token to connect to slack")
	email           = flag.String("email", "", "The email to invite")
	first           = flag.String("first", "", "optional first name")
	last            = flag.String("last", "", "optional last name")
	channelNamesStr = flag.String("channels", "", "comma separated list of channels and groups")
	t               = flag.String("type", "regular", "type of invite - regular/restricted/ultra")
)

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	s, err := slack.New(
		slack.SetToken(*token),
		slack.SetErrorLog(log.New(os.Stderr, "ERR:", log.Lshortfile)),
		slack.SetTraceLog(log.New(os.Stderr, "DEBUG:", log.Lshortfile)))
	check(err)
	// Translate channel
	channels, err := s.ChannelList(true)
	var channelIDs []string
	channelNames := strings.Split(*channelNamesStr, ",")
	for i := range channelNames {
		for j := range channels.Channels {
			if strings.ToLower(channels.Channels[j].Name) == strings.ToLower(strings.TrimSpace(channelNames[i])) {
				channelIDs = append(channelIDs, channels.Channels[j].ID)
				break
			}
		}
	}
	groups, err := s.GroupList(true)
	check(err)
	for i := range channelNames {
		for j := range groups.Groups {
			if strings.ToLower(groups.Groups[j].Name) == strings.ToLower(strings.TrimSpace(channelNames[i])) {
				channelIDs = append(channelIDs, groups.Groups[j].ID)
				break
			}
		}
	}
	inviteType := slack.InviteeRegular
	if *t == "restricted" {
		inviteType = slack.InviteeRestricted
	} else if *t == "ultra" {
		inviteType = slack.InviteeUltraRestricted
	}
	err = s.InviteToSlack(slack.UserInviteDetails{Email: *email, FirstName: *first, LastName: *last}, channelIDs, inviteType)
	check(err)
}
