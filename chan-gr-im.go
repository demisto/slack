package slack

import (
	"net/url"
)

// ChannelTopicPurpose holds the topic or purpose of a channel
type ChannelTopicPurpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

// BaseChannel holds information about channel / group / IM
type BaseChannel struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Created            int64               `json:"created"`
	Creator            string              `json:"creator"`
	IsArchived         bool                `json:"is_archived"`
	IsOpen             bool                `json:"is_open"`
	Members            []string            `json:"members"`
	Topic              ChannelTopicPurpose `json:"topic"`
	Purpose            ChannelTopicPurpose `json:"purpose"`
	LastRead           string              `json:"last_read,omitempty"`
	Latest             Message             `json:"latest,omitempty"`
	UnreadCount        int                 `json:"unread_count,omitempty"`
	UnreadCountDisplay int                 `json:"unread_count_display,omitempty"`
	NumMembers         int                 `json:"num_members,omitempty"`
}

// Channel holds information about the channel
type Channel struct {
	BaseChannel
	IsGeneral bool `json:"is_general"`
	IsChannel bool `json:"is_channel"`
	IsMember  bool `json:"is_member"`
}

// Group holds information about the group
type Group struct {
	BaseChannel
	IsGroup bool `json:"is_group"`
}

// IM holds information about IM
type IM struct {
	BaseChannel
	IsIM          bool   `json:"is_im"`
	User          string `json:"user"`
	IsUserDeleted bool   `json:"is_user_deleted"`
}

// ChannelResponse holds a response to a channel request
type ChannelResponse struct {
	slackResponse
	Channel Channel `json:"channel"`
}

// ChannelListResponse holds a response to a channel list request
type ChannelListResponse struct {
	slackResponse
	Channels []Channel `json:"channels"`
}

// ChannelArchive archives a channel
func (s *Slack) ChannelArchive(channel string) (Response, error) {
	params := url.Values{"channel": {channel}}
	r := &slackResponse{}
	err := s.do("channels.archive", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ChannelCreate creates a channel
func (s *Slack) ChannelCreate(name string) (*ChannelResponse, error) {
	params := url.Values{"name": {name}}
	r := &ChannelResponse{}
	err := s.do("channels.create", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ChannelInfo returns info about the channel
func (s *Slack) ChannelInfo(channel string) (*ChannelResponse, error) {
	params := url.Values{"channel": {channel}}
	r := &ChannelResponse{}
	err := s.do("channels.info", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ChannelList returns the list of channels
func (s *Slack) ChannelList(excludeArchived bool) (*ChannelListResponse, error) {
	params := url.Values{}
	if excludeArchived {
		params.Set("exclude_archived", "1")
	}
	r := &ChannelListResponse{}
	err := s.do("channels.list", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Mark marks the given channel as read. Automatically detects channel/group/im
func (s *Slack) Mark(channel, ts string) error {
	r := &slackResponse{}
	params := url.Values{"channel": {channel}, "ts": {ts}}
	path := "channel.mark"
	switch channel[0:1] {
	case "G":
		path = "group.mark"
	case "D":
		path = "im.mark"
	}
	err := s.do(path, params, r)
	if err != nil {
		return err
	}
	return nil
}
