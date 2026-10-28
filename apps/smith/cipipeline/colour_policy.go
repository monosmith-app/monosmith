package cipipeline

import (
	"os"
	"strings"

	"golang.org/x/term"
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

// colourWanted decides, in this one place, whether the log carries ANSI
// colour. GitHub Actions, Azure Pipelines and GitLab render it; elsewhere it
// is for a terminal only. NO_COLOR and TERM=dumb always win; FORCE_COLOR turns
// it on for a captured sample.
func colourWanted(getenv func(string) string, terminal bool) bool {
	switch {
	case getenv("NO_COLOR") != "" || getenv("TERM") == "dumb":
		return false
	case getenv("FORCE_COLOR") != "" && getenv("FORCE_COLOR") != "0":
		return true
	case getenv("GITHUB_ACTIONS") == "true", strings.EqualFold(getenv("TF_BUILD"), "true"), getenv("GITLAB_CI") != "":
		return true
	}
	return terminal
}

// colourOn is colourWanted for this process.
func colourOn() bool {
	return colourWanted(os.Getenv, term.IsTerminal(int(os.Stdout.Fd())))
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
	{"✓", tintGreen}, {"[v]", tintGreen}, {"✗", tintRed}, {"[x]", tintRed},
	{"▲", tintYellow}, {"[!]", tintYellow}, {"↷", tintDim}, {"[-]", tintDim},
	{"○", tintDim}, {"[ ]", tintDim}, {"⚡", tintDim}, {"[c]", tintDim},
}

func glyphTint(cell string) tint {
	for _, g := range glyphTints {
		if strings.HasPrefix(cell, g.prefix+" ") || cell == g.prefix {
			return g.tint
		}
	}
	return tintNone
}
