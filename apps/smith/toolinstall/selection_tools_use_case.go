package toolinstall

import (
	"context"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// SessionSources reads what the loaded plugins declare for this system.
//
// Parameters: s is the open session.
// Returns: one source per plugin that declares a tool.
func SessionSources(s *session.Session) []Source { return sourcesOf(s) }

// SelectionStatuses detects the tools a chosen stack needs: the tools of its
// language, of its function host, and of its desktop shell (Tauri). It reads
// the tool data the plugins publish (ADR 0077) and runs no install.
//
// Parameters: ctx bounds the detection; d is the detector; workspace is the
// absolute workspace root, where a version pin is looked up; sources are the
// plugins' declarations; language is the language plugin's name; selection are
// the chosen choices and toolchain plugins.
// Returns: one status per tool, in the order the plugins declare them.
func SelectionStatuses(ctx context.Context, d Detector, workspace string, sources []Source, language string, selection []string) []Status {
	uses := []Use{{Language: language, Choices: selection}}
	return DetectAll(ctx, d, Needed(workspace, nil, sources, uses, nil, nil))
}
