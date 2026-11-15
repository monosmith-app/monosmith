package pluginadmit

import (
	"bufio"
	"io"
	"sync"
)

// Prompter is a person who can be asked whether to trust a plugin.
type Prompter struct {
	Out io.Writer
	In  *bufio.Reader
}

var (
	mu      sync.Mutex
	current *Prompter
)

// SetPrompter says that a person is at the terminal and may be asked, for the
// rest of the process. A command that has nobody to ask (smith serve, smith
// mcp, the shell) never sets one, and nothing is asked in CI.
func SetPrompter(p *Prompter) {
	mu.Lock()
	defer mu.Unlock()
	current = p
}

// CurrentPrompter is the prompter set for this process, nil when there is
// nobody to ask.
func CurrentPrompter() *Prompter {
	mu.Lock()
	defer mu.Unlock()
	return current
}
