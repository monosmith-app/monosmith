package tauri

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// PrepareRequest is one project's sidecars and resources to get ready.
type PrepareRequest struct {
	WorkspaceRoot string
	ProjectRoot   string
	Extras        Extras
	// Triple is the target triple the app is built for.
	Triple string
	// Program returns the built program of a sidecar that names a project.
	Program func(Sidecar) (string, error)
}

// PrepareExtras copies each sidecar to <name>-<triple>[.exe] and each resource
// to its place, under PrepareDir, and writes tauri.config.json there: the
// bundle.externalBin and bundle.resources entries the developer's
// tauri.conf.json lacks, which `tauri build --config` merges. What the
// developer wrote stays; a sidecar the developer already lists is copied to
// the path they gave. It returns a line for each thing it did.
func PrepareExtras(req PrepareRequest) ([]string, error) {
	prep := filepath.Join(req.WorkspaceRoot, filepath.FromSlash(PrepareDir(req.ProjectRoot)))
	if err := os.RemoveAll(prep); err != nil {
		return nil, err
	}
	srcTauri := filepath.Join(req.WorkspaceRoot, filepath.FromSlash(req.ProjectRoot), SrcDir)
	conf, err := readBundle(srcTauri)
	if err != nil {
		return nil, err
	}
	var done []string
	bin, back := conf.externalBin, fromSrcTauri(req.ProjectRoot)+PrepareDir(req.ProjectRoot)
	for _, s := range req.Extras.Sidecars {
		src, err := sidecarSource(req, s)
		if err != nil {
			return nil, err
		}
		declared := conf.declaredPath(s.Name)
		dest := filepath.Join(prep, "binaries", s.Name)
		if declared != "" {
			dest = filepath.Join(srcTauri, filepath.FromSlash(declared))
		} else {
			bin = append(bin, back+"/binaries/"+s.Name)
		}
		dest += "-" + req.Triple + exeSuffix(req.Triple)
		if err := copyFile(src, dest, 0o755); err != nil {
			return nil, err
		}
		done = append(done, fmt.Sprintf("sidecar %s: %s -> %s", s.Name, rel(req.WorkspaceRoot, src), rel(req.WorkspaceRoot, dest)))
	}
	resources := conf.resources
	if len(req.Extras.Resources) > 0 {
		stage := filepath.Join(prep, "resources")
		for _, r := range req.Extras.Resources {
			n, err := copyResource(req.WorkspaceRoot, stage, r)
			if err != nil {
				return nil, err
			}
			done = append(done, fmt.Sprintf("resource %s: %d file(s)", r.From, n))
		}
		if conf.resourceList {
			return nil, fmt.Errorf("%s/tauri.conf.json lists bundle.resources as an array; smith adds its resources to the map form ({ \"folder/\": \"target/\" }): write yours that way, or declare them all in smith.toml", filepath.ToSlash(path.Join(req.ProjectRoot, SrcDir)))
		}
		if resources == nil {
			resources = map[string]any{}
		}
		resources[back+"/resources"] = "."
	}
	bundle := map[string]any{}
	if len(bin) > len(conf.externalBin) {
		bundle["externalBin"] = bin
	}
	if len(req.Extras.Resources) > 0 {
		bundle["resources"] = resources
	}
	out, _ := json.MarshalIndent(map[string]any{"bundle": bundle}, "", "  ")
	if err := os.MkdirAll(prep, 0o755); err != nil {
		return nil, err
	}
	return done, os.WriteFile(filepath.Join(prep, "tauri.config.json"), append(out, '\n'), 0o644)
}

// bundleConf is the part of the developer's tauri.conf.json that smith adds to.
type bundleConf struct {
	externalBin  []string
	resources    map[string]any
	resourceList bool
}

