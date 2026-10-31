package releaserehearsal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Answer is a registry's reply to one request.
type Answer struct {
	Status    int
	Challenge Challenge
	Body      []byte
}

// Probe sends the requests of a rehearsal. Every request reads, except the
// PyPI one, which is an upload with no file that the index refuses after it
// has checked the credential, so nothing is ever published.
type Probe struct {
	Client *http.Client
}

// NewProbe makes a probe with a 20 second timeout per request.
func NewProbe() *Probe {
	return &Probe{Client: &http.Client{Timeout: 20 * time.Second}}
}

// Get asks for url with an optional Authorization header value.
func (p *Probe) Get(url, authorization string) (Answer, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Answer{}, err
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	return p.do(req)
}

// EmptyUpload posts a PyPI upload form that has no file, with Basic sign-in.
func (p *Probe) EmptyUpload(url, user, password string) (Answer, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField(":action", "file_upload")
	_ = mw.WriteField("protocol_version", "1")
	_ = mw.Close()
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return Answer{}, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.SetBasicAuth(user, password)
	return p.do(req)
}

func (p *Probe) do(req *http.Request) (Answer, error) {
	resp, err := p.Client.Do(req)
	if err != nil {
		return Answer{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return Answer{Status: resp.StatusCode, Challenge: ParseChallenge(resp.Header.Get("WWW-Authenticate")), Body: body}, nil
}

// Authorization builds the header value for a token: Basic when the feed
// challenged with Basic (Azure Artifacts), else Bearer.
func Authorization(c Challenge, token string) string {
	if strings.EqualFold(c.Scheme, "basic") {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte("smith:"+token))
	}
	return "Bearer " + token
}

// userOf reads the user name from an npm `whoami` reply.
func userOf(body []byte) string {
	var who struct {
		Username string `json:"username"`
	}
	_ = json.Unmarshal(body, &who)
	return who.Username
}
