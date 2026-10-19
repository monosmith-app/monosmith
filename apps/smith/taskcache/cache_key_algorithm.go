package taskcache

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
)

// KeySchema versions the key encoding. Changing what goes into a key, or how,
// requires a new schema so old entries can never be mistaken for new ones.
const KeySchema = "smith-task-key-v1"

// FileDigest is one input file: workspace-relative path and content digest.
type FileDigest struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

// KeyInputs is everything that can change a task's result.
// Each cache entry keeps the inputs its key was computed from, so a later miss
// can say what changed.
type KeyInputs struct {
	Task string `json:"task"`
	// Command is the task's argv before placeholders are expanded, so the key
	// does not depend on where the workspace is checked out.
	Command []string          `json:"command"`
	Env     map[string]string `json:"env,omitempty"`
	Files   []FileDigest      `json:"files"`
	// Upstream are digests of the sources of every project this one depends
	// on, transitively, by project name.
	Upstream map[string]string `json:"upstream,omitempty"`
	// Deps are the keys of the tasks this one depends on, by task id. Only
	// the keys go into this key.
	Deps     map[string]string `json:"deps,omitempty"`
	Platform string            `json:"platform"`
	// Tool is the version of smith and of every plugin that shapes the
	// project ("0.2.0/node@0.2.0+jest@0.2.0").
	Tool string `json:"tool"`
	// Toolchain is the version of the compiler or runtime the task runs with
	// ("node 20.11.0"), as its language plugin probed it; empty when it
	// cannot tell. It joins the key only when set, so a plugin that reports
	// nothing keeps the keys it always had.
	Toolchain string `json:"toolchain,omitempty"`
	// Serve is the app the task runs against ("web:start http://127.0.0.1:5173");
	// empty for a task that serves none, and then it is left out of the key.
	Serve string `json:"serve,omitempty"`
}

// Key computes the cache key: a SHA-256 over a length-prefixed encoding of
// each field, with every collection sorted.
func Key(in KeyInputs) string {
	h := sha256.New()
	write := func(s string) {
		var n [8]byte
		binary.BigEndian.PutUint64(n[:], uint64(len(s)))
		h.Write(n[:])
		h.Write([]byte(s))
	}
	section := func(name string, count int) {
		write(name)
		write(string(rune(count)))
	}
	write(KeySchema)
	write(in.Task)
	write(in.Platform)
	write(in.Tool)
	if in.Toolchain != "" {
		write("toolchain")
		write(in.Toolchain)
	}
	if in.Serve != "" {
		write("serve")
		write(in.Serve)
	}
	section("command", len(in.Command))
	for _, a := range in.Command {
		write(a)
	}
	section("env", len(in.Env))
	for _, k := range sortedKeys(in.Env) {
		write(k)
		write(in.Env[k])
	}
	files := append([]FileDigest(nil), in.Files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	section("files", len(files))
	for _, f := range files {
		write(f.Path)
		write(f.Digest)
	}
	section("upstream", len(in.Upstream))
	for _, k := range sortedKeys(in.Upstream) {
		write(k)
		write(in.Upstream[k])
	}
	deps := make([]string, 0, len(in.Deps))
	for _, k := range in.Deps {
		deps = append(deps, k)
	}
	sort.Strings(deps)
	section("deps", len(deps))
	for _, d := range deps {
		write(d)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Combine digests a list of file digests into one, for a project's sources.
func Combine(files []FileDigest) string {
	return Key(KeyInputs{Task: "sources", Files: files})
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
