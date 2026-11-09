package adoptwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/adoptflow"
	"github.com/monosmith-app/monosmith/apps/smith/convert"
)

func TestTheResultDoesNotSayWorkspaceWhenNoProjectWasFound(t *testing.T) {
	text := doneText(adoptflow.Found{Name: "empty"}, convert.Highlights{}, "", nil, nil, nil, nil, false)
	if strings.Contains(text, "is a smith workspace now") || !strings.Contains(text, "found no project") || !strings.Contains(text, "smith add") {
		t.Fatalf("result:\n%s", text)
	}
}

func TestTheResultSaysWhetherTheChangeIsACommit(t *testing.T) {
	f := adoptflow.Found{Name: "app", Projects: []adoptflow.Project{{Name: "a"}}}
	if text := doneText(f, convert.Highlights{}, "", nil, nil, nil, nil, true); !strings.Contains(text, "smith undo reverts it") {
		t.Fatalf("committed:\n%s", text)
	}
	if text := doneText(f, convert.Highlights{}, "", nil, nil, nil, nil, false); !strings.Contains(text, "Not committed: review with git status") {
		t.Fatalf("not committed:\n%s", text)
	}
}

func TestTheConservativeReviewOffersTheCommitBox(t *testing.T) {
	f := adoptflow.Found{Name: "app"}
	if text := conservativeText(f, "cmd", adoptflow.Changes{}, "", convert.Highlights{}, false); !strings.Contains(text, "[ ] Commit this change so it can be undone") {
		t.Fatalf("unticked:\n%s", text)
	}
	if text := conservativeText(f, "cmd", adoptflow.Changes{}, "", convert.Highlights{}, true); !strings.Contains(text, "[x] Commit this change so it can be undone") {
		t.Fatalf("ticked:\n%s", text)
	}
}

func TestTheConversionReviewOffersTheCommitBoxToo(t *testing.T) {
	f := adoptflow.Found{Name: "app", Kind: adoptflow.KindNx, Source: "Nx workspace"}
	if text := conservativeText(f, "", adoptflow.Changes{}, "", convert.Highlights{}, true); !strings.Contains(text, "[x] Commit this change so it can be undone") {
		t.Fatalf("ticked:\n%s", text)
	}
}