// declaredPath is the externalBin entry whose file name is name, or "".
func (b bundleConf) declaredPath(name string) string {
	for _, e := range b.externalBin {
		if path.Base(strings.ReplaceAll(e, `\`, "/")) == name {
			return e
		}
	}
	return ""
}

func readBundle(srcTauri string) (bundleConf, error) {
	var b bundleConf
	raw, err := os.ReadFile(filepath.Join(srcTauri, "tauri.conf.json"))
	if os.IsNotExist(err) {
		return b, nil
	}
	if err != nil {
		return b, err
	}
	var conf struct {
		Bundle struct {
			ExternalBin []string        `json:"externalBin"`
			Resources   json.RawMessage `json:"resources"`
		} `json:"bundle"`
	}
	if err := json.Unmarshal(raw, &conf); err != nil {
		return b, fmt.Errorf("%s/tauri.conf.json is not valid JSON: %w", filepath.ToSlash(srcTauri), err)
	}
	b.externalBin = conf.Bundle.ExternalBin
	if len(conf.Bundle.Resources) > 0 {
		if json.Unmarshal(conf.Bundle.Resources, &b.resources) != nil {
			b.resourceList = true
		}
	}
	return b, nil
}

var envRef = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`)

// ExpandSource turns a `from` into an absolute path: $NAME and ${NAME} are
// environment variables (an unset one is an error that names it), and a
// relative path is relative to the workspace.
func ExpandSource(workspaceRoot, from string) (string, error) {
	var missing string
	out := envRef.ReplaceAllStringFunc(from, func(m string) string {
		name := strings.Trim(m, "${}")
		v := os.Getenv(name)
		if v == "" && missing == "" {
			missing = name
		}
		return v
	})
	if missing != "" {
		return "", fmt.Errorf("%q needs the environment variable %s, which is not set", from, missing)
	}
	if !filepath.IsAbs(out) {
		out = filepath.Join(workspaceRoot, filepath.FromSlash(out))
	}
	return filepath.Clean(out), nil
}

func sidecarSource(req PrepareRequest, s Sidecar) (string, error) {
	if s.Project != "" {
		return req.Program(s)
	}
	src, err := ExpandSource(req.WorkspaceRoot, s.From)
	if err != nil {
		return "", fmt.Errorf("sidecar %s: %w", s.Name, err)
	}
	// A program built for Windows is name.exe, whether or not the setting says so.
	for _, candidate := range []string{src, src + ".exe"} {
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("sidecar %s: %s does not exist (build it first, or fix from)", s.Name, rel(req.WorkspaceRoot, src))
}

// copyResource copies one resource into stage and returns how many files.
func copyResource(workspaceRoot, stage string, r Resource) (int, error) {
	from := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(r.From, "/*"), "/**"), "/")
	src, err := ExpandSource(workspaceRoot, from)
	if err != nil {
		return 0, fmt.Errorf("resource: %w", err)
	}
	matches := []string{src}
	if strings.ContainsAny(src, "*?[") {
		if matches, err = filepath.Glob(src); err != nil {
			return 0, fmt.Errorf("resource %s: %w", r.From, err)
		}
		if len(matches) == 0 {
			return 0, fmt.Errorf("resource %s: nothing matches", r.From)
		}
	}
	to := r.To
	if to == "" && len(matches) == 1 && !strings.ContainsAny(src, "*?[") {
		to = filepath.Base(src)
	}
	count := 0
	for _, m := range matches {
		st, err := os.Stat(m)
		if err != nil {
			return 0, fmt.Errorf("resource %s: %s does not exist", r.From, rel(workspaceRoot, m))
		}
		dest := filepath.Join(stage, filepath.FromSlash(to))
		switch {
		case st.IsDir():
			if len(matches) > 1 {
				dest = filepath.Join(dest, filepath.Base(m))
			}
		case strings.HasSuffix(to, "/") || len(matches) > 1:
			dest = filepath.Join(dest, filepath.Base(m))
		}
		n, err := copyTree(m, dest)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, fmt.Errorf("resource %s: %s holds no files, so there is nothing to bundle", r.From, rel(workspaceRoot, m))
		}
		count += n
	}
	return count, nil
}

// copyTree copies a file or a folder's files to dest and counts the files.
func copyTree(src, dest string) (int, error) {
	n := 0
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		r, _ := filepath.Rel(src, p)
		n++
		return copyFile(p, filepath.Join(dest, r), 0o644)
	})
	return n, err
}

func copyFile(src, dest string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if st, err := in.Stat(); err == nil && st.Mode()&0o111 != 0 {
		perm = 0o755
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func rel(workspaceRoot, p string) string {
	if r, err := filepath.Rel(workspaceRoot, p); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(p)
}
