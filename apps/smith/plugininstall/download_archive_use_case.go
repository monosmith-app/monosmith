package plugininstall

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MaxDownloadBytes caps one download: a plugin is a few binaries.
const MaxDownloadBytes int64 = 512 << 20

// NewClient is the HTTP client downloads use: a timeout, at most five
// redirects, and every redirect must satisfy CheckURL too, so an https link
// cannot be sent on to plain http.
func NewClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return CheckURL(req.URL.String())
		},
	}
}

// download copies the body of url into dst, refusing a body over limit bytes,
// and returns its length and SHA-256. A non-200 answer is an error.
func download(ctx context.Context, client *http.Client, rawURL, token string, limit int64, dst io.Writer) (int64, string, error) {
	if err := CheckURL(rawURL); err != nil {
		return 0, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("User-Agent", "smith")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("downloading %s: %w", rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("downloading %s: %s", rawURL, resp.Status)
	}
	if resp.ContentLength > limit {
		return 0, "", fmt.Errorf("%s is %d bytes; the download limit is %d", rawURL, resp.ContentLength, limit)
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(dst, h), io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return 0, "", fmt.Errorf("downloading %s: %w", rawURL, err)
	}
	if n > limit {
		return 0, "", fmt.Errorf("%s is larger than the download limit of %d bytes", rawURL, limit)
	}
	return n, hex.EncodeToString(h.Sum(nil)), nil
}
