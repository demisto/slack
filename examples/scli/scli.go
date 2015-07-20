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
	"github.com/slavikm/liner"
)

var (
	conf            = flag.String("c", "~/.scli", "Location of the configuration file")
	hist            = flag.String("h", "~/.scli_hist", "Location of the history file")
	verbose         = flag.Bool("v", true, "Be a bit more talkative about our internal behavior")
	token           = flag.String("t", "", "The Slack token which you can get at - https://api.slack.com/web")
	channel         = flag.String("ch", "", "Override the default channel")
	shouldLoadFiles = flag.Bool("loadfiles", false, "Should we load files for auto-completion")
	debug           = flag.Bool("debug", false, "Debug prints")
)

var (
	s             *slack.Slack
	info          *slack.RTMStartReply // The global info for the team
	currChannelID string               // The ID of the channel
	files         []slack.File         // The files for the team
)

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
				line.PrintAbovePrompt(fmt.Sprintf("%s %s: %s", time.Now().Format(time.Stamp), channelName(msg.Channel), msg.Text))
				latest[msg.Channel] = msg.Timestamp
			} else if msg.Type == "error" {
				line.PrintAbovePrompt(msg.Error.Msg)
				var err error
				// Try the channel and messaging
				for err != nil {
					time.Sleep(1 * time.Minute)
					in := make(chan slack.Message)
					info, err = s.RTMStart("", in, nil)
					if err == nil {
						line.PrintAbovePrompt("Reconnected...")
					}
				}
			}
		}
	}
}

func postMessage(msg string) {
	m := &slack.PostMessageRequest{
		AsUser:  true,
		Channel: currChannelID,
		Text:    msg,
	}
	_, err := s.PostMessage(m, true)
	if err != nil {
		fmt.Printf("Unable to post to channel %s - %v\n", channelName(currChannelID), err)
	}
}

// Manually load the info because we are not doing interactive
func loadInfo() {
	info = &slack.RTMStartReply{}
	channels, err := s.ChannelList(false)
	check(err)
	info.Channels = channels.Channels
	groups, err := s.GroupList(false)
	check(err)
	info.Groups = groups.Groups
	ims, err := s.IMList()
	check(err)
	info.IMS = ims.IMs
	users, err := s.UserList()
	check(err)
	info.Users = users.Members
}

// loadFiles loads all the files for the team - good for now but not very scalable
func loadFiles() {
	page, count := 1, 100
	for {
		r, err := s.FileList("", "", "", nil, count, page)
		if err != nil {
			fmt.Printf("Unable to load files, file auto-completion will not work - %v\n", err)
			return
		}
		files = append(files, r.Files...)
		if r.Paging.Count < count || r.Paging.Page == r.Paging.Pages {
			break
		}
		page++
	}
}

func cleanup(stop []chan bool) {
	for i := range stop {
		stop[i] <- true
	}
	s.RTMClose()
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
	s, err = slack.New(slack.SetErrorLog(log.New(os.Stderr, "", log.Lshortfile)), slack.SetToken(Options.Token))
	check(err)
	if *debug {
		slack.SetTraceLog(log.New(os.Stderr, "", log.Lshortfile))(s)
	}
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
	line.SetCompleter(completer)
	line.SetTabCompletionStyle(liner.TabPrints)

	var stop []chan bool
	if liner.TerminalSupported() && !line.InputRedirected() {
		stopHistory := make(chan bool)
		go saveHistory(line, stopHistory)
		stop = append(stop, stopHistory)

		in := make(chan slack.Message)
		info, err = s.RTMStart("", in, nil)
		check(err)
		if !switchChannel(Options.DefaultChannel) {
			fmt.Printf("Default channel %s not found\n", Options.DefaultChannel)
			for i := range info.Channels {
				if info.Channels[i].IsGeneral {
					currChannelID = info.Channels[i].ID
					fmt.Printf("Using %s as initial channel\n", channelName(currChannelID))
					break
				}
			}
		}
		stopReceiving := make(chan bool)
		stop = append(stop, stopReceiving)
		go receiveMessages(line, s, in, stopReceiving)
	} else {
		loadInfo()
	}
	if *shouldLoadFiles {
		loadFiles()
	}

	// The prompt loop
	for {
		if data, err := line.Prompt(channelName(currChannelID) + "> "); err != nil {
			if err != io.EOF && err != liner.ErrNotTerminalOutput {
				log.Print("Error reading line: ", err)
			}
			break
		} else {
			if len(data) == 0 {
				continue
			}
			line.AppendHistory(data)
			if strings.HasPrefix(data, Options.CommandPrefix) {
				shouldExit := handleCommand(data)
				if shouldExit {
					break
				}
			} else {
				go postMessage(data)
			}
		}
	}
	cleanup(stop)
}
