package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/demisto/goxforce"
	"github.com/demisto/slack"
	"github.com/slavikm/govt"
)

var token = flag.String("token", "", "token to connect to slack")
var vtToken = flag.String("vt", "", "API token to VirusTotal")

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func joinMap(m map[string]bool) string {
	res := ""
	for k, v := range m {
		if v {
			res += k + ","
		}
	}
	if len(res) > 0 {
		return res[0 : len(res)-1]
	}
	return res
}

func joinMapInt(m map[string]int) string {
	res := ""
	for k, v := range m {
		res += fmt.Sprintf("%s (%d),", k, v)
	}
	if len(res) > 0 {
		return res[0 : len(res)-1]
	}
	return res
}

func handleURL(message slack.Message, c *goxforce.Client, s *slack.Slack, vt *govt.Client) {
	start := strings.Index(message.Text, "<http")
	end := strings.Index(message.Text[start:], ">")
	if end > 0 {
		end = end + start
		filter := strings.Index(message.Text[start:end], "|")
		if filter > 0 {
			end = start + filter
		}
		url := message.Text[start+1 : end]
		fmt.Printf("URL found - %s\n", url)
		urlResp, urlRespErr := c.URL(url)
		xfeMessage := ""
		color := "good"
		if urlRespErr != nil {
			// Small hack - see if the URL was not found
			if strings.Contains(urlRespErr.Error(), "404") {
				xfeMessage = "URL reputation not found"
			} else {
				xfeMessage = urlRespErr.Error()
			}
			color = "warning"
		} else {
			xfeMessage = fmt.Sprintf("Categories: %s. Score: %v", joinMap(urlResp.Result.Cats), urlResp.Result.Score)
			if urlResp.Result.Score >= 5 {
				color = "danger"
			} else if urlResp.Result.Score >= 1 {
				color = "warning"
			}
		}
		// If there is a problem, ignore it - the fields are going to be empty
		mx := ""
		resolve, resolveErr := c.Resolve(url)
		if resolveErr == nil {
			for i := range resolve.MX {
				mx += fmt.Sprintf("%s (%d) ", resolve.MX[i].Exchange, resolve.MX[i].Priority)
			}
		}

		vtMessage := ""
		vtColor := "good"
		vtResp, err := vt.GetUrlReport(url)
		if err != nil {
			vtMessage = err.Error()
			vtColor = "warning"
		} else {
			if vtResp.ResponseCode != 1 {
				vtMessage = fmt.Sprintf("VT error %d (%s)", vtResp.ResponseCode, vtResp.VerboseMsg)
			} else {
				detected := 0
				for i := range vtResp.Scans {
					if vtResp.Scans[i].Detected {
						detected++
					}
				}
				if detected >= 5 {
					vtColor = "danger"
				} else if detected >= 1 {
					vtColor = "warning"
				}
				vtMessage = fmt.Sprintf("Scan Date: %s, Detected: %d, Total: %d", vtResp.ScanDate, detected, int(vtResp.Total))
			}
		}
		postMessage := &slack.PostMessageRequest{
			Channel:  message.Channel,
			Text:     "URL Reputation for " + url + "\t-\tPowered by <http://www.demisto.com|Demisto>",
			Username: "Watson",
			Attachments: []slack.Attachment{
				{
					Fallback:   xfeMessage,
					AuthorName: "IBM X-Force Exchange",
					Color:      color,
				},
				{
					Fallback:   vtMessage,
					AuthorName: "VirusTotal",
					Text:       vtMessage,
					Color:      vtColor,
				},
			},
		}
		if resolveErr == nil {
			postMessage.Attachments[0].Fields = []slack.AttachmentField{
				{Title: "A", Value: strings.Join(resolve.A, ","), Short: true},
				{Title: "AAAA", Value: strings.Join(resolve.AAAA, ","), Short: true},
				{Title: "TXT", Value: strings.Join(resolve.TXT, ","), Short: true},
				{Title: "MX", Value: mx, Short: true},
			}
		}
		if urlRespErr == nil {
			postMessage.Attachments[0].Fields = append(postMessage.Attachments[0].Fields,
				slack.AttachmentField{Title: "Categories", Value: joinMap(urlResp.Result.Cats), Short: true},
				slack.AttachmentField{Title: "Score", Value: fmt.Sprintf("%v", urlResp.Result.Score), Short: true})
		} else {
			postMessage.Attachments[0].Text = xfeMessage

		}
		msgReply, err := s.PostMessage(postMessage, false)
		check(err)
		fmt.Printf("%v\n", *msgReply)
	}
}

