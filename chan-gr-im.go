package slack

import (
	"net/url"
	"strconv"
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

// HistoryResponse holds a response to a history request
type HistoryResponse struct {
	slackResponse
	Latest   string    `json:"latest"`
	HasMore  bool      `json:"has_more"`
	Messages []Message `json:"messages"`
}

// GroupResponse holds a response to a group request
type GroupResponse struct {
	slackResponse
	Group Group `json:"group"`
}

// ChannelListResponse holds a response to a channel list request
type ChannelListResponse struct {
	slackResponse
	Channels []Channel `json:"channels"`
}

// GroupListResponse holds a response to a group list request
type GroupListResponse struct {
	slackResponse
	Groups []Group `json:"groups"`
}

func prefixByID(id string) string {
	path := "channels."
	switch id[0] {
	case 'G':
		path = "groups."
	case 'D':
		path = "im."
	}
	return path
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

// History retrieves history of channel, group and IM
func (s *Slack) History(channel, latest, oldest string, inclusive bool, count int) (*HistoryResponse, error) {
	params := url.Values{"channel": {channel}}
	appendNotEmpty("latest", latest, params)
	appendNotEmpty("oldest", oldest, params)
	if inclusive {
		params.Set("inclusive", "1")
	}
	if count != 0 {
		params.Set("count", strconv.Itoa(count))
	}
	r := &HistoryResponse{}
	err := s.do(prefixByID(channel)+"history", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ChannelInvite invites a user to a group
func (s *Slack) ChannelInvite(channel, user string) (*ChannelResponse, error) {
	params := url.Values{"channel": {channel}, "user": {user}}
	r := &ChannelResponse{}
	err := s.do("channels.invite", params, r)
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
	path := prefixByID(channel) + "mark"
	err := s.do(path, params, r)
	if err != nil {
		return err
	}
	return nil
}

// GroupCreate creates a new group with the given name
func (s *Slack) GroupCreate(name string) (*GroupResponse, error) {
	params := url.Values{"name": {name}}
	r := &GroupResponse{}
	err := s.do("groups.create", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GroupInfo returns info about the group
func (s *Slack) GroupInfo(group string) (*GroupResponse, error) {
	params := url.Values{"channel": {group}}
	r := &GroupResponse{}
	err := s.do("groups.info", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GroupInvite invites a user to a group
func (s *Slack) GroupInvite(channel, user string) (*GroupResponse, error) {
	params := url.Values{"channel": {channel}, "user": {user}}
	r := &GroupResponse{}
	err := s.do("groups.invite", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GroupList returns the list of channels
func (s *Slack) GroupList(excludeArchived bool) (*GroupListResponse, error) {
	params := url.Values{}
	if excludeArchived {
		params.Set("exclude_archived", "1")
	}
	r := &GroupListResponse{}
	err := s.do("groups.list", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
