package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	for _, name := range names {
		if endsSpace(line) {
			if !in(parts, name) {
				completions = append(completions, line+name)
			}
		} else if strings.HasPrefix(name, parts[l-1]) {
			if !in(parts[:l-1], name) {
				completions = append(completions, line+name[len(parts[l-1]):])
			}
		}
	}
	return completions
}

func channelCompletions(line string, parts []string) []string {
	var names []string
	for i := range info.Channels {
		names = append(names, info.Channels[i].Name)
	}
	return findCompletions(line, parts, names)
}

func groupCompletions(line string, parts []string) []string {
	var names []string
	for i := range info.Groups {
		names = append(names, info.Groups[i].Name)
	}
	return findCompletions(line, parts, names)
}

func imCompletions(line string, parts []string) []string {
	var names []string
	for i := range info.IMS {
		names = append(names, findUser(info.IMS[i].User).Name)
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

func fileCompletions(line string, parts []string) []string {
	var names []string
	for i := range files {
		names = append(names, files[i].Name)
	}
	return findCompletions(line, parts, names)
}

func osFileCompletions(line string, parts []string) []string {
	var completions []string
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working dir - %v", err)
		return completions
	}
	file := ""
	// if this is not a new file we are starting with
	if !endsSpace(line) || strings.HasSuffix(line, "\\ ") {
		lineCopy := strings.Replace(line, "\\ ", " ", -1)
		index := strings.Index(lineCopy, " ")
		for index != -1 {
			if line[index-1] != '\\' {
				lineCopy = lineCopy[index+1:]
				index = strings.Index(lineCopy, " ")
			} else {
				index = strings.Index(lineCopy[index+1:], " ")
			}
		}
		if lineCopy[0] == os.PathSeparator || len(lineCopy) > 2 && lineCopy[1] == ':' {
			dir, file = filepath.Split(lineCopy)
		} else {
			dir, file = filepath.Split(fmt.Sprintf("%s%c%s", dir, os.PathSeparator, lineCopy))
		}
	}
	fmt.Printf("%s %s\n", dir, file)
	dirFile, err := os.Open(dir)
	if err != nil {
		return completions
	}
	fi, err := dirFile.Readdir(-1)
	if err != nil {
		return completions
	}
	for i := range fi {
		if strings.HasPrefix(fi[i].Name(), file) {
			name := strings.Replace(fi[i].Name(), " ", "\\ ", -1)
			if fi[i].IsDir() {
				name = fmt.Sprintf("%s%c", name, os.PathSeparator)
			}
			completions = append(completions, line[:len(line)-len(file)]+name)
		}
	}
	return completions
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
		case "c-archive", "c-create", "c-info", "c-join", "c-leave", "c-unarchive":
			completions = channelCompletions(line, parts[1:])
		case "c", "c-history", "c-invite", "c-kick", "c-rename", "c-purpose", "c-topic":
			// Since if len is 1 then it has to end with space if we are here
			if l == 1 || l == 2 && !endsSpace(line) {
				completions = channelCompletions(line, parts[1:])
			} else if cmd == "c-invite" || cmd == "c-kick" {
				if l == 2 && endsWithSpace || l >= 3 && !endsWithSpace {
					completions = userCompletions(line, parts[2:])
				}
			}
		case "g-archive", "g-close", "g-create", "g-createChild", "g-info", "g-leave", "g-open", "g-unarchive":
			completions = groupCompletions(line, parts[1:])
		case "g", "g-history", "g-invite", "g-kick", "g-rename", "g-purpose", "g-topic":
			// Since if len is 1 then it has to end with space if we are here
			if l == 1 || l == 2 && !endsSpace(line) {
				completions = groupCompletions(line, parts[1:])
			} else if cmd == "g-invite" || cmd == "g-kick" {
				if l == 2 && endsWithSpace || l >= 3 && !endsWithSpace {
					completions = userCompletions(line, parts[2:])
				}
			}
		case "d-close", "d-open":
			completions = imCompletions(line, parts[1:])
		case "d", "d-history", "d-list":
			// Since if len is 1 then it has to end with space if we are here
			if l == 1 || l == 2 && !endsSpace(line) {
				completions = imCompletions(line, parts[1:])
			}
		case "f-delete", "f-info":
			completions = fileCompletions(line, parts[1:])
		case "f":
			completions = osFileCompletions(line, parts[1:])
		}
	}
	return completions
}
