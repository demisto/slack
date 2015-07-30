package slack

import "net/url"

// EmojiListResponse is returned for the emoji list request
type EmojiListResponse struct {
	slackResponse
	Emoji map[string]string `json:"emoji"`
}

// EmojiList returns the list of emoji
func (s *Slack) EmojiList() (*EmojiListResponse, error) {
	params := url.Values{}
	r := &EmojiListResponse{}
	err := s.do("emoji.list", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
