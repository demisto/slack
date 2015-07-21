package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Options anonymous struct holds the global configuration options
var Options struct {
	Token          string            // The token to authenticate to Slack
	CommandPrefix  string            // For internal commands, what is the prefix we are expecting
	SingleRoom     bool              // Should we show all messages with channel prefix in a single window or switch between channels / groups / IM
	DefaultChannel string            // The default channel we will post to at the start
	ChannelTimeout time.Duration     // After how long we revert back to the default channel for our messages
	Colors         map[string]string // The colors to use for output
}

// Load loads configuration from a file.
func Load(filename string) error {
	Options.CommandPrefix, Options.SingleRoom, Options.DefaultChannel, Options.ChannelTimeout, Options.Colors =
		"!", true, "general", time.Minute, map[string]string{"date": "blue", "user": "red", "channel": "yellow", "text": "white"}
	options, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(options, &Options)
	if err != nil {
		return err
	}
	return nil
}
