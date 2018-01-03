package slack

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

// AttachmentField holds information about an attachment field
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Attachment holds information about an attachment
type Attachment struct {
	ServiceName string            `json:"service_name"`
	Fallback    string            `json:"fallback"`
	Color       string            `json:"color,omitempty"`
	Pretext     string            `json:"pretext,omitempty"`
	AuthorName  string            `json:"author_name,omitempty"`
	AuthorLink  string            `json:"author_link,omitempty"`
	AuthorIcon  string            `json:"author_icon,omitempty"`
	Title       string            `json:"title,omitempty"`
	TitleLink   string            `json:"title_link,omitempty"`
	Text        string            `json:"text"`
	ImageURL    string            `json:"image_url,omitempty"`
	ThumbURL    string            `json:"thumb_url,omitempty"`
	ThumbWidth  int               `json:"thumb_width"`
	ThumbHeight int               `json:"thumb_height"`
	FromURL     string            `json:"from_url"`
	Fields      []AttachmentField `json:"fields,omitempty"`
	MarkdownIn  []string          `json:"mrkdwn_in,omitempty"`
}

// PostMessageRequest includes all the fields in the post message request - see https://api.slack.com/methods/chat.postMessage
type PostMessageRequest struct {
	Channel     string       `json:"channel"`
	Text        string       `json:"text"`
	Username    string       `json:"username"`
	AsUser      bool         `json:"as_user"`
	Parse       string       `json:"parse"`
	LinkNames   int          `json:"link_names"`
	Attachments []Attachment `json:"attachments"`
	UnfurlLinks bool         `json:"unfurl_links"`
	UnfurlMedia bool         `json:"unfurl_media"`
	IconURL     string       `json:"icon_url"`
	IconEmoji   string       `json:"icon_emoji"`
	ThreadID    string       `json:"thread_ts"`
}

// PostMessageReply is the reply to the post message request - see https://api.slack.com/methods/chat.postMessage
type PostMessageReply struct {
	slackResponse
	Channel   string             `json:"channel"`
	Timestamp string             `json:"ts"`
	Message   PostMessageRequest `json:"message"`
}

// PostMessage posts a message to a channel
func (s *Slack) PostMessage(m *PostMessageRequest, escape bool) (*PostMessageReply, error) {
	// Escape the special chars
	text := ""
	if escape {
		replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
		text = replacer.Replace(m.Text)
	} else {
		text = m.Text
	}
	params := url.Values{
		"channel": {m.Channel},
		"text":    {text},
	}
	if m.Username != "" {
		params.Set("username", m.Username)
	}
	params.Set("as_user", strconv.FormatBool(m.AsUser))
	params.Set("parse", m.Parse)
	params.Set("link_names", strconv.Itoa(m.LinkNames))
	params.Set("thread_id", m.ThreadID)
	if len(m.Attachments) > 0 {
		attachments, err := json.Marshal(m.Attachments)
		if err != nil {
			return nil, err
		}
		params.Set("attachments", string(attachments))
	}
	params.Set("unfurl_links", strconv.FormatBool(m.UnfurlLinks))
	params.Set("unfurl_media", strconv.FormatBool(m.UnfurlMedia))
	if m.IconURL != "" {
		params.Set("icon_url", m.IconURL)
	}
	if m.IconEmoji != "" {
		params.Set("icon_emoji", m.IconEmoji)
	}
	r := &PostMessageReply{}
	err := s.do("chat.postMessage", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
