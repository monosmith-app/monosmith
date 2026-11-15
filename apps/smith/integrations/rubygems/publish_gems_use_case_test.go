package rubygems

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPublishSkipsWhatTheHostHas(t *testing.T) {
	root := t.TempDir()
	old := push
	defer func() { push = old }()
	push = func(g Gem) error {
		if strings.Contains(filepath.Base(g.File), "musl") {
			return ErrExists
		}
		return nil
	}
	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Published) != 2 || len(res.Existing) != 1 || res.Existing[0] != "acme-5.0.0.pre.rc.1-x86_64-linux-musl.gem" {
		t.Errorf("result = %+v", res)
	}
}

func TestPublishDryRunOnlyNamesTheGems(t *testing.T) {
	root := t.TempDir()
	old := push
	defer func() { push = old }()
	push = func(Gem) error { t.Error("a dry run pushed"); return nil }
	res, err := Publish(root, release(t, root), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Published) != 3 || len(res.Skipped) != 1 {
		t.Errorf("result = %+v", res)
	}
}
