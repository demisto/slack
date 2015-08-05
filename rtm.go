package slack

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

// WSMessageResponse holds a response to a WS request
type WSMessageResponse struct {
	OK      bool `json:"ok"`
	ReplyTo int  `json:"reply_to"`
	Error   struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"error"`
}

// RTMStartReply holds the reply to the RTM start message with info about everything
type RTMStartReply struct {
	slackResponse
	URL  string `json:"url"`
	Self struct {
		ID             string                 `json:"id"`
		Name           string                 `json:"name"`
		Prefs          map[string]interface{} `json:"prefs"`
		Created        int64                  `json:"created"`
		ManualPresence string                 `json:"manual_presence"`
	} `json:"self"`
	Team          Team      `json:"team"`
	LatestEventTS string    `json:"latest_event_ts"`
	Channels      []Channel `json:"channels"`
	Groups        []Group   `json:"groups"`
	IMS           []IM      `json:"ims"`
	Users         []User    `json:"users"`
	Bots          []Bot     `json:"bots"`
}

// RTMStart starts the websocket
func (s *Slack) RTMStart(origin string, in chan *Message, context interface{}) (*RTMStartReply, error) {
	r := &RTMStartReply{}
	err := s.do("rtm.start", url.Values{}, r)
	if err != nil {
		return nil, err
	}
	header := http.Header{"Origin": {origin}}
	s.ws, _, err = websocket.DefaultDialer.Dial(r.URL, header)
	if err != nil {
		return nil, err
	}
	// Start reading the messages and pumping them to the channel
	go func(ws *websocket.Conn, in chan *Message) {
		defer func() {
			ws.Close()
		}()
		// Make sure we are receiving pongs
		// ws.SetReadDeadline(t)
		for {
			msg := &Message{}
			err := ws.ReadJSON(msg)
			if err != nil {
				msg.Type = "error"
				msg.Error.Code, msg.Error.Msg = 0, err.Error()
			}
			// Set the custom data for every message
			msg.Context = context
			in <- msg
			if err != nil {
				close(in)
				break
			}
		}
	}(s.ws, in)
	return r, nil
}

// RTMMessage is sent on the channel for simple text
type RTMMessage struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// RTMSend a simple text message to a channel/group/dm
func (s *Slack) RTMSend(channel, text string) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.ws == nil {
		return 0, errors.New("RTM channel is not open")
	}
	s.mid++
	err := s.ws.WriteJSON(&RTMMessage{
		ID:      s.mid,
		Type:    "message",
		Channel: channel,
		Text:    text,
	})
	return s.mid, err
}

// RTMStop closes the WebSocket which in turn closes the in channel passed in RTMStart
func (s *Slack) RTMStop() error {
	err := s.ws.Close()
	s.ws = nil
	return err
}
