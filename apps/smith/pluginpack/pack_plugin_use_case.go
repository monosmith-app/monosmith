package pluginpack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginarchive"
)

// ArchiveName is the file name a plugin is packed to.
func ArchiveName(p Plugin) string {
	return p.Name + "-" + p.Version + pluginarchive.Suffix
}

// Pack packs the plugin directory dir into outDir as <name>-<version>.smith-plugin.zip,
// with <that file>.sha256 beside it. Every file of the directory goes in,
// except version-control folders, earlier archives and the checksums file,
// which is written fresh. The files the manifest's command names for any
// platform are marked executable, so a build on Windows still installs
// runnable on Linux. Nothing is run.
func Pack(dir string, p Plugin, outDir string) (Result, error) {
	if err := pluginarchive.ValidatePluginName(p.Name); err != nil {
		return Result{}, err
	}
	if err := pluginarchive.ValidatePluginVersion(p.Version); err != nil {
		return Result{}, err
	}
	entries, err := pluginarchive.DirEntries(dir, skipped)
	if err != nil {
		return Result{}, err
	}
	if len(entries) == 0 {
		return Result{}, fmt.Errorf("%s has no files to pack", dir)
	}
	var programs []string
	if len(p.Command) > 0 {
		if re, ok := programPattern(p.Command[0]); ok {
			for i := range entries {
				if re.MatchString(entries[i].Name) {
					entries[i].Executable = true
					programs = append(programs, entries[i].Name)
				}
			}
			if len(programs) == 0 {
				return Result{}, fmt.Errorf("plugin.json runs %s, but no file in %s matches it; build the program for each platform first", p.Command[0], dir)
			}
		}
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name)
	}
	sort.Strings(names)
	sort.Strings(programs)

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return Result{}, err
	}
	final := filepath.Join(outDir, ArchiveName(p))
	tmp, err := os.CreateTemp(outDir, ".pack-*.tmp")
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	h := sha256.New()
	counter := &countingWriter{}
	if _, err := pluginarchive.Write(io.MultiWriter(tmp, h, counter), entries); err != nil {
		_ = tmp.Close()
		return Result{}, err
	}
	if err := tmp.Close(); err != nil {
		return Result{}, err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return Result{}, err
	}
	if err := replaceFile(tmp.Name(), final); err != nil {
		return Result{}, err
	}
	digest := hex.EncodeToString(h.Sum(nil))
	sidecar := final + ".sha256"
	if err := os.WriteFile(sidecar, []byte(digest+"  "+filepath.Base(final)+"\n"), 0o644); err != nil {
		return Result{}, err
	}
	return Result{Path: final, DigestPath: sidecar, SHA256: digest, Files: names, Programs: programs, Bytes: counter.n}, nil
}

// skipped leaves out what is not part of a plugin: version-control data, an
// earlier archive and its digest, and the checksums file smith rewrites.
func skipped(rel string, isDir bool) bool {
	base := rel[strings.LastIndex(rel, "/")+1:]
	switch {
	case isDir:
		return base == ".git" || base == ".hg" || base == ".svn"
	case rel == pluginarchive.ChecksumsFile, base == ".DS_Store":
		return true
	}
	return strings.HasSuffix(base, pluginarchive.Suffix) || strings.HasSuffix(base, pluginarchive.Suffix+".sha256")
}

type countingWriter struct{ n int64 }

func (c *countingWriter) Write(b []byte) (int, error) {
	c.n += int64(len(b))
	return len(b), nil
}

// replaceFile moves src over dst; Windows cannot rename over a file that exists.
func replaceFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	_ = os.Remove(dst)
	return os.Rename(src, dst)
}
