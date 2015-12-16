package slack

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// TypedMessage holds the common fields to all messages
type TypedMessage interface {
	MessageType() string
	ErrorCode() int
	ErrorMsg() string
}

// baseTypeMessage is used to parse the message type before we parse the entire message
type baseTypeMessage struct {
	Type  string `json:"type"`
	Error struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"error,omitempty"`
}

// Message holds the information about incoming messages in the RTM
type Message struct {
	Type      string `json:"type"`
	Channel   string `json:"channel"`
	User      string `json:"user"`
	Text      string `json:"text"`
	Timestamp string `json:"ts"`
	Hidden    bool   `json:"hidden,omitempty"`
	Subtype   string `json:"subtype,omitempty"`
	Edited    struct {
		User      string `json:"user"`
		Timestamp string `json:"ts"`
	} `json:"edited,omitempty"`
	Message struct {
		Type      string `json:"type"`
		User      string `json:"user"`
		Text      string `json:"text"`
		Timestamp string `json:"ts"`
		Edited    struct {
			User      string `json:"user"`
			Timestamp string `json:"ts"`
		} `json:"edited,omitempty"`
	} `json:"message,omitempty"`
	DeletedTS      string      `json:"deleted_ts,omitempty"`
	Topic          string      `json:"topic,omitempty"`
	Purpose        string      `json:"purpose,omitempty"`
	Name           string      `json:"name,omitempty"`
	OldName        string      `json:"old_name,omitempty"`
	Members        []string    `json:"members,omitempty"`
	Upload         bool        `json:"upload,omitempty"`
	File           File        `json:"file,omitempty"`
	Comment        Comment     `json:"comment,omitempty"`
	Reactions      []Reaction  `json:"reactions,omitempty"`
	Presence       string      `json:"presence,omitempty"`
	Value          interface{} `json:"value,omitempty"`
	Plan           string      `json:"plan,omitempty"`
	URL            string      `json:"url,omitempty"`
	Domain         string      `json:"domain,omitempty"`
	EmailDomain    string      `json:"email_domain,omitempty"`
	EventTimestamp string      `json:"event_ts,omitempty"`
	Error          struct {
		Code       int    `json:"code"`
		Msg        string `json:"msg"`
		Unmarshall bool   `json:"unmarshall"` // Is this an unmarshall error and not request error
	} `json:"error,omitempty"`
	Context interface{} `json:"context,omitempty"` // A piece of data that will be passed with every message from RTMStart
}

// MessageType of message is returned
func (m *Message) MessageType() string {
	return m.Type
}

// ErrorCode if exists is returned
func (m *Message) ErrorCode() int {
	return m.Error.Code
}

// ErrorMsg if exists is returned
func (m *Message) ErrorMsg() string {
	return m.Error.Msg
}

// ChannelEvent is sent when a message contains an actual channel and not ID
type ChannelEvent struct {
	Type    string  `json:"type"`
	Channel Channel `json:"channel"`
}

// UserEvent is used when a message contains an actual user and not ID
type UserEvent struct {
	Type string `json:"type"`
	User User   `json:"user"`
}

// TimestampToTime converter
func TimestampToTime(timestamp string) (time.Time, error) {
	parts := strings.Split(timestamp, ".")
	if len(parts) > 0 && parts[0] != "" {
		sec, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(sec, 0), nil
	}
	return time.Time{}, errors.New("Invalid timestamp")
}
