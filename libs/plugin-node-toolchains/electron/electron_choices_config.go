package electron

import (
	"embed"

	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
)

// choiceFiles are the choices of the plugin as data (ADR 0045): the electron
// choice, the version catalog and the shared base.
//
//go:embed choices
var choiceFiles embed.FS

// electronChoices is the checked set of the embedded files.
var electronChoices = choicedata.MustLoad(choiceFiles, "choices")
