package cipipeline

import (
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/termcolour"
)

// tint is a colour a log word or glyph may carry; tintNone leaves it plain.
type tint int

const (
	tintNone tint = iota
	tintGreen
	tintYellow
	tintRed
	tintDim
)

var tintCodes = map[tint]string{tintGreen: "32", tintYellow: "33", tintRed: "31", tintDim: "2"}

// colourOn is the single colour decision (termcolour.Wanted) for this process.
func colourOn() bool {
	return termcolour.Wanted(os.Getenv, term.IsTerminal(int(os.Stdout.Fd())))
}

// paint wraps text in the tint's ANSI code when colour is on.
func paint(text string, t tint, on bool) string {
	code, ok := tintCodes[t]
	if !on || !ok || text == "" {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

// glyphTints colours a status by its leading symbol (Unicode or ASCII set), so
// every table reads the same: the word and the glyph carry the meaning, the
// colour only backs it.
var glyphTints = []struct {
	prefix string
	tint   tint
}{
	{marks.Passed, tintGreen}, {"[v]", tintGreen}, {marks.Failed, tintRed}, {"[x]", tintRed},
	{marks.Warning, tintYellow}, {"[!]", tintYellow}, {marks.Skipped, tintDim}, {"[-]", tintDim},
	{"○", tintDim}, {"[ ]", tintDim}, {marks.Cached, tintDim}, {"[c]", tintDim},
}

func glyphTint(cell string) tint {
	for _, g := range glyphTints {
		if strings.HasPrefix(cell, g.prefix+" ") || cell == g.prefix {
			return g.tint
		}
	}
	return tintNone
}
