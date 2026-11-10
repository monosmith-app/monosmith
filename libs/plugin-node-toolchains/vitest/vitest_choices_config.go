package vitest

import (
	"embed"

	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
)

// choiceFiles are the choices of the plugin as data (ADR 0045). The test axis
// has the choice vitest; the starter test it writes depends on the workspace,
// so that stays in setup.
//
//go:embed all:choices
var choiceFiles embed.FS

// choices is the checked set of the embedded files.
var choices = choicedata.MustLoad(choiceFiles, "choices")
