package pipeline

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Channel is installed when .flutter-version does not pin a release.
const Channel = "stable"

// flutterProject reports whether a project builds with Flutter: its tasks
// run flutter.
func flutterProject(p protocol.Project) bool {
	for _, t := range p.Tasks {
		if len(t.Command) > 0 && t.Command[0] == "flutter" {
			return true
		}
	}
	return false
}

// FlutterVersion is the release .flutter-version pins, else "".
func FlutterVersion(workspaceRoot string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, ".flutter-version"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

// flutter installs the Flutter SDK, which carries Dart.
func flutter(version string) protocol.Toolchain {
	inputs := []protocol.Input{{Name: "channel", Value: Channel}, {Name: "cache", Value: "true"}}
	ref, shown := Channel, Channel
	if version != "" {
		inputs = append(inputs, protocol.Input{Name: "flutter-version", Value: version})
		ref, shown = version, version
	}
	return protocol.Toolchain{
		Name: "flutter", Version: shown,
		GitHub: &protocol.ActionStep{Uses: "subosito/flutter-action@v2", Inputs: inputs},
		// Elsewhere (Azure, GitLab): a shallow clone of the SDK, on PATH for
		// this job and, on Azure, for the steps after it.
		Install: "git clone --depth 1 --branch " + ref + " https://github.com/flutter/flutter.git \"$HOME/flutter\"\n" +
			"export PATH=\"$HOME/flutter/bin:$PATH\"\n" +
			"echo \"##vso[task.prependpath]$HOME/flutter/bin\"",
	}
}

// dart installs the Dart SDK alone.
func dart() protocol.Toolchain {
	return protocol.Toolchain{
		Name: "dart", Version: Channel,
		GitHub: &protocol.ActionStep{Uses: "dart-lang/setup-dart@v1", Inputs: []protocol.Input{{Name: "sdk", Value: Channel}}},
		Install: "curl -fsSL https://storage.googleapis.com/dart-archive/channels/" + Channel + "/release/latest/sdk/dartsdk-linux-x64-release.zip -o /tmp/dart.zip\n" +
			"unzip -q /tmp/dart.zip -d \"$HOME\"\n" +
			"export PATH=\"$HOME/dart-sdk/bin:$PATH\"\n" +
			"echo \"##vso[task.prependpath]$HOME/dart-sdk/bin\"",
	}
}

// CI returns the Dart contribution: Flutter when a package uses it, else
// Dart; one pub get for the whole workspace; and PUB_TOKEN for a package
// that publishes to its own repository.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	usesFlutter := false
	for _, p := range params.Projects {
		usesFlutter = usesFlutter || flutterProject(p)
	}
	result := protocol.CIResult{Toolchains: []protocol.Toolchain{dart()}, Setup: []protocol.Step{{Name: "Resolve the pub workspace", Command: []string{"dart", "pub", "get"}, Blocking: true}}}
	if usesFlutter {
		result.Toolchains = []protocol.Toolchain{flutter(FlutterVersion(workspaceRoot))}
		result.Setup[0].Command[0] = "flutter"
	}
	for _, p := range params.Projects {
		if p.Publishes() {
			result.Credentials = append(result.Credentials, protocol.Credential{Env: "PUB_TOKEN", Description: "token for the repository a package's publish_to names (pub.dev publishes with GitHub's OIDC token instead)"})
			break
		}
	}
	return result
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// Doctor checks that Dart (and Flutter, when a package uses it) is installed,
// and that the SDK lives outside the workspace.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	_, dartErr := probe(workspaceRoot, []string{"dart", "--version"})
	findings := []protocol.Finding{{Check: "the Dart SDK is installed", OK: dartErr == nil, Detail: "dart is not on PATH", Remedy: "install Dart (https://dart.dev/get-dart) or Flutter, which carries it"}}
	for _, p := range params.Projects {
		if flutterProject(p) {
			_, err := probe(workspaceRoot, []string{"flutter", "--version"})
			findings = append(findings, protocol.Finding{Check: "Flutter is installed", OK: err == nil, Detail: "flutter is not on PATH", Remedy: "install Flutter (https://docs.flutter.dev/get-started/install)"})
			break
		}
	}
	for _, sdk := range []string{"flutter", "dart-sdk", ".fvm"} {
		if _, err := os.Stat(filepath.Join(workspaceRoot, sdk)); err == nil {
			findings = append(findings, protocol.Finding{Check: "the Dart and Flutter SDKs live outside the workspace", Detail: sdk + "/ is inside it", Remedy: "install the SDK outside the workspace and put it on PATH"})
			return protocol.DoctorResult{Findings: findings}
		}
	}
	findings = append(findings, protocol.Finding{Check: "the Dart and Flutter SDKs live outside the workspace", OK: true})
	return protocol.DoctorResult{Findings: findings}
}
