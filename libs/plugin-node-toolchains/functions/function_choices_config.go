package functions

import (
	"embed"

	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
)

// choiceFiles are the choices of the plugin as data (ADR 0045): the three hosts,
// the triggers, the Node runtimes, the version catalog and the shared base. To
// add a runtime or a trigger, add a file here.
//
//go:embed choices
var choiceFiles embed.FS

// functionChoices is the checked set of the embedded files.
var functionChoices = choicedata.MustLoad(choiceFiles, "choices")
