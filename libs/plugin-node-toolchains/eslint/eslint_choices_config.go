package eslint

import (
	"embed"

	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
)

// choiceFiles are the choices of the plugin as data (ADR 0045). Lint is an
// axis; eslint is its default. The files it writes depend on the workspace, so
// they stay in setup.
//
//go:embed all:choices
var choiceFiles embed.FS

// choices is the checked set of the embedded files.
var choices = choicedata.MustLoad(choiceFiles, "choices")
