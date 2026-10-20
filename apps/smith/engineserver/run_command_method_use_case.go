package engineserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/mcpserver"
	"github.com/monosmith-app/monosmith/apps/smith/runevents"
)

// argsOf turns a method's parameters into the smith arguments of the MCP tool
// the method maps to.
func argsOf(toolName string, raw json.RawMessage) ([]string, *rpcError) {
	input := map[string]any{}
	if err := decodeParams(raw, &input); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	for _, t := range mcpserver.Tools {
		if t.Name == toolName {
			args, err := t.Args(input)
			if err != nil {
				return nil, &rpcError{codeInvalidParams, err.Error()}
			}
			return args, nil
		}
	}
	return nil, &rpcError{codeFailed, "no command for " + toolName}
}

func (s *Server) execute() Execute {
	if s.Execute != nil {
		return s.Execute
	}
	return ExecuteSelf
}

// query answers with the JSON a smith command prints, changing nothing. A
// command that reports a problem through its exit code (doctor) still answers
// with its JSON.
func (s *Server) query(ctx context.Context, tool string, raw json.RawMessage) (any, *rpcError) {
	args, rerr := argsOf(tool, raw)
	if rerr != nil {
		return nil, rerr
	}
	return s.queryArgs(ctx, args)
}

// queryArgs runs a smith command and answers with the JSON it prints.
func (s *Server) queryArgs(ctx context.Context, args []string) (any, *rpcError) {
	stdout, code, err := s.execute()(ctx, s.dir(), args, nil)
	if err != nil {
		return nil, &rpcError{codeFailed, err.Error()}
	}
	body := strings.TrimSpace(string(stdout))
	if json.Valid([]byte(body)) {
		return json.RawMessage(body), nil
	}
	return nil, &rpcError{codeFailed, fmt.Sprintf("smith %s exited with code %d and printed no JSON: %s", strings.Join(args, " "), code, body)}
}

// commandArgs is the smith command a method that changes the workspace
// stands for: from the server's own table, else from the MCP tool the method
// maps to. The run param names the run; it is not an input of the command.
func commandArgs(method string, raw json.RawMessage) (args []string, id string, rerr *rpcError) {
	input := map[string]any{}
	_ = decodeParams(raw, &input)
	id, _ = input["run"].(string)
	delete(input, "run")
	cleaned, _ := json.Marshal(input)
	if build, ok := serverCommands[method]; ok {
		args, rerr = build(cleaned)
		return args, id, rerr
	}
	args, rerr = argsOf(commandMethods[method], cleaned)
	return args, id, rerr
}

// command runs a smith command that changes the workspace as a run of its own:
// run.started, the command's lines as task.output, task.finished, run.finished.
func (s *Server) command(ctx context.Context, method string, raw json.RawMessage) (any, *rpcError) {
	args, id, rerr := commandArgs(method, raw)
	if rerr != nil {
		return nil, rerr
	}
	if id == "" {
		id = runevents.NewRunID()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := s.runs.begin(id, cancel); err != nil {
		return nil, &rpcError{codeInvalidParams, err.Error()}
	}
	defer s.runs.end(id)

	stream := runevents.NewRunStream(id, s.sink())
	stream.Emit(runevents.Event{Event: runevents.RunStarted, Command: "smith " + strings.Join(args, " "), Tasks: []string{method}})
	stream.Emit(runevents.Event{Event: runevents.TaskQueued, Task: method})
	stream.Emit(runevents.Event{Event: runevents.TaskStarted, Task: method})
	began := time.Now()
	_, code, err := s.execute()(ctx, s.dir(), args, func(st runevents.Stream, line string) {
		stream.Emit(runevents.Event{Event: runevents.TaskOutput, Task: method, Stream: st, Line: line})
	})
	took := time.Since(began)
	finished := runevents.Event{Event: runevents.TaskFinished, Task: method, Status: runevents.Passed, DurationMs: took.Milliseconds()}
	summary := runevents.Summarize(1, 0, 0, 0, took)
	ok := code == 0 && err == nil
	if !ok {
		finished.Status = runevents.Failed
		finished.Reason = fmt.Sprintf("exited with code %d", code)
		if err != nil {
			finished.Reason = "cancelled"
			if ctx.Err() == nil {
				finished.Reason = err.Error()
			}
		}
		summary = runevents.Summarize(0, 1, 0, 0, took)
	}
	stream.Emit(finished)
	stream.Emit(runevents.Event{Event: runevents.RunFinished, OK: &ok, Summary: summary})
	return CommandResult{Run: id, OK: ok, ExitCode: code}, nil
}
