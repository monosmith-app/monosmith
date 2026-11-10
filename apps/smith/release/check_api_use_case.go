package release

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// API check modes, the values of [release] api_check.
const (
	APICheckWarn  = "warn"
	APICheckError = "error"
	APICheckOff   = "off"
)

// surfaceReader lists the public API of a project at a git ref ("" is the
// working tree). ok is false when the project's plugin cannot list one.
type surfaceReader func(ctx context.Context, item Item, ref string) (protocol.APISurfaceResult, bool)

// CheckAPI compares the public API of every package in the plan with the API
// it had at its last release, and writes a warning on each item whose commits
// ask for a smaller bump than the change needs (see CheckAPIBump). It is
// advice: a plugin that cannot list an API, or fails, gives no warning. Only
// [release] api_check = "error" turns a warning into an error, which says
// what to fix. "off" reads nothing.
func CheckAPI(ctx context.Context, s *session.Session, plan *Plan) error {
	mode := s.Config.Release.APICheck
	if mode == APICheckOff {
		return nil
	}
	read := func(ctx context.Context, item Item, ref string) (protocol.APISurfaceResult, bool) {
		p := s.Graph.Get(item.Project)
		var result protocol.APISurfaceResult
		ok, err := s.Plugins.CallContext(ctx, p.Plugin, protocol.MethodAPISurface, protocol.APISurfaceParams{Project: p.Project, Ref: ref}, &result, time.Minute)
		return result, ok && err == nil && result.Unreadable == ""
	}
	findings := checkItems(ctx, plan, func(item Item) string { return TagPrefix(s.Graph.Get(item.Project)) + item.Current }, read)
	return apiCheckError(mode, findings)
}

// apiCheckError is the error a mode makes of the findings: only "error" makes
// one.
func apiCheckError(mode string, findings []string) error {
	if len(findings) == 0 || mode != APICheckError {
		return nil
	}
	return fmt.Errorf("cannot release: the commits ask for a smaller bump than the public API change needs ([release] api_check = \"error\"):\n  - %s",
		strings.Join(findings, "\n  - "))
}

// checkItems fills Item.APIWarnings for every item with a finding and returns
// the warnings. tagOf names the tag of an item's last release.
func checkItems(ctx context.Context, plan *Plan, tagOf func(Item) string, read surfaceReader) []string {
	var all []string
	for i := range plan.Items {
		item := plan.Items[i]
		if item.Current == "" || item.Forced || ctx.Err() != nil {
			continue
		}
		before, ok := read(ctx, item, tagOf(item))
		if !ok {
			continue
		}
		after, ok := read(ctx, item, "")
		if !ok {
			continue
		}
		if f, found := CheckAPIBump(item, DiffAPI(before.Names, after.Names)); found {
			plan.Items[i].APIWarnings = []string{f.Sentence()}
			all = append(all, f.Sentence())
		}
	}
	return all
}
