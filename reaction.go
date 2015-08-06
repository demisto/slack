package slack

import (
	"errors"
	"net/url"
)

// Reaction contains the reaction details
type Reaction struct {
	Name  string   `json:"name"`
	Count int      `json:"count"`
	Uesrs []string `json:"users"`
}

// ReactionsGetResponse is the response to the ReactionsGet request
type ReactionsGetResponse struct {
	slackResponse
	Type    string  `json:"type"`
	Channel string  `json:"channel,omitempty"`
	Message Message `json:"message,omitempty"`
	File    File    `json:"file,omitempty"`
	Comment Comment `json:"comment,omitempty"`
}

// ReactionsListResponse is the response to the ReactionsList request
type ReactionsListResponse struct {
	slackResponse
	Items  ReactionsGetResponse `json:"items"`
	Paging paging               `json:"paging"`
}

func (s *Slack) reactionsAction(name, file, fileComment, channel, timestamp, action string) (Response, error) {
	if name == "" {
		return nil, errors.New("Please provide the emoji name")
	}
	if file == "" && fileComment == "" && (channel == "" || timestamp == "") {
		return nil, errors.New("Please provide file or fileComment or both channel and timestamp")
	}
	params := url.Values{"name": {name}}
	appendNotEmpty("file", file, params)
	appendNotEmpty("file_comment", fileComment, params)
	appendNotEmpty("channel", channel, params)
	appendNotEmpty("timestamp", timestamp, params)
	r := &slackResponse{}
	err := s.do(action, params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ReactionsAdd to either file, fileComment or a combination of channel and timestamp
func (s *Slack) ReactionsAdd(name, file, fileComment, channel, timestamp string) (Response, error) {
	return s.reactionsAction(name, file, fileComment, channel, timestamp, "reactions.add")
}

// ReactionsRemove from either file, fileComment or a combination of channel and timestamp
func (s *Slack) ReactionsRemove(name, file, fileComment, channel, timestamp string) (Response, error) {
	return s.reactionsAction(name, file, fileComment, channel, timestamp, "reactions.remove")
}

// ReactionsGet for either file, fileComment or a combination of channel and timestamp
func (s *Slack) ReactionsGet(file, fileComment, channel, timestamp string, full bool) (*ReactionsGetResponse, error) {
	params := url.Values{}
	appendNotEmpty("file", file, params)
	appendNotEmpty("file_comment", fileComment, params)
	appendNotEmpty("channel", channel, params)
	appendNotEmpty("timestamp", timestamp, params)
	if full {
		params.Set("full", "true")
	}
	r := &ReactionsGetResponse{}
	err := s.do("reactions.get", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ReactionsList filtered by user (optional and defaults to current)
func (s *Slack) ReactionsList(user string, full bool, count, page int) (*ReactionsListResponse, error) {
	params := url.Values{}
	appendNotEmpty("user", user, params)
	if full {
		params.Set("full", "true")
	}
	r := &ReactionsListResponse{}
	err := s.do("reactions.list", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
