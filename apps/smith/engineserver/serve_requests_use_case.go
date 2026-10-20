package engineserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sort"
	"sync"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
)

// Server answers JSON-RPC requests for one workspace.
type Server struct {
	Version string
	// Dir is the workspace directory; empty means the current one.
	Dir string
	// RunTasks runs a request; nil means taskrun.RunInDirectory. A test
	// replaces it to run a graph of its own.
	RunTasks func(ctx context.Context, req taskrun.Request) (taskrun.Summary, error)
	// Execute runs smith for the methods that stand for a command; nil means
	// ExecuteSelf.
	Execute Execute

	runs runRegistry
	wmu  sync.Mutex
	out  io.Writer
}

// Serve reads requests from in, one JSON message per line, and answers on out
// until in ends or the client sends shutdown. Requests run concurrently, so a
// cancel is heard while a run is in flight; runs still in flight when the
// client goes away are cancelled, which stops their processes.
func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	s.out = out
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	var wg sync.WaitGroup
	defer wg.Wait()
	defer s.runs.cancelAll()

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1<<20), 16<<20)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			s.send(response{JSONRPC: "2.0", ID: json.RawMessage("null"), Error: &rpcError{codeParse, "parse error: " + err.Error()}})
			continue
		}
		if req.Method == "shutdown" {
			s.runs.cancelAll()
			wg.Wait()
			s.reply(req, nil, nil)
			return nil
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := s.dispatch(ctx, req)
			s.reply(req, result, err)
		}()
	}
	return scanner.Err()
}

// reply answers a request; a notification (no id) gets no answer.
func (s *Server) reply(req request, result any, err *rpcError) {
	if len(req.ID) == 0 {
		return
	}
	if err != nil {
		s.send(response{JSONRPC: "2.0", ID: req.ID, Error: err})
		return
	}
	if result == nil {
		result = struct{}{}
	}
	s.send(response{JSONRPC: "2.0", ID: req.ID, Result: result})
}

func (s *Server) send(v any) {
	line, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.wmu.Lock()
	defer s.wmu.Unlock()
	_, _ = s.out.Write(append(line, '\n'))
}

// sink sends run events to the client as notifications.
func (s *Server) sink() runevents.Sink {
	return runevents.SinkFunc(func(e runevents.Event) {
		s.send(notification{JSONRPC: "2.0", Method: EventNotification, Params: e})
	})
}

func (s *Server) dispatch(ctx context.Context, req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		return s.initialize(), nil
	case "ping":
		return struct{}{}, nil
	case "run":
		return s.run(ctx, req.Params)
	case "cancel":
		var p CancelParams
		if err := decodeParams(req.Params, &p); err != nil || p.Run == "" {
			return nil, &rpcError{codeInvalidParams, "cancel needs the run to cancel: {\"run\": \"...\"}"}
		}
		return map[string]bool{"cancelled": s.runs.cancel(p.Run)}, nil
	}
	if tool, ok := queryMethods[req.Method]; ok {
		return s.query(ctx, tool, req.Params)
	}
	if tool, ok := commandMethods[req.Method]; ok {
		return s.command(ctx, req.Method, tool, req.Params)
	}
	return nil, &rpcError{codeMethodNotFound, "method not found: " + req.Method}
}

func (s *Server) initialize() InitializeResult {
	methods := []string{"initialize", "ping", "run", "cancel", "shutdown"}
	for m := range queryMethods {
		methods = append(methods, m)
	}
	for m := range commandMethods {
		methods = append(methods, m)
	}
	sort.Strings(methods)
	dir := s.dir()
	return InitializeResult{Name: "smith", Version: s.Version, ProtocolVersion: ProtocolVersion, EventVersion: runevents.Version, Workspace: dir, Methods: methods}
}

func (s *Server) dir() string {
	if s.Dir == "" {
		return "."
	}
	return s.Dir
}

func decodeParams(raw json.RawMessage, into any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, into)
}
