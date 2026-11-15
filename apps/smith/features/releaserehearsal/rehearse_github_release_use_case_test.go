package releaserehearsal

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/githubrelease"
)

type fakePush struct {
	ok    bool
	err   error
	cache int64
}

func (f fakePush) CanPush(githubrelease.Repository) (bool, error) { return f.ok, f.err }

func (f fakePush) CacheUsage(githubrelease.Repository) (int64, error) { return f.cache, nil }

var repo = githubrelease.Repository{Owner: "acme", Name: "app"}

func TestTheGitHubRehearsalPassesWhenTheTokenMayWrite(t *testing.T) {
	var r Report
	CheckGitHubRelease(&r, []string{"core", "cli"}, repo, fakePush{ok: true}, nil)
	if r.Passed != 2 || r.Failed != 0 || r.Results[0].Project != "core, cli" || r.Results[1].Registry != "github.com/acme/app" {
		t.Fatalf("%+v", r)
	}
}

func TestTheGitHubRehearsalFailsWithTheFixWhenTheTokenCannotWrite(t *testing.T) {
	var r Report
	CheckGitHubRelease(&r, []string{"core"}, repo, fakePush{}, nil)
	if r.Failed != 1 || !strings.Contains(r.Results[1].Fix, "contents: write") {
		t.Fatalf("%+v", r)
	}
}

func TestTheGitHubRehearsalFailsWithoutAToken(t *testing.T) {
	var r Report
	CheckGitHubRelease(&r, []string{"core"}, repo, fakePush{ok: true}, githubrelease.ErrNoToken)
	if r.Failed != 1 || len(r.Results) != 1 || !strings.Contains(r.Results[0].Fix, "gh auth login") {
		t.Fatalf("%+v", r)
	}
}

func TestTheGitHubRehearsalNamesAnInvisibleRepository(t *testing.T) {
	var r Report
	CheckGitHubRelease(&r, []string{"core"}, repo, fakePush{err: errors.New("404 Not Found")}, nil)
	if r.Failed != 1 || !strings.Contains(r.Results[1].Message, "404") {
		t.Fatalf("%+v", r)
	}
}

func TestNoProjectNoCheck(t *testing.T) {
	var r Report
	CheckGitHubRelease(&r, nil, repo, fakePush{}, nil)
	if len(r.Results) != 0 {
		t.Fatal("nothing attaches to GitHub, so nothing is asked")
	}
}

func TestTheGitHubRehearsalWarnsWhenTheActionsCacheIsNearlyFull(t *testing.T) {
	var r Report
	CheckGitHubRelease(&r, []string{"core"}, repo, fakePush{ok: true, cache: 9_500_000_000}, nil)
	last := r.Results[len(r.Results)-1]
	if last.Check != CheckStorage || last.Status != Warn || !strings.Contains(last.Message, "9.5 GB of the 10 GB") {
		t.Fatalf("want a storage warning, got %+v", last)
	}
	r = Report{}
	CheckGitHubRelease(&r, []string{"core"}, repo, fakePush{ok: true, cache: 4_000_000_000}, nil)
	for _, res := range r.Results {
		if res.Check == CheckStorage {
			t.Fatalf("a cache with room left needs no word: %+v", res)
		}
	}
}
