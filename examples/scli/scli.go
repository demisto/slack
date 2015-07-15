package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/demisto/slack"
	"github.com/peterh/liner"
)

var (
	conf    = flag.String("c", "~/.scli", "Location of the configuration file")
	hist    = flag.String("h", "~/.scli_hist", "Location of the history file")
	verbose = flag.Bool("v", true, "Be a bit more talkative about our internal behavior")
	token   = flag.String("t", "", "The Slack token which you can get at - https://api.slack.com/web")
	channel = flag.String("ch", "", "Override the default channel")
)

// The global info for the team
var info *slack.RTMStartReply

// The ID of the channel
var currChannelID string

func check(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

// expandHome expands the home dir if configuration flags include it
func expandHome() {
	// If conf / hist file is relative to home dir, expand it
	if len(*conf) > 2 && (*conf)[0:2] == "~/" ||
		len(*hist) > 2 && (*hist)[0:2] == "~/" {
		usr, err := user.Current()
		check(err)
		if len(*conf) > 2 && (*conf)[0:2] == "~/" {
			*conf = usr.HomeDir + (*conf)[1:]
		}
		if len(*hist) > 2 && (*hist)[0:2] == "~/" {
			*hist = usr.HomeDir + (*hist)[1:]
		}
	}
}

// saveHistory periodically saves the line history to the history file
func saveHistory(line *liner.State, stop chan bool) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			break
		case <-ticker.C:
			if f, err := os.Create(*hist); err != nil {
				if *verbose {
					log.Print("Error writing history file: ", err)
				}
			} else {
				line.WriteHistory(f)
				f.Close()
			}
		}
	}
}

func channelName(ch string) string {
	if ch == "" {
		return ""
	}
	switch ch[0] {
	case 'C':
		for i := range info.Channels {
			if info.Channels[i].ID == ch {
				return info.Channels[i].Name
			}
		}
	case 'G':
		for i := range info.Groups {
			if info.Groups[i].ID == ch {
				return info.Groups[i].Name
			}
		}
	case 'D':
		for i := range info.IMS {
			if info.IMS[i].ID == ch {
				return info.IMS[i].Name
			}
		}
	}
	return ""
}

func switchChannel(ch string) bool {
	for i := range info.Channels {
		if strings.ToLower(info.Channels[i].Name) == strings.ToLower(ch) {
			currChannelID = info.Channels[i].ID
			return true
		}
	}
	for i := range info.Groups {
		if strings.ToLower(info.Groups[i].Name) == strings.ToLower(ch) {
			currChannelID = info.Groups[i].ID
			return true
		}
	}
	for i := range info.IMS {
		if strings.ToLower(info.IMS[i].Name) == strings.ToLower(ch) {
			currChannelID = info.IMS[i].ID
			return true
		}
	}
	return false
}

func receiveMessages(line *liner.State, s *slack.Slack, in chan slack.Message, stop chan bool) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	latest := make(map[string]string)
	for {
		select {
		case <-stop:
			break
		case <-ticker.C:
			for k, v := range latest {
				s.Mark(k, v)
			}
		case msg := <-in:
			if msg.Type == "message" && msg.User != info.Self.ID {
				line.InjectString(channelName(msg.Channel) + ": " + msg.Text)
				latest[msg.Channel] = msg.Timestamp
			}
		}
	}
}

func cleanup(stopHistory, stopReceiving chan bool) {
	stopHistory <- true
	stopReceiving <- true
}

func main() {
	flag.Parse()
	expandHome()
	if *verbose {
		log.Printf("Using configuration file %s and history file %s\n", *conf, *hist)
	}
	err := Load(*conf)
	if err != nil && *verbose {
		log.Println("Unable load configuration, using defaults")
	}
	if *token != "" {
		Options.Token = *token
	}
	if Options.Token == "" {
		log.Println("Please provide the token from - https://api.slack.com/web")
		os.Exit(1)
	}
	if *channel != "" {
		Options.DefaultChannel = *channel
	}

	// Let's make sure that the token is valid before anything else
	s, err := slack.New(slack.SetToken(Options.Token))
	check(err)
	test, err := s.AuthTest()
	if err != nil {
		log.Println("Unable to authenticate to Slack: ", err)
	}
	if *verbose {
		log.Printf("Logged in as %s to team %s\n", test.User, test.Team)
	}

	line := liner.NewLiner()
	defer line.Close()

	if f, err := os.Open(*hist); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
	stopHistory := make(chan bool)
	go saveHistory(line, stopHistory)

	in := make(chan slack.Message)
	info, err = s.RTMStart("", in, nil)
	check(err)
	if !switchChannel(Options.DefaultChannel) {
		fmt.Printf("Default channel %s not found\n", Options.DefaultChannel)
	}

	stopReceiving := make(chan bool)
	go receiveMessages(line, s, in, stopReceiving)

	// The prompt loop
	for {
		if data, err := line.Prompt(channelName(currChannelID) + "> "); err != nil {
			switch err {
			case io.EOF:
				cleanup(stopHistory, stopReceiving)
			case liner.ErrNotTerminalOutput:
				cleanup(stopHistory, stopReceiving)
			default:
				log.Print("Error reading line: ", err)
			}
			break
		} else {
			if len(data) == 0 {
				continue
			}
			if len(data) > len(Options.CommandPrefix) && data[:len(Options.CommandPrefix)] == Options.CommandPrefix {
				parts := strings.Split(data, " ")
				cmd := strings.ToLower(parts[0][1:])
				switch cmd {
				case "exit":
					cleanup(stopHistory, stopReceiving)
					break
				case "ch":
					if len(parts) > 1 {
						if !switchChannel(parts[1]) {
							fmt.Printf("Unable to switch channel - channel %s not found\n", parts[1])
						} else {
							fmt.Printf("Switched to channel %s\n", parts[1])
						}
					}
				}
			} else {
				line.AppendHistory(data)
				m := &slack.PostMessageRequest{
					AsUser:  true,
					Channel: currChannelID,
					Text:    data,
				}
				_, err = s.PostMessage(m, true)
				if err != nil {
					fmt.Printf("Unable to post to channel %s - %v\n", channelName(currChannelID), err)
				}
			}
		}
	}
}
