/*
Package slack is a library implementing the Slack Web and RTM API.

Written by Slavik Markovich at Demisto
*/
package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// DefaultURL points to the default Slack API
	DefaultURL = "https://slack.com/api/"
)

// Error is returned when there is a known condition error in the API
type Error struct {
	ID     string `json:"id"`
	Detail string `json:"detail"`
}

func newError(code, format string, args ...interface{}) *Error {
	return &Error{code, fmt.Sprintf(format, args...)}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.ID, e.Detail)
}

var (
	// ErrBadToken is returned when a bad token is passed to the API
	ErrBadToken = &Error{"bad_token", "Bad token was provided to the API"}
	// ErrNoToken is returned when the token is missing
	ErrNoToken = &Error{"no_token", "You must provide a Slack token to use the API"}
	// ErrBadOAuth is returned when OAuth credentials are bad
	ErrBadOAuth = &Error{"bad_oauth", "Bad OAuth credentuals provided"}
)

// Slack is the client to the Slack API.
type Slack struct {
	token        string          // The token to use for requests. Required.
	url          string          // The URL for the API.
	errorlog     *log.Logger     // Optional logger to write errors to
	tracelog     *log.Logger     // Optional logger to write trace and debug data to
	c            *http.Client    // The client to use for requests
	clientID     string          // OAuth
	clientSecret string          // OAuth
	code         string          // OAuth
	redirectURI  string          // OAuth
	ws           *websocket.Conn // WS connection
	mid          int             // WS message ID
	mutex        sync.Mutex      // WS mutex to protect changes
}

// OptionFunc is a function that configures a Client.
// It is used in New
type OptionFunc func(*Slack) error

// errorf logs to the error log.
func (s *Slack) errorf(format string, args ...interface{}) {
	if s.errorlog != nil {
		s.errorlog.Printf(format, args...)
	}
}

// tracef logs to the trace log.
func (s *Slack) tracef(format string, args ...interface{}) {
	if s.tracelog != nil {
		s.tracelog.Printf(format, args...)
	}
}

// New creates a new Slack client.
//
// The caller can configure the new client by passing configuration options to the func.
//
// Example:
//
//   s, err := slack.New(
//     slack.SetURL("https://some.url.com:port/"),
//     slack.SetErrorLog(log.New(os.Stderr, "Slack: ", log.Lshortfile),
//     slack.SetToken("Your-Token"))
//
// You must provide either a token or the OAuth parameters to retrieve the token.
// If no URL is configured, Slack uses DefaultURL by default.
//
// If no HttpClient is configured, then http.DefaultClient is used.
// You can use your own http.Client with some http.Transport for advanced scenarios.
//
// An error is also returned when some configuration option is invalid.
func New(options ...OptionFunc) (*Slack, error) {
	// Set up the client
	s := &Slack{
		url: "",
		c:   http.DefaultClient,
	}

	// Run the options on it
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}
	if s.url == "" {
		s.url = DefaultURL
	}
	s.tracef("Using URL [%s]\n", s.url)

	// If no API key was specified and no OAuth details
	if s.token == "" && s.clientID == "" {
		s.errorf("%s\n", ErrNoToken.Error())
		return nil, ErrNoToken
	}

	// Let's get OAuth token
	if s.token == "" {
		r, err := s.OAuthAccess()
		if err != nil {
			return nil, err
		}
		s.token = r.AccessToken
	}

	return s, nil
}

// Initialization functions

// SetToken sets the Slack API token to use
func SetToken(token string) OptionFunc {
	return func(s *Slack) error {
		if token == "" {
			s.errorf("%s\n", ErrBadToken.Error())
			return ErrBadToken
		}
		s.token = token
		return nil
	}
}

// SetOAuthCredentials provides the OAuth details to convert OAuth to token
func SetOAuthCredentials(clientID, clientSecret, code, redirectURI string) OptionFunc {
	return func(s *Slack) error {
		if clientID == "" || clientSecret == "" || code == "" {
			s.errorf("%s\n", ErrBadOAuth.Error())
			return ErrBadOAuth
		}
		s.clientID, s.clientSecret, s.code, s.redirectURI = clientID, clientSecret, code, redirectURI
		return nil
	}
}

