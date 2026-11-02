package tui

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

func TestHomeOffersAddAProjectAndComposeAStackWithKeysAndMouse(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	l.has("Where to go", "[ Add a project  a ]", "[ Compose a stack  A ]")
	l.clickText("[ Add a project")
	l.has("What are you building?")

	l = newHomeLab(t, 140, 34)
	l.keys("a")
	l.has("What are you building?")
}

func TestHomeComposeButtonStartsTheComposer(t *testing.T) {
	l := newHomeLab(t, 140, 34)
	l.clickText("[ Compose a stack")
	l.has("Reading what your plugins offer")
}

func TestProjectsListStartsWithANewProjectButtonAndTheAKey(t *testing.T) {
	l := newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.has("[ + New project  a ]")
	newLab(t, true, 200, 34).openTab(commands.TabProjects).has("a add project")
	l.clickText("[ + New project")
	l.has("What are you building?")

	l = newLab(t, true, 140, 34).openTab(commands.TabProjects)
	l.keys("a")
	l.has("What are you building?")
}
