package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/demisto/slack"
)

const (
	bucket = "hist"
)

var (
	token    = flag.String("t", "", "The Slack token which you can get at - https://api.slack.com/web")
	address  = flag.String("address", ":8080", "The address we want to listen on")
	channels = flag.String("ch", "", "Specify comma separated list of channels to archive - all if not specified")
	certFile = flag.String("cert", "", "The certificate file to serve HTTPS")
	keyFile  = flag.String("key", "", "The private key file to serve HTTPS")
	dbFile   = flag.String("db", "", "The database to store the messages to")
	debug    = flag.Bool("debug", false, "Debug prints")
)

var (
	s    *slack.Slack         // The Slack API client
	info *slack.RTMStartReply // The global info for the team
	h    *handler             // The handler of the messages
)

// Bytify return a json []byte that is not indented
func Bytify(i interface{}) []byte {
	b, err := json.Marshal(i)
	if err != nil {
		return nil
	}
	return b
}

// message that we save in the DB
type message struct {
	User    string    `json:"user"`
	Channel string    `json:"channel"`
	Text    string    `json:"text"`
	TS      time.Time `json:"ts"`
}

// Key for the message that is also sortable
func (m *message) Key() string {
	return m.TS.Format(time.RFC3339) + "|" + m.User
}

// handler of the DB saving
type handler struct {
	messages chan *message // messages channel to receive messages on
	db       *bolt.DB      // db to store the messages
}

// Handle the message
func (h *handler) Handle(m *message) {
	h.messages <- m
}

// Loop to save the incoming messages in the DB
func (h *handler) Loop() {
	for {
		m, ok := <-h.messages
		if !ok {
			return
		}
		if m == nil {
			continue
		}
		err := h.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))
			return b.Put([]byte(m.Key()), Bytify(m))
		})
		check(err)
	}
}

func (h *handler) Hist(from, to string) (messages []message, err error) {
	f := time.Time{}
	t := time.Time{}

	if from != "" {
		// Format is year and month
		if len(from) == 6 {
			f, err = time.Parse("200601", from)
		} else {
			f, err = time.Parse(time.RFC3339, from)
		}
		if err != nil {
			return
		}
	}
	if to != "" {
		// Format is month and year
		if len(to) == 6 {
			t, err = time.Parse("200601", to)
		} else {
			t, err = time.Parse(time.RFC3339, to)
		}
		if err != nil {
			return
		}
	}
	err = h.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()
		// Iterate over the range
		for k, v := c.Seek([]byte(f.Format(time.RFC3339))); k != nil && (t.IsZero() || bytes.Compare(k, []byte(t.Format(time.RFC3339))) <= 0); k, v = c.Next() {
			var msg message
			err := json.Unmarshal(v, &msg)
			if err != nil {
				return err
			}
			messages = append(messages, msg)
		}
		return nil
	})
	return
}

// NewHandler to open the DB, etc
func NewHandler(dbFile string) (h *handler, err error) {
	err = os.MkdirAll(filepath.Dir(dbFile), 0755)
	if err != nil {
		return
	}
	h = &handler{messages: make(chan *message, 1000)}
	h.db, err = bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 30 * time.Second})
	if err != nil {
		return
	}
	err = h.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucket))
		return err
	})
	go h.Loop()
	return
}

// Stop the handler loop
func (h *handler) Stop() {
	if h.messages != nil {
		close(h.messages)
		h.messages = nil
	}
	h.db.Close()
}

func check(err error) {
	if err != nil {
		log.Fatalf("%v\n", err)
	}
}

func translateUser(id string) string {
	for i := range info.Users {
		if info.Users[i].ID == id {
			return info.Users[i].Name
		}
	}
	u, err := s.UserInfo(id)
	if err != nil {
		log.Printf("Unable to locate user %s\n", id)
		return id
	}
	info.Users = append(info.Users, u.User)
	return u.User.Name
}

func translateChannel(id string) string {
	if id[0] != 'C' {
		return id
	}
	for i := range info.Channels {
		if info.Channels[i].ID == id {
			return info.Channels[i].Name
		}
	}
	c, err := s.ChannelInfo(id)
	if err != nil {
		log.Printf("Unable to locate channel %s\n", id)
		return id
	}
	info.Channels = append(info.Channels, c.Channel)
	return c.Channel.Name
}

func formatTime(msg *slack.Message) time.Time {
	t, err := slack.TimestampToTime(msg.Timestamp)
	if err != nil {
		t = time.Now()
	}
	return t
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

func translateMessage(msg *slack.Message, ch string) *message {
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
	return &message{User: translateUser(msg.User), Channel: ch, Text: text, TS: formatTime(msg)}
}

func receiveMessages(s *slack.Slack, in chan *slack.Message, stop chan bool, h *handler) {
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
					if msg.Channel[0] != 'C' {
						continue
					}
					ch := translateChannel(msg.Channel)
					if interestedIn(ch) {
						m := translateMessage(msg, ch)
						if m.Text != "" {
							if *debug {
								log.Printf("Saving: %v\n", m)
							}
							h.Handle(m)
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

// handleHist of archive - send messages back to front end based on month and year
func handleHist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	messages, err := h.Hist(r.FormValue("from"), r.FormValue("to"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": http.StatusInternalServerError, "message": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(messages)
}

func main() {
	flag.Parse()
	if *token == "" {
		log.Fatal("Please provide the token from - https://api.slack.com/web")
	}
	var err error
	// Let's make sure that the token is valid before anything else
	s, err = slack.New(slack.SetErrorLog(log.New(os.Stderr, "", log.Lshortfile)), slack.SetToken(*token))
	check(err)
	if *debug {
		slack.SetTraceLog(log.New(os.Stderr, "", log.Lshortfile))(s)
	}
	test, err := s.AuthTest()
	check(err)
	log.Printf("Connected to Slack team [%s] with user [%s]\n", test.Team, test.User)
	h, err = NewHandler(*dbFile)
	defer h.Stop()
	log.Printf("Database [%s] open\n", *dbFile)
	in := make(chan *slack.Message)
	info, err = s.RTMStart("", in, nil)
	check(err)
	defer s.RTMStop()
	log.Println("Started RTM socket")
	stop := make(chan bool, 1)
	go receiveMessages(s, in, stop, h)
	defer func() {
		stop <- true
	}()
	shut := make(chan bool, 1) // shutdown channel
	http.HandleFunc("/hist", handleHist)
	// HTTP stuff
	go func() {
		log.Printf("Starting listener on %s\n", *address)
		if *keyFile != "" && *certFile != "" {
			err = http.ListenAndServeTLS(*address, *certFile, *keyFile, nil)
		} else {
			err = http.ListenAndServe(*address, nil)
		}
		shut <- true
	}()
	// Handle OS signals to gracefully shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	log.Println("Listening to OS signals")
	// Block until one of the signals above is received
	select {
	case <-signalCh:
		log.Println("Signal received, initializing clean shutdown...")
	case <-shut:
		log.Println("A service went down, shutting down...")
	}
	closeChannel := make(chan bool)
	go func() {
		h.Stop()
		s.RTMStop()
		stop <- true
		closeChannel <- true
	}()
	// Block again until another signal is received, a shutdown timeout elapses,
	// or the Command is gracefully closed
	log.Println("Waiting for clean shutdown...")
	select {
	case <-signalCh:
		log.Println("Second signal received, initializing hard shutdown")
	case <-time.After(time.Second * 30):
		log.Println("Time limit reached, initializing hard shutdown")
	case <-closeChannel:
	}
}
