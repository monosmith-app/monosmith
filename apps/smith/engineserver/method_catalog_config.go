package engineserver

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