// SetHTTPClient can be used to specify the http.Client to use when making
// requests to Slack.
func SetHTTPClient(httpClient *http.Client) OptionFunc {
	return func(s *Slack) error {
		if httpClient != nil {
			s.c = httpClient
		} else {
			s.c = http.DefaultClient
		}
		return nil
	}
}

// SetURL defines the URL endpoint for Slack
func SetURL(rawurl string) OptionFunc {
	return func(s *Slack) error {
		if rawurl == "" {
			rawurl = DefaultURL
		}
		u, err := url.Parse(rawurl)
		if err != nil {
			e := newError("bad_url", "Invalid URL [%s] - %v\n", rawurl, err.Error())
			s.errorf("%s\n", e.Error())
			return e
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			err = newError("bad_url", "Invalid schema specified [%s]", rawurl)
			s.errorf("%v", err)
			return err
		}
		s.url = rawurl
		if !strings.HasSuffix(s.url, "/") {
			s.url += "/"
		}
		return nil
	}
}

// SetErrorLog sets the logger for critical messages. It is nil by default.
func SetErrorLog(logger *log.Logger) func(*Slack) error {
	return func(s *Slack) error {
		s.errorlog = logger
		return nil
	}
}

// SetTraceLog specifies the logger to use for output of trace messages like
// HTTP requests and responses. It is nil by default.
func SetTraceLog(logger *log.Logger) func(*Slack) error {
	return func(s *Slack) error {
		s.tracelog = logger
		return nil
	}
}

// dumpRequest dumps a request to the debug logger if it was defined
func (s *Slack) dumpRequest(req *http.Request) {
	if s.tracelog != nil {
		out, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			s.tracef("%s\n", string(out))
		}
	}
}

// dumpResponse dumps a response to the debug logger if it was defined
func (s *Slack) dumpResponse(resp *http.Response) {
	if s.tracelog != nil {
		out, err := httputil.DumpResponse(resp, true)
		if err == nil {
			s.tracef("%s\n", string(out))
		}
	}
}

// Request handling functions

// handleError will handle responses with status code different from success
func (s *Slack) handleError(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if s.errorlog != nil {
			out, err := httputil.DumpResponse(resp, true)
			if err == nil {
				s.errorf("%s\n", string(out))
			}
		}
		e := newError("http_error", "Unexpected status code: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
		s.errorf("%s\n", e.Error())
		return e
	}
	return nil
}

// do executes the API request.
// Returns the response if the status code is between 200 and 299
func (s *Slack) do(path string, params url.Values, result interface{}) error {
	appendNotEmpty("token", s.token, params)
	var t time.Time
	if s.tracelog != nil {
		t = time.Now()
		s.tracef("Start request %s at %v", path, t)
	}
	resp, err := s.c.PostForm(s.url+path, params)
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
		switch result.(type) {
		// Should we just dump the response body
		case io.Writer:
			if _, err = io.Copy(result.(io.Writer), resp.Body); err != nil {
				return err
			}
		case Response:
			if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
				return err
			}
			// Handle ok response parameter
			sm := result.(Response)
			if !sm.IsOK() {
				s.errorf("%s\n", sm.Error())
				return sm
			}
		default:
			// Try parsing the message anyway
			if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
				return err
			}
		}
	}
	return nil
}

// Helper functions

func appendNotEmpty(name, val string, params url.Values) {
	if val != "" {
		params.Add(name, val)
	}
}

// Slack API types

// Response interface represents any reply from Slack with the basic ok, error methods
type Response interface {
	IsOK() bool
	Error() string
}

// Common response to all messages
type slackResponse struct {
	OK  bool   `json:"ok"`
	Err string `json:"error"`
}

func (r *slackResponse) IsOK() bool {
	return r.OK
}

func (r *slackResponse) Error() string {
	return r.Err
}
