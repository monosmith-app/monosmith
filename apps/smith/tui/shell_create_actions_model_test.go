package tui

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

func TestHomeOffersAddAProjectAndComposeAStackWithKeysAndMouse(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	l.has("Where to go", "[Add a project (a)]", "[Compose a stack (A)]")
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
	l.clickText("[Compose a stack")
	l.has("Compose a stack", "Reading what your plugins offer")
}

func TestProjectsListStartsWithANewProjectButtonAndTheAKey(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.has("[+ New project (a)]")
	newLab(t, true, 260, 34).openTab(commands.TabProjects).has("add project (a)")
	l.clickText("[+ New project")
	l.has("What are you building?")

	l = newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.keys("a")
	l.has("What are you building?")
}
