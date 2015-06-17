package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/demisto/slack"
)

var token = flag.String("token", "", "token to connect to slack")
var file = flag.String("file", "", "file to upload")
var title = flag.String("title", "", "optional file title")
var comment = flag.String("comment", "", "optional initial comment")
var channel = flag.String("channel", "", "optional the channel to share to")

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	if *file == "" {
		check(fmt.Errorf("Please provide the filename to upload"))
	}
	s, err := slack.New(
		slack.SetToken(*token),
		slack.SetErrorLog(log.New(os.Stderr, "ERR:", log.Lshortfile)),
		slack.SetTraceLog(log.New(os.Stderr, "DEBUG:", log.Lshortfile)))
	check(err)
	f, err := os.Open(*file)
	check(err)
	defer f.Close()
	// Translate channel
	channels, err := s.ChannelList(true)
	var channelID string
	for i := range channels.Channels {
		if strings.ToLower(channels.Channels[i].Name) == strings.ToLower(*channel) {
			channelID = channels.Channels[i].ID
			break
		}
	}
	if channelID == "" {
		check(fmt.Errorf("Channel %s not found", *channel))
	}
	resp, err := s.Upload(*title, "", filepath.Base(*file), *comment, []string{channelID}, f)
	check(err)
	fmt.Printf("Response: %v", resp)
}
