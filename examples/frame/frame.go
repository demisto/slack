package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/demisto/slack"
)

var (
	colors         = flag.String("colors", "", "Colors to use for the various parts of the message: date,dateSep,dateSepBack,user,channel,text,background")
	token          = flag.String("t", "", "The Slack token which you can get at - https://api.slack.com/web")
	address        = flag.String("address", ":8080", "The address we want to listen on")
	channels       = flag.String("ch", "", "Specify comma separated list of channels to display")
	certFile       = flag.String("cert", "", "The certificate file to serve HTTPS")
	keyFile        = flag.String("key", "", "The private key file to serve HTTPS")
	inviteChannels = flag.String("inviteChannels", "general", "Which channels to invite by default")
	secret         = flag.String("secret", "", "Secret to invites")
	debug          = flag.Bool("debug", false, "Debug prints")
)

const lastMessagesSize = 1000

var (
	s            *slack.Slack
	info         *slack.RTMStartReply // The global info for the team
	lastMessages = make(wsMessages, 0, lastMessagesSize)
)

func check(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func translateUser(id string) string {
	for i := range info.Users {
		if info.Users[i].ID == id {
			return info.Users[i].Name
		}
	}
	// Just return the ID for now - need to listen to add user event
	return id
}

func translateChannel(id string) string {
	for i := range info.Channels {
		if info.Channels[i].ID == id {
			return info.Channels[i].Name
		}
	}
	// Just return the ID for now - need to listen to add user event
	return id
}

func formatTime(msg *slack.Message) int64 {
	t, err := slack.TimestampToTime(msg.Timestamp)
	if err != nil {
		t = time.Now()
	}
	return t.Unix()
}

func interestedIn(ch string) bool {
	if *channels == "" {
		return true
	}
	list := strings.Split(*channels, ",")
	for _, c := range list {
		if strings.EqualFold(ch, c) {
			return true
		}
	}
	return false
}

func translateMessage(msg *slack.Message, ch string) *wsMessage {
	text := msg.Text
	for {
		index := strings.Index(text, "<@")
		if index < 0 {
			break
		}
		end := strings.Index(text[index+2:], ">")
		if end < 0 {
			break
		}
		parts := strings.Split(text[index+2:index+2+end], "|")
		if len(parts) == 2 {
			if parts[0][0] == 'U' {
				text = "@" + parts[1] + text[end+index+3:]
			} else {
				text = "#" + parts[1] + text[end+index+3:]
			}
		} else if len(parts) == 1 {
			if len(parts[0]) > 0 {
				if parts[0][0] == 'U' {
					text = "@" + translateUser(parts[0]) + text[end+index+3:]
				} else if parts[0][0] == 'C' {
					text = "#" + translateChannel(parts[0]) + text[end+index+3:]
				}
			}
		}
	}
	return &wsMessage{User: translateUser(msg.User), Channel: ch, Text: text, TS: formatTime(msg)}
}

func receiveMessages(s *slack.Slack, in chan *slack.Message, stop chan bool, handler *wsHandler) {
	for {
		select {
		case <-stop:
			break
		case msg := <-in:
			if msg == nil || msg.Type == "error" {
				if msg.Error.Unmarshall {
					// Ignore unmarshall errors
					continue
				}
				if msg == nil {
					log.Println("Incoming messages are closed, reconnecting")
				} else {
					log.Printf("Received error from Slack - %d (%v)\n", msg.ErrorCode(), msg.ErrorMsg())
				}
				s.RTMStop()
				close(in)
				in = make(chan *slack.Message)
				// Try the channel and messaging
				for {
					time.Sleep(1 * time.Minute)
					var err error
					info, err = s.RTMStart("", in, nil)
					if err == nil {
						break
					}
				}
			} else {
				switch msg.Type {
				case "message":
					ch := translateChannel(msg.Channel)
					if interestedIn(ch) {
						wsm := translateMessage(msg, ch)
						if wsm.Text != "" {
							if len(lastMessages) < lastMessagesSize {
								lastMessages = append(lastMessages, wsm)
							} else {
								// Copy so that it will not continue to grow
								newLastMessages := make([]*wsMessage, 0, 100)
								copy(newLastMessages, lastMessages[1:])
								lastMessages = append(newLastMessages, wsm)
							}
							if *debug {
								log.Printf("Sending: %v\n", wsm)
							}
							handler.send(wsm)
						}
					}
				case "channel_created":
					info.Channels = append(info.Channels, slack.Channel{BaseChannel: slack.BaseChannel{ID: msg.Channel, Name: msg.Name}})
				case "team_join":
					info.Users = append(info.Users, slack.User{ID: msg.User, Name: msg.Name})
				}
			}
		}
	}
}

func populateInitialMessages() {
	for i := range info.Channels {
		if interestedIn(info.Channels[i].Name) {
			r, err := s.History(info.Channels[i].ID, "", "", true, false, 100)
			if err != nil {
				log.Printf("Error retrieving history - %v", err)
				continue
			}
			if !r.IsOK() {
				log.Printf("Error retrieving history - %v", r.Error())
				continue
			}
			for m := range r.Messages {
				wsm := translateMessage(&r.Messages[m], info.Channels[i].Name)
				if wsm.Text != "" {
					if *debug {
						log.Printf("Appending to latest: %v\n", wsm)
					}
					lastMessages = append(lastMessages, wsm)
				}
			}
		}
	}
	sort.Sort(lastMessages)
}

func handleState(w http.ResponseWriter, r *http.Request) {
	colObj := make(map[string]string)
	params := strings.Split(*colors, ",")
	for _, c := range params {
		param := strings.Split(c, ":")
		if len(param) == 2 {
			field := strings.TrimSpace(strings.ToLower(param[0]))
			color := strings.TrimSpace(strings.ToLower(param[1]))
			if color != "" && field != "" {
				colObj[field] = color
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(colObj)
}

func handleHist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lastMessages)
}

func in(l []string, v string) bool {
	for _, i := range l {
		if i == v {
			return true
		}
	}
	return false
}

func handleInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	first := r.FormValue("fname")
	last := r.FormValue("lname")
	email := r.FormValue("email")
	if first == "" || last == "" || email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "Missing parameters"})
		return
	}
	header := r.Header.Get("secret")
	if header != *secret {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "Missing secret"})
		return
	}
	invCh := make([]string, 0)
	configuredChannels := strings.Split(*inviteChannels, ",")
	for i := range info.Channels {
		for j := range configuredChannels {
			if strings.EqualFold(info.Channels[i].Name, configuredChannels[j]) && !in(invCh, info.Channels[i].ID) {
				invCh = append(invCh, info.Channels[i].ID)
			}
		}
	}
	err := s.InviteToSlack(slack.UserInviteDetails{Email: email, FirstName: first, LastName: last}, invCh, slack.InviteeRegular)
	if err == nil || err.Error() == "already_in_team" || err.Error() == "already_invited" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "OK"})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": err.Error()})
	}
}

