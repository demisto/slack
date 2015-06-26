package slack

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
	DeletedTS string   `json:"deleted_ts,omitempty"`
	Topic     string   `json:"topic,omitempty"`
	Purpose   string   `json:"purpose,omitempty"`
	Name      string   `json:"name,omitempty"`
	OldName   string   `json:"old_name,omitempty"`
	Members   []string `json:"members,omitempty"`
	Upload    bool     `json:"upload,omitempty"`
	File      File     `json:"file,omitempty"`
	Comment   Comment  `json:"comment,omitempty"`
	Error     struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"error"`
	Context interface{} `json:"context"` // A piece of data that will be passed with every message from RTMStart
}
