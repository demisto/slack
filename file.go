package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Comment holds information about a file comment
type Comment struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	User      string `json:"user"`
	Comment   string `json:"comment"`
	Created   int64  `json:"created,omitempty"`
}

// File holds information about a file
type File struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`

	Name       string `json:"name"`
	Title      string `json:"title"`
	Mimetype   string `json:"mimetype"`
	Filetype   string `json:"filetype"`
	PrettyType string `json:"pretty_type"`
	UserID     string `json:"user"`

	Mode         string `json:"mode"`
	Editable     bool   `json:"editable"`
	IsExternal   bool   `json:"is_external"`
	ExternalType string `json:"external_type"`

	Size int `json:"size"`

	URL                string `json:"url"`
	URLDownload        string `json:"url_download"`
	URLPrivate         string `json:"url_private"`
	URLPrivateDownload string `json:"url_private_download"`

	Thumb64     string `json:"thumb_64"`
	Thumb80     string `json:"thumb_80"`
	Thumb360    string `json:"thumb_360"`
	Thumb360Gif string `json:"thumb_360_gif"`
	Thumb360W   int    `json:"thumb_360_w"`
	Thumb360H   int    `json:"thumb_360_h"`

	Permalink        string `json:"permalink"`
	EditLink         string `json:"edit_link"`
	Preview          string `json:"preview"`
	PreviewHighlight string `json:"preview_highlight"`
	Lines            int    `json:"lines"`
	LinesMore        int    `json:"lines_more"`

	IsPublic        bool     `json:"is_public"`
	PublicURLShared bool     `json:"public_url_shared"`
	Channels        []string `json:"channels"`
	Groups          []string `json:"groups"`
	InitialComment  Comment  `json:"initial_comment"`
	NumStars        int      `json:"num_stars"`
	IsStarred       bool     `json:"is_starred"`
}

// FileUploadResponse is the response to the file upload
type FileUploadResponse struct {
	slackResponse
	File File `json:"file"`
}

// doUpload executes the API request for file upload
// Returns the response if the status code is between 200 and 299
func (s *Slack) doUpload(path, filename string, params url.Values, data io.Reader, result interface{}) error {
	appendNotEmpty("token", s.token, params)
	var t time.Time
	if s.tracelog != nil {
		t = time.Now()
		s.tracef("Start request %s at %v", path, t)
	}
	// Pipe the file so as not to read it into memory
	bodyReader, bodyWriter := io.Pipe()
	// create a multipat/mime writer
	writer := multipart.NewWriter(bodyWriter)
	// get the Content-Type of our form data
	fdct := writer.FormDataContentType()
	// Read file errors from the channel
	errChan := make(chan error, 1)
	go func() {
		defer bodyWriter.Close()
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			errChan <- err
			return
		}
		if _, err := io.Copy(part, data); err != nil {
			errChan <- err
			return
		}
		for k, v := range params {
			if err := writer.WriteField(k, v[0]); err != nil {
				errChan <- err
				return
			}
		}
		errChan <- writer.Close()
	}()

	// create a HTTP request with our body, that contains our file
	postReq, err := http.NewRequest("POST", s.url+path, bodyReader)
	if err != nil {
		return err
	}
	// add the Content-Type we got earlier to the request header.
	postReq.Header.Add("Content-Type", fdct)

	s.dumpRequest(postReq)

	// send our request off, get response and/or error
	resp, err := s.c.Do(postReq)
	if cerr := <-errChan; cerr != nil {
		return cerr
	}

	if s.tracelog != nil {
		s.tracef("End request %s at %v - took %v", path, time.Now(), time.Since(t))
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err = s.handleError(resp); err != nil {
		return err
	}
	s.dumpResponse(resp)
	if result != nil {
		if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
		// Handle ok response parameter
		sm := result.(Response)
		if !sm.IsOK() {
			s.errorf("%s\n", sm.Error())
			return sm
		}
	}
	return nil
}

// Upload a file to Slack optionally sharing it on given channels
func (s *Slack) Upload(title, filetype, filename, initialComment string, channels []string, data io.Reader) (*FileUploadResponse, error) {
	if filename == "" {
		return nil, fmt.Errorf("You must specify the filename for the upload")
	}
	params := url.Values{}
	appendNotEmpty("title", title, params)
	appendNotEmpty("filetype", filetype, params)
	appendNotEmpty("filename", filename, params)
	appendNotEmpty("initial_comment", initialComment, params)
	if len(channels) > 0 {
		params.Set("channels", strings.Join(channels, ","))
	}
	r := &FileUploadResponse{}
	err := s.doUpload("files.upload", filename, params, data, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
