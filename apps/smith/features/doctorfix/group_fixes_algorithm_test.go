package doctorfix

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func addDeps(project string) *protocol.Fix {
	return &protocol.Fix{
		Description: "declare the packages each project imports", AppliesTo: []string{project},
		Files: []protocol.FileEdit{{Path: "packages/" + project + "/package.json", Content: project}},
		Steps: []protocol.FixStep{{Description: "install", Argv: []string{"npm", "install"}}},
	}
}

func TestGroupFoldsTheSameFixOnManyProjectsIntoOne(t *testing.T) {
	var items []Item
	for _, p := range []string{"a", "b", "c"} {
		items = append(items, Item{Check: "packages/" + p + " declares the packages it imports", Fix: addDeps(p)})
	}
	plan := Group(items)
	if len(plan) != 1 {
		t.Fatalf("one fix for the three projects: %+v", plan)
	}
	p := plan[0]
	if !slices.Equal(p.AppliesTo, []string{"a", "b", "c"}) || len(p.Files) != 3 || len(p.Steps) != 1 || len(p.Checks) != 3 {
		t.Fatalf("files of all, the step once, a check each: %+v", p)
	}
	if !p.Ask || Scope(p.AppliesTo) != "applies to 3 projects: a, b, c" {
		t.Fatalf("not safe, so asked; scope %q", Scope(p.AppliesTo))
	}
}

func TestGroupKeepsFixesApartThatWriteTheSameFileOrAsk(t *testing.T) {
	one := &protocol.Fix{Description: "d", AppliesTo: []string{"x"}, Safe: true, Files: []protocol.FileEdit{{Path: "smith.toml", Content: "1"}}}
	two := &protocol.Fix{Description: "d", AppliesTo: []string{"y"}, Safe: true, Files: []protocol.FileEdit{{Path: "smith.toml", Content: "2"}}}
	if plan := Group([]Item{{Check: "x", Fix: one}, {Check: "y", Fix: two}}); len(plan) != 2 {
		t.Fatalf("two edits of one file cannot fold: %+v", plan)
	}
	q := func(subject string) *protocol.Fix {
		return &protocol.Fix{Description: "q", AppliesTo: []string{subject}, Choices: []protocol.FixChoice{{Question: "which " + subject}}}
	}
	plan := Group([]Item{{Check: "x", Fix: q("x")}, {Check: "y", Fix: q("y")}, {Check: "z"}})
	if len(plan) != 2 || !plan[0].Ask || !plan[1].Ask {
		t.Fatalf("each fix with a question stays its own, and asks: %+v", plan)
	}
}

func TestGroupMergesIdenticalFixesAndComputesAsk(t *testing.T) {
	safe := &protocol.Fix{Description: "ignore", Safe: true, Files: []protocol.FileEdit{{Path: ".gitignore", Content: "x"}}}
	risky := &protocol.Fix{Description: "rewrite", Files: []protocol.FileEdit{{Path: "ci.yml", Content: "y"}}}
	plan := Group([]Item{{Check: "a", Fix: safe}, {Check: "b", Fix: safe}, {Check: "c", Fix: risky}})
	if len(plan) != 2 || len(plan[0].Checks) != 2 || plan[0].Ask || !plan[1].Ask {
		t.Fatalf("%+v", plan)
	}
	if !Asks(protocol.Fix{Safe: true, Choices: []protocol.FixChoice{{}}}) || Asks(protocol.Fix{Safe: true}) {
		t.Fatal("a question always asks; a safe fix does not")
	}
}

func TestResolveAnswersQuestionsInOrderAndReturnsTheNextOne(t *testing.T) {
	install := protocol.Fix{Steps: []protocol.FixStep{{Argv: []string{"npm", "i", "-D", "oxfmt"}}}}
	remove := protocol.Fix{Files: []protocol.FileEdit{{Path: "package.json", Content: "without"}}}
	again := protocol.Fix{Choices: []protocol.FixChoice{{Question: "also?", Options: []protocol.FixChoiceOption{{Label: "yes", Fix: install}}}}}
	fix := protocol.Fix{Description: "oxfmt", Choices: []protocol.FixChoice{
		{Question: "install or remove?", Options: []protocol.FixChoiceOption{{Label: "install", Fix: install}, {Label: "remove", Fix: again}}},
		{Question: "second", Options: []protocol.FixChoiceOption{{Label: "remove", Fix: remove}}},
	}}
	got, next, err := Resolve(fix, nil)
	if err != nil || next == nil || next.Question != "install or remove?" || len(got.Choices) != 2 {
		t.Fatalf("with no answers the first question: %+v %v", next, err)
	}
	got, next, err = Resolve(fix, []int{1})
	if err != nil || next == nil || next.Question != "second" || len(got.Steps) != 1 {
		t.Fatalf("the first answer adds its step: %+v %+v", got, next)
	}
	got, next, err = Resolve(fix, []int{2, 1})
	if err != nil || next == nil || next.Question != "also?" || len(got.Files) != 1 {
		t.Fatalf("an option's own question comes after the pending ones: %+v %+v", got, next)
	}
	got, next, err = Resolve(fix, []int{2, 1, 1})
	if err != nil || next != nil || len(got.Choices) != 0 || len(got.Steps) != 1 || len(got.Files) != 1 {
		t.Fatalf("all answered: %+v %+v %v", got, next, err)
	}
	if _, _, err := Resolve(fix, []int{3}); err == nil {
		t.Fatal("an answer outside the options is refused")
	}
	if len(fix.Choices) != 2 {
		t.Fatal("resolving must not change the fix it was given")
	}
}
