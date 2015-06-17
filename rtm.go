package slack

import (
	"net/http"
	"net/url"
	"time"

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
	Team struct {
		ID          string                 `json:"id"`
		Name        string                 `json:"name"`
		EmailDomain string                 `json:"email_domain"`
		Domain      string                 `json:"domain"`
		Prefs       map[string]interface{} `json:"prefs"`
		Icon        struct {
			Image34      string `json:"image_34"`
			Image44      string `json:"image_44"`
			Image68      string `json:"image_68"`
			Image88      string `json:"image_88"`
			Image102     string `json:"image_102"`
			Image132     string `json:"image_132"`
			ImageDefault bool   `json:"image_default"`
		} `json:"icon"`
		OverStorageLimit bool   `json:"over_storage_limit"`
		Plan             string `json:"plan"`
	} `json:"team"`
	LatestEventTS string    `json:"latest_event_ts"`
	Channels      []Channel `json:"channels"`
	Groups        []Group   `json:"groups"`
	IMS           []IM      `json:"ims"`
	Users         []User    `json:"users"`
	Bots          []Bot     `json:"bots"`
}

// RTMStart starts the websocket
func (s *Slack) RTMStart(origin string, in chan Message) (*RTMStartReply, error) {
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
	markChannel := make(chan Message)
	// 5 sec timer to mark channels, groups, etc.
	go func() {
		messages := make(map[string]Message)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg := <-markChannel:
				messages[msg.Channel] = msg
			case <-ticker.C:
				for _, v := range messages {
					s.Mark(v.Channel, v.Timestamp)
				}
				messages = make(map[string]Message)
			}
		}
	}()
	// Start reading the messages and pumping them to the channel
	go func(ws *websocket.Conn, in chan Message) {
		defer func() {
			ws.Close()
		}()
		// Make sure we are receiving pongs
		// ws.SetReadDeadline(t)
		for {
			msg := Message{}
			err := ws.ReadJSON(&msg)
			if err != nil {
				msg.Type = "error"
				msg.Error.Code, msg.Error.Msg = 0, err.Error()
			}
			in <- msg
			if err != nil {
				break
			}
			// Push message to be marked as read
			if msg.Type == "message" {
				markChannel <- msg
			}
		}
	}(s.ws, in)
	return r, nil
}
