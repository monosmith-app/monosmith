package npmbinaries

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
)

func archive(t *testing.T, root, name, file string) string {
	t.Helper()
	f, err := os.Create(filepath.Join(root, name))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create(file)
	_, _ = w.Write([]byte("binary " + file))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return name
}

func release(t *testing.T, root string) Release {
	return Release{
		Names: []string{"@acme/cli", "acme"}, Binary: "acme", Version: "5.0.0", OutDir: "out",
		Archives: []Archive{
			{Path: archive(t, root, "acme_windows_amd64.zip", "acme.exe"), OS: "windows", Arch: "amd64"},
			{Path: archive(t, root, "acme_linux_amd64.zip", "acme"), OS: "linux", Arch: "amd64"},
			{Path: archive(t, root, "acme_plan9_amd64.zip", "acme"), OS: "plan9", Arch: "amd64"},
		},
	}
}

// fakeNPM replaces the npm CLI for one test.
func fakeNPM(t *testing.T, f func(dir string, args ...string) (string, string, error)) {
	t.Helper()
	original := npm
	npm = f
	t.Cleanup(func() { npm = original })
}

// fakeAwait replaces the wait for the platform packages with a fixed answer.
func fakeAwait(t *testing.T, ready bool, waitedFor *[]string) {
	t.Helper()
	original := awaitPlatforms
	awaitPlatforms = func(_ string, _ Release, names []string) npmpropagation.Outcome {
		if waitedFor != nil {
			*waitedFor = names
		}
		return npmpropagation.Outcome{Ready: ready}
	}
	t.Cleanup(func() { awaitPlatforms = original })
}

func TestPublishSendsPlatformPackagesBeforeWrappersAndSkipsWhatTheRegistryHas(t *testing.T) {
	root := t.TempDir()
	fakeAwait(t, true, nil)
	t.Setenv("NODE_AUTH_TOKEN", "secret")
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	var published []string
	fakeNPM(t, func(dir string, args ...string) (string, string, error) {
		if args[0] == "view" {
			if args[1] == "@acme/cli-linux-x64@5.0.0" {
				return "5.0.0\n", "", nil
			}
			return "", "", nil
		}
		raw, _ := os.ReadFile(filepath.Join(dir, "package.json"))
		published = append(published, filepath.Base(dir)+" "+strings.Join(args, " "))
		if !strings.Contains(string(raw), `"version": "5.0.0"`) {
			t.Errorf("%s package.json has no version", dir)
		}
		return "", "", nil
	})

	res, err := Publish(root, release(t, root), false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"acme-cli-win32-x64 publish --access public",
		"acme-cli publish --access public",
		"acme publish --access public",
	}
	if !slices.Equal(published, want) {
		t.Errorf("published\n%v\nwant\n%v", published, want)
	}
	if !slices.Equal(res.Existing, []string{"@acme/cli-linux-x64@5.0.0"}) || !slices.Equal(res.Skipped, []string{"plan9/amd64"}) {
		t.Errorf("existing %v skipped %v", res.Existing, res.Skipped)
	}
	bin, err := os.ReadFile(filepath.Join(root, "out", "acme-cli-win32-x64", "bin", "acme.exe"))
	if err != nil || string(bin) != "binary acme.exe" {
		t.Errorf("platform binary = %q, %v", bin, err)
	}
	if _, err := os.Stat(filepath.Join(root, "out", "acme", "bin", "acme.js")); err != nil {
		t.Error("the wrapper has no shim")
	}
	if rc, _ := os.ReadFile(filepath.Join(root, "out", "acme", ".npmrc")); !strings.Contains(string(rc), "${NODE_AUTH_TOKEN}") {
		t.Errorf(".npmrc = %q", rc)
	}
}

func TestPublishAddsProvenanceAndTagsPrereleases(t *testing.T) {
	root := t.TempDir()
	fakeAwait(t, true, nil)
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "https://token")
	var args []string
	fakeNPM(t, func(_ string, a ...string) (string, string, error) {
		if a[0] == "publish" {
			args = a
		}
		return "", "", nil
	})
	r := release(t, root)
	r.Version, r.Repository = "5.0.0-rc.1", "https://github.com/acme/cli"
	if _, err := Publish(root, r, false); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(args, " "); got != "publish --access public --tag next --provenance" {
		t.Errorf("args = %s", got)
	}
}

func TestDryRunNamesThePackagesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	fakeNPM(t, func(string, ...string) (string, string, error) {
		t.Error("npm ran on a dry run")
		return "", "", nil
	})
	res, err := Publish(root, release(t, root), true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"@acme/cli-linux-x64@5.0.0", "@acme/cli-win32-x64@5.0.0", "@acme/cli@5.0.0", "acme@5.0.0"}
	if !slices.Equal(res.Published, want) {
		t.Errorf("published %v", res.Published)
	}
	if _, err := os.Stat(filepath.Join(root, "out")); !os.IsNotExist(err) {
		t.Error("a dry run wrote packages")
	}
}

func TestWrappersTakeThePendingTagWhenThePlatformPackagesAreNotFetchable(t *testing.T) {
	for _, c := range []struct{ version, tag, moveTo string }{
		{"5.0.0-rc.1", "next", "next"}, {"5.0.0", "", "latest"},
	} {
		t.Run(c.version, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
			var waitedFor []string
			fakeAwait(t, false, &waitedFor)
			var tags []string
			fakeNPM(t, func(dir string, a ...string) (string, string, error) {
				if a[0] == "publish" {
					tags = append(tags, filepath.Base(dir)+" "+strings.Join(a, " "))
				}
				return "", "", nil
			})
			r := release(t, root)
			r.Version = c.version
			res, err := Publish(root, r, false)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(waitedFor, []string{"@acme/cli-linux-x64", "@acme/cli-win32-x64"}) {
				t.Errorf("waited for %v", waitedFor)
			}
			wantPlatform := "publish --access public"
			if c.tag != "" {
				wantPlatform += " --tag next"
			}
			want := []string{"acme-cli-linux-x64 " + wantPlatform, "acme-cli-win32-x64 " + wantPlatform, "acme-cli publish --access public --tag pending", "acme publish --access public --tag pending"}
			if !slices.Equal(tags, want) {
				t.Errorf("publishes %v, want %v", tags, want)
			}
			wantTags := []string{"npm dist-tag add @acme/cli@" + c.version + " " + c.moveTo, "npm dist-tag add acme@" + c.version + " " + c.moveTo}
			if !slices.Equal(res.PendingTags, wantTags) {
				t.Errorf("pending tags %v, want %v", res.PendingTags, wantTags)
			}
		})
	}
}

func TestNoWaitWhenEveryWrapperIsAlreadyPublished(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")
	original := awaitPlatforms
	awaitPlatforms = func(string, Release, []string) npmpropagation.Outcome {
		t.Error("waited although nothing is left to publish")
		return npmpropagation.Outcome{}
	}
	t.Cleanup(func() { awaitPlatforms = original })
	fakeNPM(t, func(_ string, a ...string) (string, string, error) {
		if a[0] == "view" {
			return "5.0.0\n", "", nil
		}
		t.Error("published")
		return "", "", nil
	})
	if _, err := Publish(root, release(t, root), false); err != nil {
		t.Fatal(err)
	}
}
