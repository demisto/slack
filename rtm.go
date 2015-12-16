package slack

import (
	"encoding/json"
	"errors"
	"io"
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
			var unmarshallError bool
			// Manually read the next message so that if there is JSON error we can
			// dump the error to log
			// This is a hack because the events have different fields with different structs
			// while initially we defined just a simple Message event so now we need to translate
			// the various events to message to keep compatibility
			_, p, err := ws.ReadMessage()
			if err == nil {
				typeMsg := &baseTypeMessage{}
				err = json.Unmarshal(p, typeMsg)
				// Ignore specific messages like user_change for now
				if err == nil {
					switch typeMsg.Type {
					case "channel_created", "channel_joined", "channel_rename", "im_created", "group_joined", "group_left", "group_rename":
						channelEvent := &ChannelEvent{}
						err = json.Unmarshal(p, channelEvent)
						if err == nil {
							msg.Type = channelEvent.Type
							msg.Channel = channelEvent.Channel.ID
							msg.User = channelEvent.Channel.Creator
							msg.Name = channelEvent.Channel.Name
						}
					case "user_change", "team_join":
						userEvent := &UserEvent{}
						err = json.Unmarshal(p, userEvent)
						if err == nil {
							msg.Type = userEvent.Type
							msg.User = userEvent.User.ID
							msg.Name = userEvent.User.Name
						}
					default:
						err = json.Unmarshal(p, msg)
					}
				}
				if err == io.EOF {
					// One value is expected in the message.
					err = io.ErrUnexpectedEOF
				}
				if err != nil {
					s.errorf("Error unmarshaling message - %s\n", string(p))
					unmarshallError = true
				}
			}
			// err := ws.ReadJSON(msg)
			if err != nil {
				msg.Type = "error"
				msg.Error.Code, msg.Error.Msg = 0, err.Error()
				msg.Error.Unmarshall = unmarshallError
			}
			// Set the custom data for every message
			msg.Context = context
			in <- msg
			if err != nil && !unmarshallError {
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
	if s.ws != nil {
		err := s.ws.Close()
		s.ws = nil
		return err
	}
	return nil
}
