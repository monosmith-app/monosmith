package serve

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/transport"
)

// Main serves the plugin on the process's stdin and stdout and exits. Use it
// as the whole body of a plugin's main function.
func Main(p Plugin) {
	if err := Serve(p, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Serve answers requests until the host sends shutdown or closes the stream.
//
// It reads messages on the calling goroutine and runs each request on a
// bounded pool of workers, as many as the plugin's Describe declared in
// Concurrency (one when it said nothing, which is a plugin that handles one
// request at a time, in the order they arrive). Responses may be written in
// any order; the host matches them by id. A plugin that is Cancellable has the
// context of a request cancelled when the host sends a cancel notification for
// it.
func Serve(p Plugin, in io.Reader, out io.Writer) error {
	s := &server{plugin: p, stream: transport.NewStream(in, out), pool: newRequestPool()}
	_, s.cancellable = p.(Cancellable)
	return s.loop()
}

// server is one Serve call.
type server struct {
	plugin      Plugin
	stream      *transport.Stream
	pool        *requestPool
	cancellable bool
	initialized bool
	// writeErr is the first error writing a response, set by a worker.
	writeErr atomic.Pointer[error]
}

func (s *server) loop() error {
	defer s.stop()
	for {
		if err := s.writeFailure(); err != nil {
			return err
		}
		body, err := s.stream.Read()
		if errors.Is(err, io.EOF) {
			return s.writeFailure()
		}
		if err != nil {
			return err
		}
		var req protocol.Request
		if err := json.Unmarshal(body, &req); err != nil {
			s.reply(nil, nil, &protocol.ResponseError{Code: protocol.CodeParseError, Message: err.Error()})
			continue
		}
		if req.ID == nil {
			s.notified(req)
			continue
		}
		if req.Method != protocol.MethodInitialize && !s.initialized {
			s.reply(req.ID, nil, &protocol.ResponseError{Code: protocol.CodeInvalidRequest, Message: "initialize must come first"})
			continue
		}
		switch req.Method {
		case protocol.MethodInitialize:
			result, rpcErr := dispatch(s.plugin, req)
			if rpcErr == nil {
				s.initialized = true
				s.pool.setLimit(workersFor(concurrencyOf(result)), s.run)
			}
			s.reply(req.ID, result, rpcErr)
		case protocol.MethodShutdown:
			// Everything read before shutdown, started or still waiting, is
			// answered with its result first, in both SDKs; only a cancel
			// message stops a request. The host sends shutdown with nothing in
			// flight, and kills a plugin that takes longer than its timeout.
			s.pool.wait()
			s.reply(req.ID, struct{}{}, nil)
			return s.writeFailure()
		default:
			s.submit(req)
		}
	}
}

// notified handles a message without an id. cancel is the only one the
// protocol defines for a plugin; every other is ignored, and so is a cancel
// to a plugin that did not advertise it.
func (s *server) notified(req protocol.Request) {
	if req.Method != protocol.MethodCancel || !s.cancellable {
		return
	}
	var params protocol.CancelParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return
	}
	if waiting := s.pool.cancel(params.ID); waiting != nil {
		s.answerCancelled(waiting)
	}
}

func (s *server) submit(req protocol.Request) {
	switch s.pool.submit(req, s.run) {
	case refusedDuplicate:
		s.reply(req.ID, nil, &protocol.ResponseError{Code: protocol.CodeInvalidRequest, Message: fmt.Sprintf("request id %d is already in flight", *req.ID)})
	case refusedFull:
		s.reply(req.ID, nil, &protocol.ResponseError{Code: protocol.CodeInvalidRequest, Message: "too many requests in flight: the host must not send more than the plugin's declared concurrency"})
	}
}

// run is a worker's job: handle one request and answer it.
func (s *server) run(pr *pooledRequest) {
	handler := s.plugin
	if c, ok := s.plugin.(Cancellable); ok {
		handler = c.WithContext(pr.ctx)
	}
	result, rpcErr := dispatch(handler, pr.req)
	if rpcErr != nil && rpcErr.Code == protocol.CodePluginFailure && pr.cancelled.Load() {
		rpcErr = &protocol.ResponseError{Code: protocol.CodeRequestCancelled, Message: "request cancelled: " + rpcErr.Message}
	}
	s.pool.release(pr)
	s.reply(pr.req.ID, result, rpcErr)
}

func (s *server) answerCancelled(pr *pooledRequest) {
	s.reply(pr.req.ID, nil, &protocol.ResponseError{Code: protocol.CodeRequestCancelled, Message: "request cancelled before it started"})
}

// drain ends the work in flight when the input ends (the host is gone): a
// cancellable plugin is asked to stop and what had not started is refused;
// any other finishes what it is doing.
func (s *server) drain() {
	if s.cancellable {
		for _, pr := range s.pool.cancelAll() {
			s.answerCancelled(pr)
		}
	}
	s.pool.wait()
}

// stop runs when the loop ends for any reason, so no worker outlives Serve.
func (s *server) stop() { s.drain() }

func (s *server) writeFailure() error {
	if err := s.writeErr.Load(); err != nil {
		return *err
	}
	return nil
}

func (s *server) reply(id *int64, result any, rpcErr *protocol.ResponseError) {
	if err := reply(s.stream, id, result, rpcErr); err != nil {
		s.writeErr.CompareAndSwap(nil, &err)
	}
}

// concurrencyOf reads the declared concurrency out of an initialize result.
func concurrencyOf(result any) int {
	if r, ok := result.(protocol.InitializeResult); ok {
		return r.Concurrency
	}
	return 0
}

func dispatch(p Plugin, req protocol.Request) (any, *protocol.ResponseError) {
	switch req.Method {
	case protocol.MethodInitialize:
		return call(req.Params, func(params protocol.InitializeParams) (protocol.InitializeResult, error) {
			return initialize(p, params)
		})
	case protocol.MethodShutdown:
		return struct{}{}, nil
	case protocol.MethodDiscover:
		if d, ok := p.(Discoverer); ok {
			return call(req.Params, d.Discover)
		}
	case protocol.MethodGenerate:
		if g, ok := p.(Generator); ok {
			return call(req.Params, g.Generate)
		}
	case protocol.MethodVersion:
		if v, ok := p.(Versioner); ok {
			return call(req.Params, v.WriteVersion)
		}
	case protocol.MethodPublish:
		if pub, ok := p.(Publisher); ok {
			return call(req.Params, pub.Publish)
		}
	case protocol.MethodCI:
		if c, ok := p.(CIContributor); ok {
			return call(req.Params, c.CI)
		}
	case protocol.MethodDoctor:
		if d, ok := p.(DoctorContributor); ok {
			return call(req.Params, d.Doctor)
		}
	case protocol.MethodContribute:
		if c, ok := p.(Contributor); ok {
			return call(req.Params, c.Contribute)
		}
	case protocol.MethodRemove:
		if r, ok := p.(Remover); ok {
			return call(req.Params, r.Remove)
		}
	case protocol.MethodChangeKind:
		if k, ok := p.(KindChanger); ok {
			return call(req.Params, k.ChangeKind)
		}
	case protocol.MethodPublished:
		if r, ok := p.(RegistryReader); ok {
			return call(req.Params, r.Published)
		}
	case protocol.MethodInstall:
		if d, ok := p.(DependencyManager); ok {
			return call(req.Params, d.Install)
		}
	case protocol.MethodUninstall:
		if d, ok := p.(DependencyManager); ok {
			return call(req.Params, d.Uninstall)
		}
	case protocol.MethodDeclared:
		if d, ok := p.(DeclarationEditor); ok {
			return call(req.Params, d.Declared)
		}
	case protocol.MethodRewrite:
		if d, ok := p.(DeclarationEditor); ok {
			return call(req.Params, d.Rewrite)
		}
	case protocol.MethodLatest:
		if l, ok := p.(LatestReader); ok {
			return call(req.Params, l.Latest)
		}
	case protocol.MethodComponents:
		if l, ok := p.(ComponentLister); ok {
			return call(req.Params, l.Components)
		}
	case protocol.MethodDerive:
		if d, ok := p.(Deriver); ok {
			return call(req.Params, d.Derive)
		}
	}
	return nil, &protocol.ResponseError{Code: protocol.CodeMethodNotFound, Message: "method not supported: " + req.Method}
}

func initialize(p Plugin, params protocol.InitializeParams) (protocol.InitializeResult, error) {
	supported := false
	for _, v := range params.ProtocolVersions {
		if v == protocol.Version {
			supported = true
		}
	}
	if !supported {
		return protocol.InitializeResult{}, fmt.Errorf("no common protocol version: host speaks %v, plugin speaks %d", params.ProtocolVersions, protocol.Version)
	}
	result, err := p.Describe(params)
	if err != nil {
		return result, err
	}
	result.ProtocolVersion = protocol.Version
	result.Capabilities = CapabilitiesOf(p)
	return result, nil
}

// CapabilitiesOf derives the advertised capabilities from the interfaces p
// implements.
func CapabilitiesOf(p Plugin) protocol.Capabilities {
	_, discover := p.(Discoverer)
	_, generate := p.(Generator)
	_, version := p.(Versioner)
	_, publish := p.(Publisher)
	_, ci := p.(CIContributor)
	_, doctor := p.(DoctorContributor)
	_, contribute := p.(Contributor)
	_, remove := p.(Remover)
	_, changeKind := p.(KindChanger)
	_, published := p.(RegistryReader)
	_, dependencies := p.(DependencyManager)
	_, declarations := p.(DeclarationEditor)
	_, latest := p.(LatestReader)
	_, cancel := p.(Cancellable)
	_, components := p.(ComponentLister)
	_, derive := p.(Deriver)
	return protocol.Capabilities{Derive: derive, Discover: discover, Generate: generate, Version: version, Publish: publish, CI: ci, Doctor: doctor, Contribute: contribute, Remove: remove, ChangeKind: changeKind, Published: published, Dependencies: dependencies, Declarations: declarations, Latest: latest, Cancel: cancel, Components: components}
}

func call[P, R any](raw json.RawMessage, fn func(P) (R, error)) (any, *protocol.ResponseError) {
	var params P
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			return nil, &protocol.ResponseError{Code: protocol.CodeInvalidParams, Message: err.Error()}
		}
	}
	result, err := fn(params)
	if err != nil {
		return nil, &protocol.ResponseError{Code: protocol.CodePluginFailure, Message: err.Error()}
	}
	return result, nil
}

func reply(stream *transport.Stream, id *int64, result any, rpcErr *protocol.ResponseError) error {
	resp := protocol.Response{JSONRPC: "2.0", ID: id, Error: rpcErr}
	if rpcErr == nil {
		raw, err := json.Marshal(result)
		if err != nil {
			resp.Error = &protocol.ResponseError{Code: protocol.CodeInternalError, Message: err.Error()}
		} else {
			resp.Result = raw
		}
	}
	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return stream.Write(body)
}
