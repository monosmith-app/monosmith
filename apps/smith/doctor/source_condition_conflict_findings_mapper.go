package doctor

import (
	"sort"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// tsNodeOptionsEnv is the variable that carries ts-node's compiler options.
const tsNodeOptionsEnv = "TS_NODE_COMPILER_OPTIONS"

// sourceConditionChecks explains a workaround it recognises in smith.toml: a
// task that sets customConditions to null in TS_NODE_COMPILER_OPTIONS. The
// finding passes. It says what the line does and why, so the next person does
// not remove it, or copy it blindly.
func sourceConditionChecks(s *session.Session) []protocol.Finding {
	return SourceConditionFindings(s.Config)
}

// SourceConditionFindings is sourceConditionChecks over the configuration.
// It returns nothing when no task group or project task resets customConditions.
func SourceConditionFindings(cfg workspace.Config) []protocol.Finding {
	var where []string
	for i, g := range cfg.TaskGroups {
		for task, t := range g.Tasks {
			if resetsCustomConditions(t.Env) {
				where = append(where, "[[task_groups]] "+strings.Join(g.Select, " and ")+" "+task+" (group "+strconv.Itoa(i+1)+")")
			}
		}
	}
	for name, p := range cfg.Projects {
		for task, t := range p.Tasks {
			if resetsCustomConditions(t.Env) {
				where = append(where, "[projects."+name+"] "+task)
			}
		}
	}
	if len(where) == 0 {
		return nil
	}
	sort.Strings(where)
	return []protocol.Finding{{
		Check:  "a tool conflict is worked around in smith.toml",
		OK:     true,
		Detail: strings.Join(where, "; ") + " set " + tsNodeOptionsEnv + " with customConditions null. tsconfig.base.json names a custom resolution condition so editors and tsc read the source of workspace packages. ts-node with module commonjs and node10 resolution cannot use it, so the task switches the condition off for ts-node only. It is one decision (resolve the source condition) written twice, in two tools.",
		Remedy: "keep the line while a task runs under ts-node; if no task does, remove it. Change the condition name in tsconfig.base.json, vite and vitest resolve.conditions and the jest resolver together",
	}}
}

// resetsCustomConditions reports whether an env map sets customConditions to
// null in the ts-node options.
func resetsCustomConditions(env map[string]string) bool {
	v, ok := env[tsNodeOptionsEnv]
	if !ok {
		return false
	}
	compact := strings.Join(strings.Fields(v), "")
	return strings.Contains(compact, `"customConditions":null`)
}
