package engineserver

import (
	"context"
	"encoding/json"
)

// serverCommands are the methods that change the workspace and have no MCP
// tool: an assistant is not handed a release or a settings edit, but an editor
// that shows the plan and asks the person may call them. Each turns the
// method's params into the smith arguments.
var serverCommands = map[string]func(raw json.RawMessage) ([]string, *rpcError){
	"release":      releaseArgs,
	"config.set":   configSetArgs,
	"config.unset": configUnsetArgs,
}

// serverMethods are answered by the server itself, in process, without
// running smith: settings and recipes are read where the workspace is.
var serverMethods = map[string]func(s *Server, ctx context.Context, raw json.RawMessage) (any, *rpcError){
	"config.list":     (*Server).configList,
	"config.get":      (*Server).configGet,
	"recipe.list":     (*Server).recipeList,
	"recipe.describe": (*Server).recipeDescribe,
	"recipe.roles":    (*Server).recipeRoles,
}

// queryMethods answer with the JSON of the smith command they stand for, and
// change nothing. Each takes the input of the MCP tool it maps to (see
// mcpserver), so what a client sends and what the command line is are decided
// in one place.
var queryMethods = map[string]string{
	"info":            "smith_info",
	"projects":        "smith_projects",
	"graph":           "smith_graph",
	"affected":        "smith_affected",
	"kinds":           "smith_kinds",
	"doctor":          "smith_doctor",
	"doctor.fixes":    "smith_doctor_fixes",
	"release.preview": "smith_release_preview",
	"updates":         "smith_updates",
	"upgrade.preview": "smith_upgrade_preview",
	"explain":         "smith_explain",
	"recipe.check":    "smith_recipe_check",
	"plugins":         "smith_plugin_list",
}

// commandMethods change the workspace. Each is a run of its own: it emits
// the event stream (its output as task.output) and answers when it ends.
var commandMethods = map[string]string{
	"install":    "smith_install",
	"uninstall":  "smith_uninstall",
	"add":        "smith_add_project",
	"remove":     "smith_remove_project",
	"upgrade":    "smith_upgrade",
	"doctor.fix": "smith_doctor_fix",
	"sync":       "smith_sync",
}
