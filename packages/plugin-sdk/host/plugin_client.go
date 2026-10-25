package host

import (
	"context"
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

// DefaultCancelGrace is how long the host waits for a plugin to answer after
// asking it to stop (or, for a plugin that cannot be asked, after the caller
// gave up) before it kills the process.
const DefaultCancelGrace = 5 * time.Second

// UnlimitedInFlight is how many requests the host keeps in flight to a plugin
// that declared unlimited concurrency.
const UnlimitedInFlight = 32

// Client is a connection to one initialized plugin. It never has more
// requests in flight than the plugin declared it can handle (one when it
// declared nothing), so with the default the calls are serialised: a call
// waits for the one before it to be answered.
type Client struct {
	Info   protocol.InitializeResult
	stream *transport.Stream
	closer io.Closer
	cmd    *exec.Cmd
	grace  time.Duration

	// slots has one token per request that may be in flight.
	slots chan struct{}

	mu       sync.Mutex
	nextID   int64
	pending  map[int64]chan outcome
	readErr  error
	inFlight int
	peak     int
}

// outcome is what the reader hands the call waiting for an id.
type outcome struct {
	resp protocol.Response
	err  error
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
	// CancelGrace is how long a plugin has to answer after it was asked to
	// stop, before the host kills it. Zero means DefaultCancelGrace.
	CancelGrace time.Duration
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
	c := newClient(stdout, stdin, stdin, opts)
	c.cmd = cmd
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
	c := newClient(r, w, w, opts)
	if err := c.initialize(opts); err != nil {
		return nil, err
	}
	return c, nil
}

func newClient(r io.Reader, w io.Writer, closer io.Closer, opts Options) *Client {
	grace := opts.CancelGrace
	if grace == 0 {
		grace = DefaultCancelGrace
	}
	c := &Client{
		stream: transport.NewStream(r, w), closer: closer, grace: grace,
		slots: make(chan struct{}, 1), pending: map[int64]chan outcome{},
	}
	go c.readResponses()
	return c
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
	c.slots = make(chan struct{}, InFlightLimit(c.Info))
	return nil
}

// InFlightLimit is how many requests the host keeps in flight to a plugin
// that answered initialize with info: its declared concurrency, one when it
// declared none, UnlimitedInFlight when it declared no limit.
func InFlightLimit(info protocol.InitializeResult) int {
	limit, unlimited := protocol.ConcurrencyLimit(info.Concurrency)
	if unlimited {
		return UnlimitedInFlight
	}
	return limit
}

// readResponses hands every response to the call waiting for its id. When the
// stream ends or breaks, every call still waiting fails with the reason.
func (c *Client) readResponses() {
	for {
		msg, err := c.stream.Read()
		if err == nil {
			var resp protocol.Response
			if err = json.Unmarshal(msg, &resp); err != nil {
				err = fmt.Errorf("malformed response: %w", err)
			} else {
				c.deliver(resp)
				continue
			}
		}
		c.mu.Lock()
		c.readErr = err
		waiting := c.pending
		c.pending = map[int64]chan outcome{}
		c.mu.Unlock()
		for _, done := range waiting {
			done <- outcome{err: err}
		}
		return
	}
}

func (c *Client) deliver(resp protocol.Response) {
	if resp.ID == nil {
		return // an error the plugin could not attribute to a request
	}
	c.mu.Lock()
	done, ok := c.pending[*resp.ID]
	delete(c.pending, *resp.ID)
	c.mu.Unlock()
	if ok {
		done <- outcome{resp: resp}
	}
}

// Call sends one request and decodes the result into result (which may be nil).
// A timeout of zero means DefaultTimeout. It is CallContext without a context.
func (c *Client) Call(method string, params, result any, timeout time.Duration) error {
	return c.CallContext(context.Background(), method, params, result, timeout)
}

// CallContext sends one request and decodes the result into result (which may
// be nil). It waits for a free place first, so a plugin never has more
// requests in flight than it declared; the timeout counts from the moment the
// request is sent. A timeout of zero means DefaultTimeout.
//
// When ctx is done or the timeout runs out, the host stops waiting: a plugin
// that advertised cancel is sent the cancel notification and has the grace
// period to answer, with an error or with the result it had already, and the
// answer is what the call returns. A plugin that does not answer in time is
// killed, because its stream can no longer be trusted, and so are all the
// other calls in flight to it. A plugin that cannot be asked to stop is killed
// at once when the time ran out, and after the grace period when ctx was
// cancelled. The error of a call given up on is context.Canceled or
// context.DeadlineExceeded (errors.Is), plus the plugin's answer when it gave one.
func (c *Client) CallContext(ctx context.Context, method string, params, result any, timeout time.Duration) error {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	if method != protocol.MethodShutdown { // shutdown must reach a plugin that is busy
		release, err := c.acquire(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", method, err)
		}
		defer release()
	}
	id, done, err := c.send(method, raw)
	if err != nil {
		return err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case o := <-done:
		return decode(method, o, result)
	case <-ctx.Done():
		return c.giveUp(method, id, done, result, ctx.Err(), timeout)
	case <-timer.C:
		return c.giveUp(method, id, done, result, context.DeadlineExceeded, timeout)
	}
}

// acquire waits for a place among the requests that may be in flight.
func (c *Client) acquire(ctx context.Context) (release func(), err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	slots := c.slots
	select {
	case slots <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	c.mu.Lock()
	c.inFlight++
	c.peak = max(c.peak, c.inFlight)
	c.mu.Unlock()
	return func() {
		c.mu.Lock()
		c.inFlight--
		c.mu.Unlock()
		<-slots
	}, nil
}

// CancelGrace is how long the plugin has to answer after it was asked to stop.
func (c *Client) CancelGrace() time.Duration { return c.grace }

// PeakInFlight is the most requests this client has had in flight at once.
func (c *Client) PeakInFlight() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.peak
}

func (c *Client) send(method string, params json.RawMessage) (int64, chan outcome, error) {
	c.mu.Lock()
	if c.readErr != nil {
		err := c.readErr
		c.mu.Unlock()
		return 0, nil, fmt.Errorf("%s: plugin stopped answering: %w", method, err)
	}
	c.nextID++
	id := c.nextID
	done := make(chan outcome, 1)
	c.pending[id] = done
	c.mu.Unlock()
	body, err := json.Marshal(protocol.Request{JSONRPC: "2.0", ID: &id, Method: method, Params: params})
	if err == nil {
		err = c.stream.Write(body)
	}
	if err != nil {
		c.forget(id)
		return 0, nil, fmt.Errorf("%s: plugin is not accepting requests: %w", method, err)
	}
	return id, done, nil
}

func (c *Client) forget(id int64) {
	c.mu.Lock()
	delete(c.pending, id)
	c.mu.Unlock()
}

// Notify sends a message that gets no answer, such as cancel.
func (c *Client) Notify(method string, params any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	body, err := json.Marshal(protocol.Request{JSONRPC: "2.0", Method: method, Params: raw})
	if err != nil {
		return err
	}
	return c.stream.Write(body)
}

func decode(method string, o outcome, result any) error {
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
}

// giveUp is the end of a call whose context is done or whose time ran out.
func (c *Client) giveUp(method string, id int64, done chan outcome, result any, cause error, timeout time.Duration) error {
	asked := c.Supports(protocol.MethodCancel)
	if asked {
		_ = c.Notify(protocol.MethodCancel, protocol.CancelParams{ID: id})
	}
	if asked || !errors.Is(cause, context.DeadlineExceeded) {
		grace := time.NewTimer(c.grace)
		defer grace.Stop()
		select {
		case o := <-done:
			return c.answeredAfterStopping(method, o, result, cause, timeout)
		case <-grace.C:
		}
	}
	c.forget(id)
	c.kill()
	if errors.Is(cause, context.DeadlineExceeded) {
		return &interruption{message: fmt.Sprintf("%s: plugin did not answer within %s", method, timeout), causes: []error{cause}, killed: true}
	}
	return &interruption{message: fmt.Sprintf("%s: cancelled; the plugin did not stop within %s and was killed", method, c.grace), causes: []error{cause}, killed: true}
}

// answeredAfterStopping is the end of a call the plugin answered while the
// host was giving up on it. A result is the work finished before the stop took
// effect and is returned as a success.
func (c *Client) answeredAfterStopping(method string, o outcome, result any, cause error, timeout time.Duration) error {
	err := decode(method, o, result)
	if err == nil {
		return nil
	}
	what := "cancelled"
	if errors.Is(cause, context.DeadlineExceeded) {
		what = fmt.Sprintf("did not answer within %s", timeout)
	}
	return &interruption{message: fmt.Sprintf("%s: %s: %v", method, what, err), causes: []error{cause, err}}
}

// Supports reports whether the plugin advertised a method.
func (c *Client) Supports(method string) bool {
	caps := c.Info.Capabilities
	switch method {
	case protocol.MethodCancel:
		return caps.Cancel
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

// kill ends the plugin: the process is killed, or, for a plugin reached
// through a reader and a writer, the connection is closed. Every call still in
// flight fails.
func (c *Client) kill() {
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		return
	}
	if c.cmd == nil {
		_ = c.closer.Close()
	}
}
