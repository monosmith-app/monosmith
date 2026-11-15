package nugettool

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultPushURL is nuget.org's push endpoint (the one `dotnet nuget push`
// uses).
const DefaultPushURL = "https://www.nuget.org/api/v2/package"

// clientVersion is the NuGet client version the push claims, the one the
// .NET 10 SDK ships.
const clientVersion = "7.0.0"

// ErrExists is returned for a package version the feed already has.
var ErrExists = errors.New("the feed already has this version")

// Pusher sends packages to a NuGet feed.
type Pusher struct {
	URL    string
	APIKey string
	Client *http.Client
}

// PusherFromEnvironment reads the API key from NUGET_API_KEY and the feed
// from NUGET_PUSH_URL (default nuget.org).
func PusherFromEnvironment() (Pusher, error) {
	p := Pusher{URL: os.Getenv("NUGET_PUSH_URL"), APIKey: os.Getenv("NUGET_API_KEY"), Client: &http.Client{Timeout: 10 * time.Minute}}
	if p.URL == "" {
		p.URL = DefaultPushURL
	}
	if p.APIKey == "" {
		return p, errors.New("set NUGET_API_KEY to push the .NET tool packages")
	}
	return p, nil
}

// Push sends one package. A version the feed already has is ErrExists, so a
// re-run completes a release.
func (p Pusher) Push(pkg Package) error {
	content, err := os.ReadFile(pkg.File)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("package", filepath.Base(pkg.File))
	if err != nil {
		return err
	}
	if _, err := part.Write(content); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, p.URL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-NuGet-ApiKey", p.APIKey)
	// nuget.org refuses a push from a client older than 4.1.0.
	req.Header.Set("X-NuGet-Client-Version", clientVersion)
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	reply, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	switch {
	case resp.StatusCode/100 == 2:
		return nil
	case resp.StatusCode == http.StatusConflict:
		return ErrExists
	}
	return fmt.Errorf("pushing %s: %s: %s", filepath.Base(pkg.File), resp.Status, strings.TrimSpace(string(reply)))
}
