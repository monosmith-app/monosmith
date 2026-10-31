package tauriadd

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Result is what AddTauri did.
type Result struct {
	// Project is the project's name and Stack the stack it has now.
	Project string `json:"project"`
	Stack   string `json:"stack"`
	// PreviousStack is the stack it had; empty when it had none.
	PreviousStack string `json:"previousStack,omitempty"`
	// DeclaredStack says the stack was written to smith.toml.
	DeclaredStack bool `json:"declaredStack"`
	// SrcTauriExisted says the project already had a src-tauri, left as it was.
	SrcTauriExisted bool `json:"srcTauriExisted"`
	// Files are the files written or edited, workspace-relative.
	Files []string `json:"files"`
}

// AddTauri makes an existing Node app a Tauri app: its stack gains the tauri
// toolchain (in smith.toml), then the tauri plugin writes src-tauri, adds the
// CLI, the API package and the tauri script, and keeps src-tauri out of a root
// Cargo workspace, never overwriting a file that exists. Dependencies are
// installed and the pipelines regenerated. When it fails smith.toml is put back.
func AddTauri(s *session.Session, name string) (res Result, err error) {
	p := s.Graph.Get(name)
	if p == nil {
		return res, problem.New("the workspace has no project "+name, "smith projects lists them")
	}
	if p.Language != "node" {
		return res, problem.New(name+" is a "+p.Language+" project; Tauri wraps a Node app", "smith add app "+name+"-desktop --stack react-tauri makes a new one")
	}
	if p.Kind != protocol.ProjectApp {
		return res, problem.New(name+" is a "+string(p.Kind)+", not an app", "smith kind "+name+" app")
	}
	if _, statErr := os.Stat(filepath.Join(s.Root, filepath.FromSlash(p.Root), "index.html")); statErr != nil {
		return res, problem.New(p.Root+" has no index.html: there is no web page for Tauri to wrap", "start from a web app (smith add app "+name+" --stack react), or write src-tauri with `npx tauri init`")
	}
	res = Result{Project: name, SrcTauriExisted: tauriapp.Detect(s.Root, p.Root)}

	var current protocol.Stack
	if cur := projectgraph.StackOf(p, s.Config); cur != "" {
		res.PreviousStack = cur
		if st, getErr := s.Stacks.Get(cur); getErr == nil {
			current = st
		} else if cur != protocol.NoStack {
			return res, problem.New("the stack "+cur+" of "+name+" cannot be used: "+getErr.Error(), "smith stacks says what is wrong with it")
		}
	}
	st, declare := DeriveStack(current, name, func(n string) (protocol.Stack, bool) {
		found, getErr := s.Stacks.Get(n)
		return found, getErr == nil
	})
	if err := pluginhost.ValidateStack(st, s.Plugins); err != nil {
		return res, problem.New("the stack "+st.Name+" would not be valid: "+err.Error(), "name the stack yourself in smith.toml with tauri last in its toolchain")
	}
	res.Stack, res.DeclaredStack = st.Name, declare

	configFile := filepath.Join(s.Root, workspace.FileName)
	original, err := os.ReadFile(configFile)
	if err != nil {
		return res, err
	}
	defer func() {
		if err != nil {
			_ = os.WriteFile(configFile, original, 0o644)
			_ = s.Refresh()
		}
	}()
	if declare {
		text := string(original)
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += "\n" + renderStack(st, name)
		if _, err = workspace.Parse([]byte(text)); err != nil {
			return res, problem.New("the stack " + st.Name + " would break smith.toml: " + err.Error())
		}
		if err = os.WriteFile(configFile, []byte(text), 0o644); err != nil {
			return res, err
		}
		res.Files = append(res.Files, workspace.FileName)
	}
	if st.Name != res.PreviousStack {
		if err = configedit.SetTableValue(s.Root, "projects."+configedit.QuoteKey(name), "stack", st.Name); err != nil {
			return res, err
		}
		if !declare {
			res.Files = append(res.Files, workspace.FileName)
		}
	}
	if err = s.Refresh(); err != nil {
		return res, err
	}

	tauri := s.Plugins.Get(tauriapp.Name)
	if tauri == nil || !tauri.Supports(protocol.MethodGenerate) {
		return res, problem.New("the tauri plugin is not loaded", "enable it: smith config set plugins.enabled with tauri in the list")
	}
	params := protocol.GenerateParams{Kind: p.Kind, Name: name, Directory: p.Root, Stack: st.Name, Workspace: protocol.WorkspaceInfo{
		Name: s.Config.Workspace.Name, Scope: s.Config.Workspace.Scope, ModulePrefix: s.Config.Workspace.ModulePrefix,
	}}
	var gen protocol.GenerateResult
	if err = tauri.Call(protocol.MethodGenerate, params, &gen, 20*time.Minute); err != nil {
		return res, err
	}
	res.Files = append(res.Files, gen.Files...)
	if gen.InstallPending {
		language, langErr := s.Plugins.ForLanguage(p.Language)
		if langErr != nil || !language.Supports(protocol.MethodInstall) {
			return res, problem.New("the dependencies are added but cannot be installed", "run smith install")
		}
		var installed protocol.DependencyResult
		if err = language.Call(protocol.MethodInstall, protocol.InstallParams{Quiet: true}, &installed, 20*time.Minute); err != nil {
			return res, problem.New("installing the dependencies failed: "+err.Error(), "run smith install to see why")
		}
		res.Files = append(res.Files, installed.Files...)
	}
	if err = s.Refresh(); err != nil {
		return res, err
	}
	changed, _, err := cipipeline.Regenerate(s)
	if err != nil {
		return res, err
	}
	res.Files = append(res.Files, changed...)
	sort.Strings(res.Files)
	res.Files = slices.Compact(res.Files)
	return res, nil
}
