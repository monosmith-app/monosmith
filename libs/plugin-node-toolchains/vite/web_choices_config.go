package vite

import (
	"embed"

	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
)

// choiceFiles are the choices of the plugin as data (ADR 0045): the frameworks
// and the styles of a browser app, the version catalog they name packages
// from, the base they build on and the stacks the plugin publishes. To add a
// framework or a style, add a file here.
//
//go:embed choices
var choiceFiles embed.FS

// webChoices is the checked set of the embedded files.
var webChoices = choicedata.MustLoad(choiceFiles, "choices")
