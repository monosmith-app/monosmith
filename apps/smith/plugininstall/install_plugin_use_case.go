package plugininstall

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/pluginarchive"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
)

// Install fetches the plugin, verifies it and installs it into
// req.PluginsDir/<name>. Nothing is installed unless every check passes:
//
//   - a digest the person gave (or the release publishes) must match the archive;
//   - an archive needs a checksums file that lists exactly its files and
//     matches every one (a directory's is checked when it has one);
//   - names stay inside the install directory, links and special files are
//     refused, sizes are bounded (see pluginarchive.DefaultLimits and
//     MaxDownloadBytes);
//   - req.Inspect accepts the files.
//
// The files are extracted to a staging directory beside the destination and
// swapped in at the end, so a failure leaves any installed plugin untouched.
// The plugin is never started.
func Install(ctx context.Context, client *http.Client, req Request) (Result, error) {
	src, err := ParseSource(req.Source)
	if err != nil {
		return Result{}, err
	}
	res := Result{Source: src}
	want, wantFrom := plugintrust.NormalizeDigest(req.SHA256), "--sha256"

	var entries []pluginarchive.Entry
	requireChecksums := true
	var cleanup []func()
	defer func() {
		for _, c := range cleanup {
			c()
		}
	}()

	switch src.Kind {
	case KindLocal:
		info, err := os.Stat(src.Location)
		if err != nil {
			return Result{}, fmt.Errorf("cannot read %s: %w", src.Location, err)
		}
		if info.IsDir() {
			if want != "" {
				return Result{}, fmt.Errorf("--sha256 is the digest of an archive; %s is a directory", src.Location)
			}
			requireChecksums = false
			entries, err = pluginarchive.DirEntries(src.Location, func(rel string, isDir bool) bool {
				return isDir && (rel == ".git" || rel == ".hg" || rel == ".svn")
			})
			if err != nil {
				return Result{}, err
			}
			break
		}
		f, err := os.Open(src.Location)
		if err != nil {
			return Result{}, err
		}
		cleanup = append(cleanup, func() { _ = f.Close() })
		entries, err = archiveEntries(f, want, wantFrom, &res)
		if err != nil {
			return Result{}, err
		}
	case KindURL, KindGitHub:
		location := src.Location
		if src.Kind == KindGitHub {
			asset, err := resolveRelease(ctx, client, src)
			if err != nil {
				return Result{}, err
			}
			location = asset.URL
			if want == "" && asset.SHA256 != "" {
				want, wantFrom = asset.SHA256, "the release's "+asset.Name+".sha256"
			}
		}
		tmp, err := os.CreateTemp("", "smith-plugin-*"+pluginarchive.Suffix)
		if err != nil {
			return Result{}, err
		}
		cleanup = append(cleanup, func() { _ = tmp.Close(); _ = os.Remove(tmp.Name()) })
		if _, _, err := download(ctx, client, location, "", MaxDownloadBytes, tmp); err != nil {
			return Result{}, err
		}
		entries, err = archiveEntries(tmp, want, wantFrom, &res)
		if err != nil {
			return Result{}, err
		}
	}

	if err := os.MkdirAll(req.PluginsDir, 0o755); err != nil {
		return Result{}, err
	}
	stage, err := os.MkdirTemp(req.PluginsDir, ".install-")
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	ex, err := pluginarchive.Extract(entries, stage, pluginarchive.Limits{}, requireChecksums)
	if err != nil {
		return Result{}, err
	}
	res.Files, res.Bytes, res.Checksummed = ex.Files, ex.Bytes, ex.Checksummed

	p, err := req.Inspect(stage)
	if err != nil {
		return Result{}, err
	}
	if err := pluginarchive.ValidatePluginName(p.Name); err != nil {
		return Result{}, err
	}
	res.Plugin = p
	res.Dir = filepath.Join(req.PluginsDir, p.Name)
	res.Replaced, err = swapIn(stage, res.Dir, req.PluginsDir, req.Replace)
	return res, err
}

// archiveEntries hashes the archive file, checks the expected digest and lists
// the files in it.
func archiveEntries(f *os.File, want, wantFrom string, res *Result) ([]pluginarchive.Entry, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return nil, err
	}
	res.ArchiveSHA256 = hex.EncodeToString(h.Sum(nil))
	if want != "" {
		if res.ArchiveSHA256 != want {
			return nil, &DigestMismatchError{Want: want, From: wantFrom, Got: res.ArchiveSHA256}
		}
		res.ArchiveVerified = true
	}
	return pluginarchive.ZipEntries(f, size)
}

// swapIn moves the staged directory to target. An installed plugin is kept
// aside until the move has worked, and put back if it has not.
func swapIn(stage, target, parent string, replace bool) (replaced bool, err error) {
	_, statErr := os.Lstat(target)
	exists := statErr == nil
	if exists && !replace {
		return false, fmt.Errorf("%w: %s; use --force to replace it", ErrAlreadyInstalled, target)
	}
	var trash string
	if exists {
		if trash, err = os.MkdirTemp(parent, ".replaced-"); err != nil {
			return false, err
		}
		defer func() { _ = os.RemoveAll(trash) }()
		if err := os.Rename(target, filepath.Join(trash, "old")); err != nil {
			return false, fmt.Errorf("cannot replace %s (is something using it?): %w", target, err)
		}
	}
	if err := os.Rename(stage, target); err != nil {
		if exists {
			_ = os.Rename(filepath.Join(trash, "old"), target)
		}
		return false, err
	}
	return exists, nil
}
