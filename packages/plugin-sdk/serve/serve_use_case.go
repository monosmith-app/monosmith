package serve

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

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
func Serve(p Plugin, in io.Reader, out io.Writer) error {
	stream := transport.NewStream(in, out)
	initialized := false
	for {
		body, err := stream.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		var req protocol.Request
		if err := json.Unmarshal(body, &req); err != nil {
			if err := reply(stream, nil, nil, &protocol.ResponseError{Code: protocol.CodeParseError, Message: err.Error()}); err != nil {
				return err
			}
			continue
		}
		if req.ID == nil {
			continue // notifications carry no reply; v1 defines none for plugins
		}
		if req.Method != protocol.MethodInitialize && !initialized {
			if err := reply(stream, req.ID, nil, &protocol.ResponseError{Code: protocol.CodeInvalidRequest, Message: "initialize must come first"}); err != nil {
				return err
			}
			continue
		}
		result, rpcErr := dispatch(p, req)
		if req.Method == protocol.MethodInitialize && rpcErr == nil {
			initialized = true
		}
		if err := reply(stream, req.ID, result, rpcErr); err != nil {
			return err
		}
		if req.Method == protocol.MethodShutdown {
			return nil
		}
	}
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
	_, published := p.(RegistryReader)
	_, dependencies := p.(DependencyManager)
	_, declarations := p.(DeclarationEditor)
	_, latest := p.(LatestReader)
	_, components := p.(ComponentLister)
	return protocol.Capabilities{Discover: discover, Generate: generate, Version: version, Publish: publish, CI: ci, Doctor: doctor, Contribute: contribute, Remove: remove, Published: published, Dependencies: dependencies, Declarations: declarations, Latest: latest, Components: components}
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
