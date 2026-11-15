package pypiwheels

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

// DefaultUploadURL is PyPI's upload endpoint (the API twine uses).
const DefaultUploadURL = "https://upload.pypi.org/legacy/"

// ErrExists is returned for a file the index already has.
var ErrExists = errors.New("the index already has this file")

// Uploader sends wheels to a Python package index.
type Uploader struct {
	URL                string
	Username, Password string
	Client             *http.Client
}

// UploaderFromEnvironment reads the index and credentials the way twine
// does: TWINE_REPOSITORY_URL (default PyPI), and PYPI_TOKEN as an API token
// or TWINE_USERNAME and TWINE_PASSWORD.
func UploaderFromEnvironment() (Uploader, error) {
	u := Uploader{URL: os.Getenv("TWINE_REPOSITORY_URL"), Client: &http.Client{Timeout: 10 * time.Minute}}
	if u.URL == "" {
		u.URL = DefaultUploadURL
	}
	switch {
	case os.Getenv("TWINE_PASSWORD") != "":
		u.Username, u.Password = os.Getenv("TWINE_USERNAME"), os.Getenv("TWINE_PASSWORD")
		if u.Username == "" {
			u.Username = "__token__"
		}
	case os.Getenv("PYPI_TOKEN") != "":
		u.Username, u.Password = "__token__", os.Getenv("PYPI_TOKEN")
	default:
		return u, errors.New("set PYPI_TOKEN (or TWINE_USERNAME and TWINE_PASSWORD) to upload wheels")
	}
	return u, nil
}

// Upload sends one wheel with its metadata. A file the index already has is
// ErrExists, so a re-run completes a release.
func (u Uploader) Upload(w Wheel) error {
	content, err := os.ReadFile(w.File)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(content)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fields := [][2]string{
		{":action", "file_upload"}, {"protocol_version", "1"},
		{"metadata_version", "2.1"}, {"name", w.Metadata.Name}, {"version", w.Metadata.Version},
		{"filetype", "bdist_wheel"}, {"pyversion", "py3"},
		{"summary", w.Metadata.Summary}, {"home_page", w.Metadata.Homepage}, {"license", w.Metadata.License},
		{"description", w.Metadata.Description}, {"description_content_type", "text/markdown"},
		{"sha256_digest", hex.EncodeToString(sum[:])},
	}
	for _, f := range fields {
		if f[1] == "" {
			continue
		}
		if err := mw.WriteField(f[0], f[1]); err != nil {
			return err
		}
	}
	part, err := mw.CreateFormFile("content", filepath.Base(w.File))
	if err != nil {
		return err
	}
	if _, err := part.Write(content); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, u.URL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.SetBasicAuth(u.Username, u.Password)
	resp, err := u.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	reply, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode/100 == 2 {
		return nil
	}
	// PyPI answers a duplicate with 400 "File already exists"; other indexes
	// with 409.
	if resp.StatusCode == http.StatusConflict || strings.Contains(strings.ToLower(resp.Status+string(reply)), "already exist") {
		return ErrExists
	}
	return fmt.Errorf("uploading %s: %s: %s", filepath.Base(w.File), resp.Status, strings.TrimSpace(string(reply)))
}
