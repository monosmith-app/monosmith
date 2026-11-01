package overridelearn

import (
	"fmt"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func off() *bool { b := false; return &b }

func cfgWith(n int, override func(i int) workspace.TaskOverride, defaults map[string]workspace.TaskOverride) workspace.Config {
	cfg := workspace.Config{Tasks: defaults, Projects: map[string]workspace.ProjectOverride{}}
	for i := 0; i < n; i++ {
		cfg.Projects[fmt.Sprintf("p%d", i)] = workspace.ProjectOverride{Tasks: map[string]workspace.TaskOverride{"test": override(i)}}
	}
	return cfg
}

func TestTheSameOverrideInEnoughProjectsIsLifted(t *testing.T) {
	cfg := cfgWith(3, func(int) workspace.TaskOverride { return workspace.TaskOverride{Cache: off()} }, nil)
	got := Analyse(cfg, 3, DefaultThreshold)
	if len(got) != 1 || got[0].Kind != Lift || got[0].Task != "test" || got[0].Field != "cache" || got[0].Value != "false" || len(got[0].Projects) != 3 {
		t.Fatalf("got %+v", got)
	}
	if got := Analyse(cfg, 3, 4); len(got) != 0 {
		t.Errorf("below the threshold: %+v", got)
	}
}

func TestAMinorityIsLeftToTaskGroups(t *testing.T) {
	cfg := cfgWith(3, func(int) workspace.TaskOverride { return workspace.TaskOverride{Cache: off()} }, nil)
	if got := Analyse(cfg, 10, DefaultThreshold); len(got) != 0 {
		t.Errorf("3 of 10 projects must not become a workspace default: %+v", got)
	}
}

func TestDifferentValuesAreNotAPattern(t *testing.T) {
	cfg := cfgWith(4, func(i int) workspace.TaskOverride {
		return workspace.TaskOverride{Env: map[string]string{"N": fmt.Sprint(i % 2)}}
	}, nil)
	if got := Analyse(cfg, 4, DefaultThreshold); len(got) != 0 {
		t.Errorf("got %+v", got)
	}
}

func TestAValueTheDefaultAlreadyHasIsNotSuggested(t *testing.T) {
	cfg := cfgWith(3, func(int) workspace.TaskOverride { return workspace.TaskOverride{Cache: off()} },
		map[string]workspace.TaskOverride{"test": {Cache: off()}})
	if got := Analyse(cfg, 3, DefaultThreshold); len(got) != 0 {
		t.Errorf("got %+v", got)
	}
}

func TestADefaultAlmostEveryProjectOverridesIsReported(t *testing.T) {
	yes := true
	defaults := map[string]workspace.TaskOverride{"test": {Cache: &yes}}
	cfg := cfgWith(4, func(int) workspace.TaskOverride { return workspace.TaskOverride{Cache: off()} }, defaults)
	var replace []Suggestion
	for _, s := range Analyse(cfg, 5, DefaultThreshold) {
		if s.Kind == Replace {
			replace = append(replace, s)
		}
	}
	if len(replace) != 1 || replace[0].Value != "true" || len(replace[0].Projects) != 4 {
		t.Fatalf("got %+v", replace)
	}
	for _, s := range Analyse(cfg, 10, DefaultThreshold) {
		if s.Kind == Replace {
			t.Errorf("4 of 10 is not nearly every project: %+v", s)
		}
	}
}
