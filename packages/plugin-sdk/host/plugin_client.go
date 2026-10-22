package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/transport"
)

// DefaultTimeout bounds one call. Generate and publish run real toolchains,
// so callers pass a longer timeout for them.
const DefaultTimeout = 2 * time.Minute

// Client is a connection to one initialized plugin. Calls are serialised.
type Client struct {
	Info   protocol.InitializeResult
	stream *transport.Stream
	closer io.Closer
	cmd    *exec.Cmd
	mu     sync.Mutex
	nextID int64
}

// Options configure Start.
type Options struct {
	// Command is the plugin's argv.
	Command       []string
	Env           []string
	WorkspaceRoot string
	// Stderr receives the plugin's log output.
	Stderr  io.Writer
	Host    protocol.Peer
	Timeout time.Duration
}

// Start launches a plugin process and initializes it.
func Start(opts Options) (*Client, error) {
	if len(opts.Command) == 0 {
		return nil, errors.New("plugin command is empty")
	}
	cmd := exec.Command(opts.Command[0], opts.Command[1:]...)
	cmd.Env = opts.Env
	cmd.Stderr = opts.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting plugin %v: %w", opts.Command, err)
	}
	c := &Client{stream: transport.NewStream(stdout, stdin), closer: stdin, cmd: cmd}
	if err := c.initialize(opts); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, err
	}
	return c, nil
}

// Connect initializes a plugin reachable through an existing reader and
// writer, for example one served in-process by serve.Serve over io.Pipe.
func Connect(r io.Reader, w io.WriteCloser, opts Options) (*Client, error) {
	c := &Client{stream: transport.NewStream(r, w), closer: w}
	if err := c.initialize(opts); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) initialize(opts Options) error {
	host := opts.Host
	if host.Name == "" {
		host = protocol.Peer{Name: "smith", Version: "dev"}
	}
	params := protocol.InitializeParams{ProtocolVersions: []int{protocol.Version}, Host: host, WorkspaceRoot: opts.WorkspaceRoot}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if err := c.Call(protocol.MethodInitialize, params, &c.Info, timeout); err != nil {
		return fmt.Errorf("initializing plugin: %w", err)
	}
	if c.Info.ProtocolVersion != protocol.Version {
		return fmt.Errorf("plugin %s chose protocol version %d, host speaks %d", c.Info.Plugin.Name, c.Info.ProtocolVersion, protocol.Version)
	}
	return nil
}

// Call sends one request and decodes the result into result (which may be nil).
// A timeout of zero means DefaultTimeout. On timeout the plugin process is
// killed, because its stream can no longer be trusted.
func (c *Client) Call(method string, params, result any, timeout time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	c.nextID++
	id := c.nextID
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body, err := json.Marshal(protocol.Request{JSONRPC: "2.0", ID: &id, Method: method, Params: raw})
	if err != nil {
		return err
	}
	if err := c.stream.Write(body); err != nil {
		return fmt.Errorf("%s: plugin is not accepting requests: %w", method, err)
	}
	type outcome struct {
		resp protocol.Response
		err  error
	}
	done := make(chan outcome, 1)
	go func() {
		for {
			msg, err := c.stream.Read()
			if err != nil {
				done <- outcome{err: err}
				return
			}
			var resp protocol.Response
			if err := json.Unmarshal(msg, &resp); err != nil {
				done <- outcome{err: fmt.Errorf("malformed response: %w", err)}
				return
			}
			if resp.ID != nil && *resp.ID == id {
				done <- outcome{resp: resp}
				return
			}
			// Responses to other ids cannot happen with serialised calls; skip.
		}
	}()
	select {
	case o := <-done:
		if o.err != nil {
			return fmt.Errorf("%s: plugin stopped answering: %w", method, o.err)
		}
		if o.resp.Error != nil {
			return o.resp.Error
		}
		if result != nil && len(o.resp.Result) > 0 {
			return json.Unmarshal(o.resp.Result, result)
		}
		return nil
	case <-time.After(timeout):
		c.kill()
		return fmt.Errorf("%s: plugin did not answer within %s", method, timeout)
	}
}

// Supports reports whether the plugin advertised a method.
func (c *Client) Supports(method string) bool {
	caps := c.Info.Capabilities
	switch method {
	case protocol.MethodDiscover:
		return caps.Discover
	case protocol.MethodGenerate:
		return caps.Generate
	case protocol.MethodVersion:
		return caps.Version
	case protocol.MethodPublish:
		return caps.Publish
	case protocol.MethodCI:
		return caps.CI
	case protocol.MethodDoctor:
		return caps.Doctor
	case protocol.MethodContribute:
		return caps.Contribute
	case protocol.MethodRemove:
		return caps.Remove
	case protocol.MethodPublished:
		return caps.Published
	case protocol.MethodInstall, protocol.MethodUninstall:
		return caps.Dependencies
	case protocol.MethodDeclared, protocol.MethodRewrite:
		return caps.Declarations
	case protocol.MethodLatest:
		return caps.Latest
	case protocol.MethodComponents:
		return caps.Components
	}
	return method == protocol.MethodInitialize || method == protocol.MethodShutdown
}

// Close asks the plugin to shut down and waits briefly for it to exit.
func (c *Client) Close() error {
	_ = c.Call(protocol.MethodShutdown, struct{}{}, nil, 5*time.Second)
	_ = c.closer.Close()
	if c.cmd == nil {
		return nil
	}
	exited := make(chan error, 1)
	go func() { exited <- c.cmd.Wait() }()
	select {
	case err := <-exited:
		return err
	case <-time.After(5 * time.Second):
		c.kill()
		return errors.New("plugin did not exit after shutdown; killed")
	}
}

func (c *Client) kill() {
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
}
