package slack

import "net/url"

// Bot holds the info about a bot
type Bot struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Deleted bool              `json:"deleted"`
	Icons   map[string]string `json:"icons"`
}

// UserProfile contains all the information details of a given user
type UserProfile struct {
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	RealName           string `json:"real_name"`
	RealNameNormalized string `json:"real_name_normalized"`
	Email              string `json:"email"`
	Skype              string `json:"skype"`
	Phone              string `json:"phone"`
	Image24            string `json:"image_24"`
	Image32            string `json:"image_32"`
	Image48            string `json:"image_48"`
	Image72            string `json:"image_72"`
	Image192           string `json:"image_192"`
	ImageOriginal      string `json:"image_original"`
	Title              string `json:"title"`
}

// User contains all the information of a user
type User struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Deleted           bool        `json:"deleted"`
	Color             string      `json:"color"`
	RealName          string      `json:"real_name"`
	TZ                string      `json:"tz,omitempty"`
	TZLabel           string      `json:"tz_label"`
	TZOffset          int         `json:"tz_offset"`
	Profile           UserProfile `json:"profile"`
	IsBot             bool        `json:"is_bot"`
	IsAdmin           bool        `json:"is_admin"`
	IsOwner           bool        `json:"is_owner"`
	IsPrimaryOwner    bool        `json:"is_primary_owner"`
	IsRestricted      bool        `json:"is_restricted"`
	IsUltraRestricted bool        `json:"is_ultra_restricted"`
	Has2FA            bool        `json:"has_2fa"`
	HasFiles          bool        `json:"has_files"`
	Presence          string      `json:"presence"`
}

// UserInfoResponse holds the response to a user info request
type UserInfoResponse struct {
	slackResponse
	User User `json:"user"`
}

// UserPresence contains details about a user online status
type UserPresence struct {
	Presence        string `json:"presence,omitempty"`
	Online          bool   `json:"online,omitempty"`
	AutoAway        bool   `json:"auto_away,omitempty"`
	ManualAway      bool   `json:"manual_away,omitempty"`
	ConnectionCount int    `json:"connection_count,omitempty"`
	LastActivity    int64  `json:"last_activity,omitempty"`
}

// Team holds information about the team
type Team struct {
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
}

// TeamInfoResponse holds thre response to the team info request
type TeamInfoResponse struct {
	slackResponse
	Team Team `json:"team"`
}

// TeamInfo returns info about the team
func (s *Slack) TeamInfo() (*TeamInfoResponse, error) {
	params := url.Values{}
	r := &TeamInfoResponse{}
	err := s.do("team.info", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// UserInfo returns info about the requested user
func (s *Slack) UserInfo(user string) (*UserInfoResponse, error) {
	params := url.Values{"user": {user}}
	r := &UserInfoResponse{}
	err := s.do("users.info", params, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