func handleIP(message slack.Message, ip string, c *goxforce.Client, s *slack.Slack, vt *govt.Client) {
	xfeMessage := ""
	color := "good"
	ipResp, ipRespErr := c.IPR(ip)
	if ipRespErr != nil {
		// Small hack - see if the URL was not found
		if strings.Contains(ipRespErr.Error(), "404") {
			xfeMessage = "IP reputation not found"
		} else {
			xfeMessage = ipRespErr.Error()
		}
		color = "warning"
	} else {
		xfeMessage = fmt.Sprintf("Categories: %s. Country: %s. Score: %v", joinMapInt(ipResp.Cats), ipResp.Geo["country"].(string), ipResp.Score)
		if ipResp.Score >= 5 {
			color = "danger"
		} else if ipResp.Score >= 1 {
			color = "warning"
		}
	}
	vtMessage := ""
	vtColor := "good"
	vtResp, err := vt.GetIpReport(ip)
	if err != nil {
		vtMessage = err.Error()
		vtColor = "warning"
	} else {
		if vtResp.ResponseCode != 1 {
			vtMessage = fmt.Sprintf("VT error %d (%s)", vtResp.ResponseCode, vtResp.VerboseMsg)
			vtColor = "warning"
		} else {
			detected := 0
			vtMessage = "Detected URLs:\n"
			for i := range vtResp.DetectedUrls {
				vtMessage += fmt.Sprintf("URL: %s, Detected: %d, Total: %d, Scan Date: %s\n",
					vtResp.DetectedUrls[i].Url, int(vtResp.DetectedUrls[i].Positives), int(vtResp.DetectedUrls[i].Total), vtResp.DetectedUrls[i].ScanDate)
				detected += int(vtResp.DetectedUrls[i].Positives)
			}
			if detected >= 10 {
				vtColor = "danger"
			} else if detected >= 5 {
				vtColor = "warning"
			}
		}
	}
	postMessage := &slack.PostMessageRequest{
		Channel:  message.Channel,
		Text:     "IP Reputation for " + ip + "\t-\tPowered by <http://www.demisto.com|Demisto>",
		Username: "Watson",
		Attachments: []slack.Attachment{
			{
				Fallback:   xfeMessage,
				AuthorName: "IBM X-Force Exchange",
				Color:      color,
			},
			{
				Fallback:   vtMessage,
				AuthorName: "VirusTotal",
				Text:       vtMessage,
				Color:      vtColor,
			},
		},
	}
	if ipRespErr == nil {
		postMessage.Attachments[0].Fields = []slack.AttachmentField{
			{Title: "Categories", Value: joinMapInt(ipResp.Cats), Short: true},
			{Title: "Country", Value: ipResp.Geo["country"].(string), Short: true},
			{Title: "Score", Value: fmt.Sprintf("%v", ipResp.Score), Short: true},
		}
	} else {
		postMessage.Attachments[0].Text = xfeMessage
	}
	msgReply, err := s.PostMessage(postMessage, false)
	check(err)
	fmt.Printf("%v\n", *msgReply)
}

