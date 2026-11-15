package nugettool

import (
	"errors"
	"strings"
	"testing"
)

// fakePush replaces the push for one test.
func fakePush(t *testing.T, f func(Package) error) {
	t.Helper()
	original := push
	push = f
	t.Cleanup(func() { push = original })
}

func TestPublishPushesPlatformPackagesBeforeTheToolAndSkipsWhatTheFeedHas(t *testing.T) {
	root := t.TempDir()
	var pushed []string
	fakePush(t, func(p Package) error {
		pushed = append(pushed, p.ID)
		if p.ID == "acme.win-x64" {
			return ErrExists
		}
		return nil
	})
	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(pushed, " "); got != "acme.linux-arm64 acme.win-x64 acme" {
		t.Errorf("pushed = %s", got)
	}
	if strings.Join(res.Published, " ") != "acme.linux-arm64.5.0.0.nupkg acme.5.0.0.nupkg" || strings.Join(res.Existing, " ") != "acme.win-x64.5.0.0.nupkg" || strings.Join(res.Skipped, " ") != "freebsd/amd64" {
		t.Errorf("result = %+v", res)
	}
}

func TestPublishStopsAtAFailedPush(t *testing.T) {
	root := t.TempDir()
	fakePush(t, func(p Package) error { return errors.New("403 Forbidden") })
	if _, err := Publish(root, release(t, root), false); err == nil {
		t.Error("no error")
	}
}

func TestPublishDryRunOnlyNamesThePackages(t *testing.T) {
	root := t.TempDir()
	fakePush(t, func(p Package) error { t.Errorf("pushed %s in a dry run", p.ID); return nil })
	res, err := Publish(root, release(t, root), true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(res.Published, " ") != "acme.linux-arm64.5.0.0.nupkg acme.win-x64.5.0.0.nupkg acme.5.0.0.nupkg" || len(res.Skipped) != 1 {
		t.Errorf("result = %+v", res)
	}
}
