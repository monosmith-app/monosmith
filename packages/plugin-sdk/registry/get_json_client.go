package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Timeout is how long GetJSON waits for one registry answer.
const Timeout = 10 * time.Second

// GetJSON reads the JSON document at url and decodes it into v.
//
// It returns an error when the request fails, when the status is not 200 (the
// error names the URL and the status), or when the body is not valid JSON for v.
func GetJSON(url string, v any) error {
	resp, err := (&http.Client{Timeout: Timeout}).Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
