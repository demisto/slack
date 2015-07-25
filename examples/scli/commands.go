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

func findIM(id string) *slack.IM {
	for i := range info.IMS {
		if info.IMS[i].ID == id {
			return &info.IMS[i]
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
	// First, let's see if the given ch is actually already an ID
	name := channelName(ch)
	if name != "" {
		return ch
	}
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
		if strings.ToLower(findUser(info.IMS[i].User).Name) == strings.ToLower(ch) {
			return info.IMS[i].ID
		}
	}
	return ""
}

func userID(u string) string {
	// Check if this is user ID and not name
	if findUser(u) != nil {
		return u
	}
	for i := range info.Users {
		if strings.ToLower(info.Users[i].Name) == strings.ToLower(u) {
			return info.Users[i].ID
		}
	}
	return ""
}

func channelName(ch string) string {
	if ch == "" {
		return ""
	}
	switch ch[0] {
	case 'C':
		for i := range info.Channels {
			if info.Channels[i].ID == ch {
				return info.Channels[i].Name
			}
		}
	case 'G':
		for i := range info.Groups {
			if info.Groups[i].ID == ch {
				return info.Groups[i].Name
			}
		}
	case 'D':
		for i := range info.IMS {
			if info.IMS[i].ID == ch {
				return findUser(info.IMS[i].User).Name
			}
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

func handleC(cmd string, parts []string) {
	if len(parts) > 0 {
		if !switchChannel(parts[0]) {
			fmt.Printf("Unable to switch - %s not found\n", parts[0])
		} else {
			fmt.Printf("Switched to %s\n", parts[0])
			// If we have a message, post it as well...
			if len(parts) > 1 {
				go postMessage(strings.Join(parts[1:], " "))
			}
		}
	}
}

func handleArchive(cmd string, parts []string) {
	for _, ch := range parts {
		id := channelID(ch)
		if id == "" {
			fmt.Printf("%s not found\n", ch)
			continue
		}
		var r slack.Response
		var err error
		var action string
		if strings.Contains(cmd, "unarchive") {
			r, err = s.Unarchive(id)
			action = "unarchive"
		} else {
			r, err = s.Archive(id)
			action = "archive"
		}
		if err != nil {
			fmt.Printf("Unable to %s %s - %v\n", action, ch, err)
			break
		} else if !r.IsOK() {
			fmt.Printf("Unable to %s %s - %s\n", action, ch, r.Error())
		} else {
			fmt.Printf("%s %sd\n", action, ch)
		}
	}
}

func handleCreate(cmd string, parts []string) {
	for _, ch := range parts {
		var r slack.Response
		var err error
		if cmd == "c-create" {
			r, err = s.ChannelCreate(ch)
		} else if cmd == "g-create" {
			r, err = s.GroupCreate(ch)
		} else {
			r, err = s.GroupCreateChild(ch)
		}
		if err != nil {
			fmt.Printf("Unable to create %s - %v\n", ch, err)
			break
		} else if !r.IsOK() {
			fmt.Printf("Unable to create %s - %s\n", ch, r.Error())
		} else {
			if cmd == "c-create" {
				fmt.Printf("%s created\n", r.(*slack.ChannelResponse).Channel.Name)
			} else {
				fmt.Printf("%s created\n", r.(*slack.GroupResponse).Group.Name)
			}
		}
	}
}

func handleHistory(cmd string, parts []string) {
	id := currChannelID
	ch := channelName(id)
	if len(parts) > 0 {
		ch = parts[0]
		id = channelID(ch)
	}
	if id == "" {
		fmt.Printf("%s not found\n", ch)
	} else {
		latest, oldest, count := "", "", 0
		if len(parts) > 1 {
			for _, arg := range parts[1:] {
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
		} else if !r.IsOK() {
			fmt.Printf("Unable to retrieve history for %s - %s\n", ch, r.Error())
		} else {
			fmt.Printf("Latest %d messages for %s (has_more=%v)\n", len(r.Messages), ch, r.HasMore)
			for i := range r.Messages {
				fmt.Printf("%s [%s]: %s\n", r.Messages[i].Timestamp, findUser(r.Messages[i].User).Name, r.Messages[i].Text)
			}
		}
	}
}

func handleInfo(cmd string, parts []string) {
	if len(parts) == 0 {
		parts = []string{channelName(currChannelID)}
	}
	for _, ch := range parts {
		id := channelID(ch)
		if id == "" {
			fmt.Printf("%s not found\n", ch)
			continue
		}
		var r slack.Response
		var err error
		if id != "" {
			if id[0] == 'C' {
				r, err = s.ChannelInfo(id)
			} else if id[0] == 'G' {
				r, err = s.GroupInfo(id)
			} else {
				fmt.Printf("IM with %s has no info\n", channelName(ch))
				continue
			}
		}
		if err != nil {
			fmt.Printf("Unable to retrieve info for %s - %v\n", ch, err)
		} else if !r.IsOK() {
			fmt.Printf("Unable to retrieve info for %s - %s\n", ch, r.Error())
		} else {
			var b []byte
			if id[0] == 'C' {
				b, err = json.MarshalIndent(r.(*slack.ChannelResponse).Channel, "", "  ")
			} else {
				b, err = json.MarshalIndent(r.(*slack.GroupResponse).Group, "", "  ")
			}
			if err != nil {
				fmt.Printf("Unable to retrieve info for %s - %v\n", ch, err)
			} else {
				fmt.Println(string(b))
			}
		}
	}
}

func handleInviteKick(cmd string, parts []string) {
	if len(parts) == 0 {
		fmt.Println("Please specify the users")
	} else {
		id := channelID(parts[0])
		users := parts[1:]
		// Use current channel and list of users
		if id == "" {
			id = currChannelID
			users = parts
		}
		for _, u := range users {
			var r slack.Response
			var err error
			var msg, msgErr string
			switch cmd {
			case "c-invite":
				r, err = s.ChannelInvite(id, userID(u))
				msg = "User %s invited to channel %s\n"
				msgErr = "Unable to invite user %s to channel %s - %v\n"
			case "c-kick":
				r, err = s.Kick(id, userID(u))
				msg = "User %s kicked from channel %s\n"
				msgErr = "Unable to kick user %s from channel %s - %v\n"
			case "g-invite":
				r, err = s.GroupInvite(id, userID(u))
				msg = "User %s invited to group %s\n"
				msgErr = "Unable to invite user %s to group %s - %v\n"
			case "g-kick":
				r, err = s.Kick(id, userID(u))
				msg = "User %s kicked from group %s\n"
				msgErr = "Unable to kick user %s from group %s - %v\n"
			}
			if err != nil {
				fmt.Printf(msgErr, u, parts[0], err)
				break
			} else if !r.IsOK() {
				fmt.Printf(msgErr, u, parts[0], r.Error())
			} else {
				fmt.Printf(msg, u, parts[0])
			}
		}
	}
}

func handleJoinLeave(cmd string, parts []string) {
	for _, ch := range parts {
		id := channelID(ch)
		if id == "" {
			fmt.Printf("%s not found\n", ch)
			continue
		}
		var r slack.Response
		var err error
		var msg, msgErr string
		switch cmd {
		case "c-join":
			r, err = s.ChannelJoin(ch)
			msg = "Joined channel %s\n"
			msgErr = "Unable to join channel %s - %v\n"
		case "c-leave":
			r, err = s.Leave(id)
			msg = "Left channel %s\n"
			msgErr = "Unable to leave channel %s - %v\n"
		case "g-leave":
			r, err = s.Leave(id)
			msg = "Left group %s\n"
			msgErr = "Unable to leave group %s - %v\n"
		}
		if err != nil {
			fmt.Printf(msgErr, ch, err)
			break
		} else if !r.IsOK() {
			fmt.Printf(msgErr, ch, r.Error())
		} else {
			fmt.Printf(msg, ch)
		}
	}
}

func handleList(cmd string, parts []string) {
	var err error
	excludeArchived := true
	if len(parts) > 0 {
		excludeArchived, err = strconv.ParseBool(parts[0])
		if err != nil {
			excludeArchived = true
		}
	}
	if cmd == "list" && len(currChannelID) > 0 {
		cmd = string(strings.ToLower(currChannelID)[0]) + "-list"
	}
	var b []byte
	switch cmd {
	case "c-list":
		r, err := s.ChannelList(excludeArchived)
		if err != nil {
			fmt.Printf("Error listing channels - %v\n", err)
			return
		}
		b, err = json.MarshalIndent(r.Channels, "", "  ")
	case "g-list":
		r, err := s.GroupList(excludeArchived)
		if err != nil {
			fmt.Printf("Error listing groups - %v\n", err)
			return
		}
		b, err = json.MarshalIndent(r.Groups, "", "  ")
	case "d-list":
		r, err := s.IMList()
		if err != nil {
			fmt.Printf("Error listing IMs - %v\n", err)
			return
		}
		b, err = json.MarshalIndent(r.IMs, "", "  ")
	}
	if err != nil {
		fmt.Printf("Unable to retrieve list - %v\n", err)
	} else {
		fmt.Println(string(b))
	}
}

func handleRename(cmd string, parts []string) {
	id := currChannelID
	newName := parts[0]
	if len(parts) > 1 {
		id = channelID(parts[0])
		if id == "" {
			fmt.Printf("Unable to find %s\n", parts[0])
		}
		newName = parts[1]
	}
	name := channelName(id)
	r, err := s.Rename(id, newName)
	if err != nil {
		fmt.Printf("Error renaming %s - %v\n", name, err)
	} else if !r.IsOK() {
		fmt.Printf("Error renaming %s - %v\n", name, r.Error())
	} else {
		fmt.Printf("%s renamed to %s\n", name, r.Channel.Name)
	}
}

func handlePurposeTopic(cmd string, parts []string) {
	// TODO - handle errors
	if len(parts) == 0 {
		switch cmd {
		case "purpose", "c-purpose", "g-purpose":
			if currChannelID[0] == 'C' {
				fmt.Println(findChannel(currChannelID).Purpose.Value)
			} else if currChannelID[0] == 'G' {
				fmt.Println(findGroup(currChannelID).Purpose.Value)
			}
		case "topic", "c-topic", "g-topic":
			if currChannelID[0] == 'C' {
				fmt.Println(findChannel(currChannelID).Topic.Value)
			} else if currChannelID[0] == 'G' {
				fmt.Println(findGroup(currChannelID).Topic.Value)
			}
		}
		return
	}
	id := channelID(parts[0])
	var newThing string
	// If we are changing the current channel or group
	if id == "" {
		newThing = strings.Join(parts, " ")
		id = currChannelID
	} else {
		if len(parts) > 1 {
			newThing = strings.Join(parts[1:], " ")
		} else {
			fmt.Println("Please specify the new value")
			return
		}
	}
	var result string
	var err error
	var resp slack.Response
	switch cmd {
	case "purpose", "c-purpose", "g-purpose":
		var r *slack.PurposeResponse
		r, err = s.SetPurpose(id, newThing)
		resp = r
		if r != nil {
			result = r.Purpose
		}
	case "topic", "c-topic", "g-topic":
		var r *slack.TopicResponse
		r, err = s.SetTopic(currChannelID, strings.Join(parts, " "))
		resp = r
		if r != nil {
			result = r.Topic
		}
	}
	if err != nil {
		fmt.Printf("Error setting new value - %v\n", err)
	} else if !resp.IsOK() {
		fmt.Printf("Error setting new value - %v\n", resp.Error())
	} else {
		fmt.Printf("New value is - %s\n", result)
	}
}

func handleClose(cmd string, parts []string) {
	for _, ch := range parts {
		id := channelID(ch)
		if id == "" {
			fmt.Printf("%s not found\n", ch)
			continue
		}
		r, err := s.CloseGroupOrIM(id)
		if err != nil {
			fmt.Printf("Unable to close %s - %v\n", ch, err)
			break
		} else if !r.IsOK() {
			fmt.Printf("Unable to close %s - %s\n", ch, r.Error())
		} else if r.AlreadyClosed {
			fmt.Printf("%s was already closed\n", ch)
		} else {
			fmt.Printf("%s closed\n", ch)
		}
	}
}

func handleOpen(cmd string, parts []string) {
	for _, ch := range parts {
		id := channelID(ch)
		if id == "" {
			fmt.Printf("%s not found\n", ch)
			continue
		}
		r, err := s.OpenGroupOrIM(id)
		if err != nil {
			fmt.Printf("Unable to open %s - %v\n", ch, err)
			break
		} else if !r.IsOK() {
			fmt.Printf("Unable to open %s - %s\n", ch, r.Error())
		} else if r.AlreadyClosed {
			fmt.Printf("%s was already open\n", ch)
		} else {
			fmt.Printf("%s opened\n", ch)
		}
	}
}

func handleFileUpload(cmd string, parts []string) {
	//for _, f := range parts {

	//}
}

func handleCommand(line string) bool {
	parts := strings.Fields(line)
	cmd := strings.ToLower(parts[0][len(Options.CommandPrefix):])
	switch cmd {
	case "exit":
		return true
	case "c", "g", "d":
		handleC(cmd, parts[1:])
	case "c-archive", "g-archive", "c-unarchive", "g-unarchive":
		handleArchive(cmd, parts[1:])
	case "c-create", "g-create", "g-createChild":
		handleCreate(cmd, parts[1:])
	case "c-history", "g-history", "d-history", "hist":
		handleHistory(cmd, parts[1:])
	case "c-info", "g-info", "info":
		handleInfo(cmd, parts[1:])
	case "c-invite", "c-kick", "g-invite", "g-kick":
		handleInviteKick(cmd, parts[1:])
	case "c-join", "c-leave", "g-leave":
		handleJoinLeave(cmd, parts[1:])
	case "c-list", "g-list", "d-list", "list":
		handleList(cmd, parts[1:])
	case "c-rename", "g-rename":
		handleRename(cmd, parts[1:])
	case "c-purpose", "g-purpose", "c-topic", "g-topic", "purpose", "topic":
		handlePurposeTopic(cmd, parts[1:])
	case "g-close", "d-close":
		handleClose(cmd, parts[1:])
	case "g-open", "d-open":
		handleOpen(cmd, parts[1:])
	case "f":
		handleFileUpload(cmd, parts[1:])
	case "f-delete", "f-info", "f-list", "f-c":
	}
	return false
}