func main() {
	flag.Parse()
	if *token == "" {
		log.Println("Please provide the token from - https://api.slack.com/web")
		os.Exit(1)
	}
	if *secret == "" {
		log.Println("Please provide the secret for invites")
		os.Exit(1)
	}
	var err error
	// Let's make sure that the token is valid before anything else
	s, err = slack.New(slack.SetErrorLog(log.New(os.Stderr, "", log.Lshortfile)), slack.SetToken(*token))
	check(err)
	if *debug {
		slack.SetTraceLog(log.New(os.Stderr, "", log.Lshortfile))(s)
	}
	_, err = s.AuthTest()
	if err != nil {
		log.Println("Unable to authenticate to Slack: ", err)
	}
	handler := newWSHandler()
	defer handler.stop()
	go handler.writeLoop()
	in := make(chan *slack.Message)
	info, err = s.RTMStart("", in, nil)
	check(err)
	defer s.RTMStop()
	populateInitialMessages()
	stop := make(chan bool)
	go receiveMessages(s, in, stop, handler)
	defer func() {
		stop <- true
	}()
	http.Handle("/ws", handler)
	http.HandleFunc("/state", handleState)
	http.HandleFunc("/hist", handleHist)
	http.HandleFunc("/invite", handleInvite)
	http.Handle("/", http.FileServer(FS(*debug)))
	// HTTP stuff
	if *keyFile != "" && *certFile != "" {
		log.Fatal(http.ListenAndServeTLS(*address, *certFile, *keyFile, nil))
	}
	log.Fatal(http.ListenAndServe(*address, nil))
}
