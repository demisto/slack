package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/demisto/slack"
)

func findChannel(id string) *slack.Channel {
	for i := range info.Channels {
		if info.Channels[i].ID == id {
			return &info.Channels[i]
		}
	}
	return nil
}

func findGroup(id string) *slack.Group {
	for i := range info.Groups {
		if info.Groups[i].ID == id {
			return &info.Groups[i]
		}
	}
	return nil
}

func findUser(id string) *slack.User {
	for i := range info.Users {
		if info.Users[i].ID == id {
			return &info.Users[i]
		}
	}
	return nil
}

func channelID(ch string) string {
	for i := range info.Channels {
		if strings.ToLower(info.Channels[i].Name) == strings.ToLower(ch) {
			return info.Channels[i].ID
		}
	}
	for i := range info.Groups {
		if strings.ToLower(info.Groups[i].Name) == strings.ToLower(ch) {
			return info.Groups[i].ID
		}
	}
	for i := range info.IMS {
		if strings.ToLower(info.IMS[i].Name) == strings.ToLower(ch) {
			return info.IMS[i].ID
		}
	}
	return ""
}

func userID(u string) string {
	for i := range info.Users {
		if strings.ToLower(info.Users[i].Name) == strings.ToLower(u) {
			return info.Users[i].ID
		}
	}
	return ""
}

func switchChannel(ch string) bool {
	id := channelID(ch)
	if id != "" {
		currChannelID = id
		return true
	}
	return false
}

func channelInfo(ch, id string) {
	r, err := s.ChannelInfo(id)
	if err != nil {
		fmt.Printf("Unable to retrieve info %s - %v\n", ch, err)
	} else if !r.IsOK() {
		fmt.Printf("Unable to retrieve info %s - %s\n", ch, r.Error())
	} else {
		b, err := json.MarshalIndent(r.Channel, "", "  ")
		if err != nil {
			fmt.Printf("Unable to retrieve info %s - %v\n", ch, err)
		} else {
			fmt.Println(string(b))
		}
	}
}

func handleCommand(line string) bool {
	parts := strings.Fields(line)
	cmd := strings.ToLower(parts[0][len(Options.CommandPrefix):])
	switch cmd {
	case "exit":
		return true
	case "c":
		if len(parts) > 1 {
			if !switchChannel(parts[1]) {
				fmt.Printf("Unable to switch channel - channel %s not found\n", parts[1])
			} else {
				fmt.Printf("Switched to channel %s\n", parts[1])
				// If we have a message, post it as well...
				if len(parts) > 2 {
					go postMessage(strings.Join(parts[2:], " "))
				}
			}
		}
	case "c-archive":
		if len(parts) > 1 {
			for _, ch := range parts[1:] {
				id := channelID(ch)
				if id == "" {
					fmt.Printf("Channel %s not found\n", ch)
					continue
				}
				r, err := s.ChannelArchive(id)
				if err != nil {
					fmt.Printf("Unable to archive %s - %v\n", ch, err)
					break
				} else if !r.IsOK() {
					fmt.Printf("Unable to archive %s - %s\n", ch, r.Error())
				} else {
					fmt.Printf("Channel %s archived\n", ch)
				}
			}
		}
	case "c-create":
		if len(parts) > 1 {
			for _, ch := range parts[1:] {
				r, err := s.ChannelCreate(ch)
				if err != nil {
					fmt.Printf("Unable to create %s - %v\n", ch, err)
					break
				} else if !r.IsOK() {
					fmt.Printf("Unable to create %s - %s\n", ch, r.Error())
				} else {
					fmt.Printf("Channel %s created\n", r.Channel.Name)
				}
			}
		}
	case "c-history":
		id := currChannelID
		ch := channelName(id)
		if len(parts) > 1 {
			ch = parts[1]
			id = channelID(ch)
		}
		if id == "" {
			fmt.Printf("Channel %s not found\n", ch)
		} else {
			latest, oldest, count := "", "", 0
			if len(parts) > 2 {
				for _, arg := range parts[2:] {
					if len(arg) < 5 {
						count, _ = strconv.Atoi(arg)
					} else {
						if latest == "" {
							latest = arg
						} else {
							oldest = arg
						}
					}
				}
			}
			r, err := s.History(id, latest, oldest, false, count)
			if err != nil {
				fmt.Printf("Unable to retrieve history for %s - %v\n", ch, err)
				break
			} else if !r.IsOK() {
				fmt.Printf("Unable to retrieve history for %s - %s\n", ch, r.Error())
			} else {
				fmt.Printf("Latest %d messages for %s (has_more=%v)\n", len(r.Messages), ch, r.HasMore)
				for i := range r.Messages {
					fmt.Printf("%s [%s]: %s\n", r.Messages[i].Timestamp, findUser(r.Messages[i].User).Name, r.Messages[i].Text)
				}
			}
		}
	case "c-info":
		if len(parts) > 1 {
			for _, ch := range parts[1:] {
				id := channelID(ch)
				if id == "" {
					fmt.Printf("Channel %s not found\n", ch)
					continue
				}
				channelInfo(ch, id)
			}
		} else {
			channelInfo(channelName(currChannelID), currChannelID)
		}
	case "c-invite":
		if len(parts) < 3 {
			fmt.Println("Please specify the channel and users to invite")
		} else {
			id := channelID(parts[1])
			users := parts[2:]
			// Use current channel and list of users
			if id == "" {
				id = currChannelID
				users = parts[1:]
			}
			for _, u := range users {
				r, err := s.ChannelInvite(id, userID(u))
				if err != nil {
					fmt.Printf("Unable to invite user %s to channel %s - %v\n", u, parts[1], err)
					break
				} else if !r.IsOK() {
					fmt.Printf("Unable to invite user %s to channel %s - %v\n", u, parts[1], r.Error())
				} else {
					fmt.Printf("User %s invited to channel %s\n", u, parts[1])
				}
			}
		}
	}
	return false
}
