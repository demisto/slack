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

// ChannelCommonResponse holds response to rename request
type ChannelCommonResponse struct {
	slackResponse
	Channel struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Created   int64  `json:"created"`
		IsChannel bool   `json:"is_channel"`
		IsGroup   bool   `json:"is_group"`
	} `json:"channel"`
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

// IMListResponse holds a response to an IM list request
type IMListResponse struct {
	slackResponse
	IMs []IM `json:"ims"`
}

// PurposeResponse is the response to setPurpose request
type PurposeResponse struct {
	slackResponse
	Purpose string `json:"purpose"`
}

// TopicResponse is the response to setTopic request
type TopicResponse struct {
	slackResponse
	Topic string `json:"topic"`
}

// CloseResponse is returned for close requests
type CloseResponse struct {
	slackResponse
	NoOp          bool `json:"no_op"`
	AlreadyClosed bool `json:"already_closed"`
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

// Archive a channel or a group
func (s *Slack) Archive(channel string) (Response, error) {
	params := url.Values{"channel": {channel}}
	r := &slackResponse{}
	err := s.do(prefixByID(channel)+"archive", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Unarchive a channel or a group
func (s *Slack) Unarchive(channel string) (Response, error) {
	params := url.Values{"channel": {channel}}
	r := &slackResponse{}
	err := s.do(prefixByID(channel)+"unarchive", params, r)
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

// Kick a user from a channel or group
func (s *Slack) Kick(channel, user string) (Response, error) {
	params := url.Values{"channel": {channel}, "user": {user}}
	r := &slackResponse{}
	err := s.do(prefixByID(channel)+"kick", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Leave a channel or a group
func (s *Slack) Leave(channel string) (Response, error) {
	params := url.Values{"channel": {channel}}
	r := &slackResponse{}
	err := s.do(prefixByID(channel)+"join", params, r)
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

// Rename a channel or a group
func (s *Slack) Rename(channel, name string) (*ChannelCommonResponse, error) {
	params := url.Values{"channel": {channel}, "name": {name}}
	r := &ChannelCommonResponse{}
	err := s.do(prefixByID(channel)+"rename", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// SetPurpose of the channel / group
func (s *Slack) SetPurpose(channel, purpose string) (*PurposeResponse, error) {
	params := url.Values{"channel": {channel}, "purpose": {purpose}}
	r := &PurposeResponse{}
	err := s.do(prefixByID(channel)+"setPurpose", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// SetTopic of the channel / group
func (s *Slack) SetTopic(channel, purpose string) (*TopicResponse, error) {
	params := url.Values{"channel": {channel}, "topic": {purpose}}
	r := &TopicResponse{}
	err := s.do(prefixByID(channel)+"setTopic", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// CloseGroupOrIM closes the given id
func (s *Slack) CloseGroupOrIM(id string) (*CloseResponse, error) {
	params := url.Values{"channel": {id}}
	r := &CloseResponse{}
	err := s.do(prefixByID(id)+"close", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// OpenGroupOrIM closes the given id
func (s *Slack) OpenGroupOrIM(id string) (*CloseResponse, error) {
	params := url.Values{"channel": {id}}
	r := &CloseResponse{}
	err := s.do(prefixByID(id)+"open", params, r)
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

// ChannelJoin joins a channel - notice that this expects channel name and not id
func (s *Slack) ChannelJoin(channel string) (*ChannelResponse, error) {
	params := url.Values{"name": {channel}}
	r := &ChannelResponse{}
	err := s.do("channels.join", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
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

// GroupCreateChild archives existing group and creates a new group with the given name
func (s *Slack) GroupCreateChild(group string) (*GroupResponse, error) {
	params := url.Values{"channel": {group}}
	r := &GroupResponse{}
	err := s.do("groups.createChild", params, r)
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

// GroupList returns the list of groups
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

// IMList returns the list of IMs
func (s *Slack) IMList() (*IMListResponse, error) {
	params := url.Values{}
	r := &IMListResponse{}
	err := s.do("ims.list", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
