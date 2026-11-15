package pypiwheels

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestPublishUploadsEachWheelAndKeepsWhatTheIndexHas(t *testing.T) {
	root := t.TempDir()
	r := release(t, root)
	r.Archives = append(r.Archives, Archive{Path: archive(t, root, "acme_linux_arm64.zip", "acme"), OS: "linux", Arch: "arm64"})
	original := upload
	t.Cleanup(func() { upload = original })
	upload = func(w Wheel) error {
		if filepath.Base(w.File) == "acme-5.0.0-py3-none-win_amd64.whl" {
			return ErrExists
		}
		return nil
	}
	res, err := Publish(root, r, false)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res.Published, []string{"acme-5.0.0-py3-none-manylinux_2_17_aarch64.manylinux2014_aarch64.musllinux_1_1_aarch64.whl"}) {
		t.Errorf("published %v", res.Published)
	}
	if !slices.Equal(res.Existing, []string{"acme-5.0.0-py3-none-win_amd64.whl"}) || !slices.Equal(res.Skipped, []string{"freebsd/amd64"}) {
		t.Errorf("existing %v skipped %v", res.Existing, res.Skipped)
	}
}

func TestDryRunNamesTheWheels(t *testing.T) {
	root := t.TempDir()
	original := upload
	t.Cleanup(func() { upload = original })
	upload = func(Wheel) error { t.Error("uploaded on a dry run"); return nil }
	r := release(t, root)
	r.Version = "5.0.0-rc.1"
	res, err := Publish(root, r, true)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res.Published, []string{"acme-5.0.0rc1-py3-none-win_amd64.whl"}) {
		t.Errorf("published %v", res.Published)
	}
}
