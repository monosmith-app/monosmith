package recipe

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// maxRecipeBytes bounds a downloaded recipe; real ones are a few kilobytes.
const maxRecipeBytes = 1 << 20

// fetch downloads a recipe over https. A nil client uses a default with a
// timeout.
func fetch(client *http.Client, url string) ([]byte, error) {
	if client == nil {
		client = &http.Client{Timeout: time.Minute}
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching the recipe at %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching the recipe at %s: %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxRecipeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("fetching the recipe at %s: %w", url, err)
	}
	if len(data) > maxRecipeBytes {
		return nil, fmt.Errorf("the recipe at %s is larger than %d bytes", url, maxRecipeBytes)
	}
	return data, nil
}
