package pipeline

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// RubyVersion is installed when .ruby-version does not pin one; bump it
// deliberately.
const RubyVersion = "3.4"

// Version is the Ruby .ruby-version pins, else RubyVersion.
func Version(workspaceRoot string) string {
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, ".ruby-version")); err == nil {
		if v := strings.TrimPrefix(strings.TrimSpace(string(raw)), "ruby-"); v != "" {
			return v
		}
	}
	return RubyVersion
}

func runtime(version string) protocol.Toolchain {
	return protocol.Toolchain{
		Name: "ruby", Version: version,
		GitHub: &protocol.ActionStep{Uses: "ruby/setup-ruby@v1", Inputs: []protocol.Input{{Name: "ruby-version", Value: version}}},
		Azure:  &protocol.ActionStep{Uses: "UseRubyVersion@0", Inputs: []protocol.Input{{Name: "versionSpec", Value: version}}},
		// Elsewhere (GitLab, on a Debian image): the distribution's Ruby.
		Install: "apt-get update -qq && apt-get install -y -qq ruby-full build-essential\n" +
			"gem install bundler --no-document",
		Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers/features/ruby:1", Options: map[string]string{"version": version}},
	}
}

// CI returns the Ruby contribution: Ruby, one bundle install for the
// workspace, and the gem server's key when a gem is published.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	result := protocol.CIResult{
		Toolchains: []protocol.Toolchain{runtime(Version(workspaceRoot))},
		Setup:      []protocol.Step{{Name: "Install the bundle", Command: []string{"bundle", "install"}, Blocking: true}},
	}
	for _, p := range params.Projects {
		if p.Publishes() {
			result.Credentials = append(result.Credentials, protocol.Credential{Env: "GEM_HOST_API_KEY", Description: "RubyGems API key that can push the workspace's gems", Required: true})
			break
		}
	}
	return result
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that Ruby and Bundler are installed.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, rubyErr := probe(workspaceRoot, []string{"ruby", "--version"})
	_, bundleErr := probe(workspaceRoot, []string{"bundle", "--version"})
	return protocol.DoctorResult{Findings: []protocol.Finding{
		{Check: "Ruby is installed", OK: rubyErr == nil, Detail: "ruby is not on PATH", Remedy: "install Ruby " + Version(workspaceRoot) + " (https://www.ruby-lang.org)"},
		{Check: "Bundler is installed", OK: bundleErr == nil, Detail: "bundle is not on PATH", Remedy: "gem install bundler"},
	}}
}
