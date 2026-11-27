package tui

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/sources"
)

func TestHomeOffersAddAProjectAndComposeAStackWithKeysAndMouse(t *testing.T) {
	l := showOnHome(newHomeLab(t, 140, 34), "[Add a project")
	l.has("[Add a project (a)]", "[Compose a stack (A)]")
	l.clickText("[Add a project")
	l.has("What are you building?")

	l = newHomeLab(t, 140, 34)
	l.keys("a")
	l.has("What are you building?")
}

func TestHomeComposeButtonStartsTheComposer(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	// The catalogue takes a moment: the dialog that says so is on screen until it arrives.
	l.slowRead(sources.Captured{OK: true, Text: "{}"})
	showOnHome(l, "[Compose a stack")
	l.clickText("[Compose a stack")
	l.has("Compose a stack", "Reading what your plugins offer")
}

func TestProjectsKeepsTheAKeyOfNewProjectWithoutItsButton(t *testing.T) {
	// The buttons of the workspace left the list: the key, the footer and the palette keep them.
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.hasNot("[+ New project (a)]")
	newLab(t, true, 260, 34).openTab(commands.TabProjects).has("a add project")
	l.keys("a")
	l.has("What are you building?")
}