func handleMD5(message slack.Message, md5 string, c *goxforce.Client, s *slack.Slack, vt *govt.Client) {
	xfeMessage := ""
	color := "good"
	md5Resp, md5RespErr := c.MalwareDetails(md5)
	if md5RespErr != nil {
		// Small hack - see if the URL was not found
		if strings.Contains(md5RespErr.Error(), "404") {
			xfeMessage = "File reputation not found"
		} else {
			xfeMessage = md5RespErr.Error()
		}
		color = "warning"
	} else {
		xfeMessage = fmt.Sprintf("Type: %s, Created: %s, Family: %s, MIME: %s",
			md5Resp.Malware.Type, md5Resp.Malware.Created.String(), strings.Join(md5Resp.Malware.Family, ","), md5Resp.Malware.MimeType)
		if len(md5Resp.Malware.Family) > 0 {
			color = "danger"
		}
	}

	vtMessage := ""
	vtColor := "good"
	vtResp, err := vt.GetFileReport(md5)
	if err != nil {
		vtMessage = err.Error()
		vtColor = "warning"
	} else {
		if vtResp.ResponseCode != 1 {
			vtMessage = fmt.Sprintf("VT error %d (%s)", vtResp.ResponseCode, vtResp.VerboseMsg)
			vtColor = "warning"
		} else {
			vtMessage = fmt.Sprintf("Scan Date %s, Positives: %d, Total: %d\n", vtResp.ScanDate, int(vtResp.Positives), int(vtResp.Total))
			if vtResp.Positives >= 5 {
				vtColor = "danger"
			} else if vtResp.Positives >= 1 {
				vtColor = "warning"
			}
		}
	}
	postMessage := &slack.PostMessageRequest{
		Channel:  message.Channel,
		Text:     "File Reputation for " + md5 + "\t-\tPowered by <http://www.demisto.com|Demisto>",
		Username: "Watson",
		Attachments: []slack.Attachment{
			{
				Fallback:   xfeMessage,
				AuthorName: "IBM X-Force Exchange",
				Color:      color,
			},
			{
				Fallback:   vtMessage,
				AuthorName: "VirusTotal",
				Text:       vtMessage,
				Color:      vtColor,
			},
		},
	}
	if md5RespErr == nil {
		postMessage.Attachments[0].Fields = []slack.AttachmentField{
			{Title: "Type", Value: md5Resp.Malware.Type, Short: true},
			{Title: "Created", Value: md5Resp.Malware.Created.String(), Short: true},
			{Title: "Family", Value: strings.Join(md5Resp.Malware.Family, ","), Short: true},
			{Title: "MIME Type", Value: md5Resp.Malware.MimeType, Short: true},
		}
	} else {
		postMessage.Attachments[0].Text = xfeMessage
	}
	msgReply, err := s.PostMessage(postMessage, false)
	check(err)
	fmt.Printf("%v\n", *msgReply)
}

func main() {
	flag.Parse()
	ipReg := regexp.MustCompile("\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b")
	md5Reg := regexp.MustCompile("\\b[a-fA-F\\d]{32}\\b")
	c, err := goxforce.New(goxforce.SetErrorLog(log.New(os.Stderr, "XFE:", log.Lshortfile)))
	check(err)
	s, err := slack.New(
		slack.SetToken(*token),
		slack.SetErrorLog(log.New(os.Stderr, "ERR:", log.Lshortfile)),
		slack.SetTraceLog(log.New(os.Stderr, "DEBUG:", log.Lshortfile)))
	check(err)
	vt, err := govt.New(govt.SetApikey(*vtToken), govt.SetErrorLog(log.New(os.Stderr, "VT:", log.Lshortfile)))
	check(err)
	in := make(chan slack.Message)
	_, err = s.RTMStart("http://example.com", in)
	check(err)
	for {
		select {
		case msg := <-in:
			fmt.Printf("%v\n", msg)
			switch msg.Type {
			case "message":
				fmt.Printf("%s\n", msg.Text)
				if msg.Subtype == "bot_message" {
					continue
				}
				if strings.Contains(msg.Text, "<http") {
					go handleURL(msg, c, s, vt)
				}
				if ip := ipReg.FindString(msg.Text); ip != "" {
					go handleIP(msg, ip, c, s, vt)
				}
				if md5 := md5Reg.FindString(msg.Text); md5 != "" {
					go handleMD5(msg, md5, c, s, vt)
				}
			}
		}
	}
}
