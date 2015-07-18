package main

import (
	"reflect"
	"strings"
)

var commands = []string{
	"c", "c-archive", "c-create", "c-history", "c-info", "c-invite", "c-join", "c-kick", "c-leave", "c-list", "c-rename", "c-purpose", "c-topic", "c-unarchive",
	"g", "g-archive", "g-close", "g-create", "g-createChild", "g-history", "g-info", "g-invite", "g-kick", "g-leave", "g-list", "g-open", "g-rename", "g-purpose", "g-topic", "g-unarchive",
	"d", "d-close", "d-history", "d-list", "d-open",
	"f-delete", "f-info", "f-list", "f",
	"r-add", "r-get", "r-list", "r-remove",
	"s", "s-files", "s-messages",
	"t-info", "t-logs",
	"u-presence", "u-info", "u-list",
}

func endsSpace(s string) bool {
	return strings.HasSuffix(s, " ")
}

// toIntf converts a slice or array of a specific type to array of interface{}
func toIntf(s interface{}) []interface{} {
	v := reflect.ValueOf(s)
	// There is no need to check, we want to panic if it's not slice or array
	intf := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		intf[i] = v.Index(i).Interface()
	}
	return intf
}

// in checks if val is in s slice
func in(s interface{}, val interface{}) bool {
	si := toIntf(s)
	for _, v := range si {
		if v == val {
			return true
		}
	}
	return false
}

func findCompletions(line string, parts, names []string) []string {
	var completions []string
	l := len(parts)
	part := parts[l-1]
	for _, name := range names {
		if endsSpace(line) {
			if !in(parts, name) {
				completions = append(completions, line+name)
			}
		} else if strings.HasPrefix(name, part) {
			if !in(parts[:l-1], name) {
				completions = append(completions, line+name[len(part):])
			}
		}
	}
	return completions
}

func channelCompetions(line string, parts []string) []string {
	var names []string
	for i := range info.Channels {
		names = append(names, info.Channels[i].Name)
	}
	return findCompletions(line, parts, names)
}

func userCompletions(line string, parts []string) []string {
	var names []string
	for i := range info.Users {
		names = append(names, info.Users[i].Name)
	}
	return findCompletions(line, parts, names)
}

func completer(line string) []string {
	var completions []string
	if !strings.HasPrefix(line, Options.CommandPrefix) {
		return completions
	}
	parts := strings.Fields(line)
	l := len(parts)
	endsWithSpace := endsSpace(line)
	// we are trying to complete command
	if l == 1 && !endsWithSpace {
		for _, c := range commands {
			cmd := Options.CommandPrefix + c
			if strings.HasPrefix(cmd, parts[0]) {
				completions = append(completions, cmd)
			}
		}
	} else {
		cmd := strings.ToLower(parts[0][len(Options.CommandPrefix):])
		switch cmd {
		case "c-archive", "c-info", "c-join", "c-leave", "c-unarchive":
			completions = channelCompetions(line, parts)
		case "c", "c-history", "c-invite", "c-kick", "c-rename", "c-purpose", "c-topic":
			// Since it has to end with space if we are here
			if l == 1 || l == 2 && !endsSpace(line) {
				completions = channelCompetions(line, parts[1:])
			} else if cmd == "c-invite" || cmd == "c-kick" {
				if l == 2 && endsWithSpace || l >= 3 && !endsWithSpace {
					completions = userCompletions(line, parts[2:])
				}
			}
		}
	}
	return completions
}
