package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// bigFile is the size from which a file is worth a warning.
const bigFile = 1 << 20

// errStopped is what a scan returns when it is cancelled. The SDK answers a
// cancelled request that fails with the request-cancelled code, whatever the
// message; a fixed one keeps this plugin and its TypeScript twin byte-equal.
var errStopped = errors.New("the scan was stopped")

// scan is a structure plugin with a doctor that walks every file of a project.
// A walk can take a while, so the plugin shows the two things a protocol
// client expects of such a method:
//
//   - it is Cancellable: Serve gives each request a copy of the plugin whose
//     context ends when smith cancels the request (WithContext), the walk
//     watches it, and the plugin advertises the cancel capability;
//   - it declares a Concurrency of 4: smith may send up to four requests at
//     once, and Serve runs them on four workers. Scanning shares no state, so
//     this costs the plugin nothing.
type scan struct {
	root  string
	delay time.Duration
	ctx   context.Context
}

// Describe answers "initialize".
func (s *scan) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	s.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:      protocol.Peer{Name: "widget-scan", Version: "0.1.0"},
		Kind:        protocol.KindConvention,
		Roles:       []protocol.Role{protocol.RoleStructure},
		Languages:   []string{"widget"},
		Concurrency: 4,
	}, nil
}

// WithContext makes scan a serve.Cancellable: it returns the plugin as it
// serves one request. The copy shares nothing mutable with the original.
func (s *scan) WithContext(ctx context.Context) serve.Plugin {
	c := *s
	c.ctx = ctx
	return &c
}

// Doctor answers "doctor": no file of a project is bigger than a mebibyte.
func (s *scan) Doctor(p protocol.DoctorParams) (protocol.DoctorResult, error) {
	result := protocol.DoctorResult{Findings: []protocol.Finding{}}
	for _, project := range p.Projects {
		big, err := s.bigFiles(filepath.Join(s.root, filepath.FromSlash(project.Root)))
		if err != nil {
			return protocol.DoctorResult{}, err
		}
		f := protocol.Finding{Check: project.Name + " has no file over 1 MiB", OK: len(big) == 0}
		if len(big) > 0 {
			f.Detail = strings.Join(big, ", ")
			f.Remedy = "keep large files out of the repository: " + big[0]
			f.Warning = true // reported, but does not fail smith doctor
		}
		result.Findings = append(result.Findings, f)
	}
	return result, nil
}

// bigFiles lists the files below dir over bigFile, relative to dir. It checks
// the context before every file, so a cancel ends it within one file's time.
func (s *scan) bigFiles(dir string) ([]string, error) {
	var found []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if s.stopped(s.delay) {
			return errStopped
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() > bigFile {
			rel, _ := filepath.Rel(dir, path)
			found = append(found, filepath.ToSlash(rel))
		}
		return nil
	})
	if errors.Is(err, errStopped) {
		return nil, errStopped
	}
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", filepath.Base(dir), err)
	}
	return found, nil
}

// stopped waits for the delay, or for the request to be cancelled, and reports
// whether it was cancelled. A plugin that was not given a context (served
// without Cancellable, as in a test of the walk alone) is never cancelled.
func (s *scan) stopped(delay time.Duration) bool {
	if s.ctx == nil {
		time.Sleep(delay)
		return false
	}
	select {
	case <-s.ctx.Done():
		return true
	case <-time.After(delay):
		return s.ctx.Err() != nil
	}
}
