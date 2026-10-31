package extras

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Findings checks, for each Tauri project, that every sidecar and
// resource its settings.tauri declares can be found, and that every
// bundle.externalBin entry smith does not manage has a program named for this
// machine's target triple (`<name>-<triple>[.exe]`), which Tauri insists on.
func Findings(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	var findings []protocol.Finding
	for _, p := range projects {
		if !tauriapp.Detect(workspaceRoot, p.Root) {
			continue
		}
		extras, err := ParseExtras(p.Settings[tauriapp.Name])
		if err != nil {
			findings = append(findings, protocol.Finding{
				Check: "the tauri settings of " + p.Name + " are valid", Detail: err.Error(),
				Remedy: "fix [projects." + p.Name + ".settings.tauri] in smith.toml",
			})
			continue
		}
		for _, s := range extras.Sidecars {
			if s.From != "" {
				findings = append(findings, sidecarFinding(workspaceRoot, p, s))
			}
		}
		for _, r := range extras.Resources {
			findings = append(findings, resourceFinding(workspaceRoot, p, r))
		}
		findings = append(findings, externalBinFindings(workspaceRoot, p, extras)...)
	}
	return findings
}

// sidecarFinding checks a sidecar given as a file; one that names a project is
// checked by smith doctor itself, which sees every project.
func sidecarFinding(workspaceRoot string, p protocol.Project, s Sidecar) protocol.Finding {
	f := protocol.Finding{Check: "the sidecar " + s.Name + " of " + p.Name + " exists", OK: true}
	src, err := ExpandSource(workspaceRoot, s.From)
	if err != nil {
		f.OK, f.Warning, f.Detail, f.Remedy = false, true, err.Error(), "set the variable (CI sets it from [[ci.sibling_repos]]) before building"
		return f
	}
	if !fileExists(src) && !fileExists(src+".exe") {
		f.OK, f.Detail, f.Remedy = false, rel(workspaceRoot, src)+" does not exist", "build it first, or fix from in [projects."+p.Name+".settings.tauri]"
		// A build output is made by the build, so its absence is no mistake yet.
		f.Warning = strings.HasPrefix(rel(workspaceRoot, src), protocol.BuildDir+"/")
	}
	return f
}

func resourceFinding(workspaceRoot string, p protocol.Project, r Resource) protocol.Finding {
	f := protocol.Finding{Check: "the resource " + r.From + " of " + p.Name + " exists", OK: true}
	src, err := ExpandSource(workspaceRoot, strings.TrimSuffix(strings.TrimSuffix(r.From, "/*"), "/**"))
	if err != nil {
		f.OK, f.Warning, f.Detail, f.Remedy = false, true, err.Error(), "set the variable (CI sets it from [[ci.sibling_repos]]) before building"
		return f
	}
	_, statErr := os.Stat(src)
	found := statErr == nil
	if strings.ContainsAny(src, "*?[") {
		m, _ := filepath.Glob(src)
		found = len(m) > 0
	}
	if !found {
		f.OK, f.Detail, f.Remedy = false, rel(workspaceRoot, src)+" does not exist or matches nothing", "fix from in [projects."+p.Name+".settings.tauri]"
		f.Warning = strings.HasPrefix(rel(workspaceRoot, src), protocol.BuildDir+"/")
	}
	return f
}

// externalBinFindings warns about an externalBin entry of tauri.conf.json that
// smith does not copy and whose program for this machine's triple is missing.
func externalBinFindings(workspaceRoot string, p protocol.Project, extras Extras) []protocol.Finding {
	srcTauri := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), tauriapp.SrcDir)
	conf, err := readBundle(srcTauri)
	triple, terr := HostTriple()
	if err != nil || terr != nil || triple == "" {
		return nil
	}
	var findings []protocol.Finding
	for _, e := range conf.externalBin {
		name := path.Base(strings.ReplaceAll(e, `\`, "/"))
		if slices.ContainsFunc(extras.Sidecars, func(s Sidecar) bool { return s.Name == name }) {
			continue
		}
		want := filepath.Join(srcTauri, filepath.FromSlash(e)+"-"+triple+exeSuffix(triple))
		findings = append(findings, protocol.Finding{
			Check:   "the sidecar " + name + " of " + p.Name + " is named for this machine",
			OK:      fileExists(want),
			Warning: true,
			Detail:  p.Root + "/" + tauriapp.SrcDir + "/tauri.conf.json lists " + e + " in bundle.externalBin, but " + rel(workspaceRoot, want) + " does not exist; Tauri wants <name>-<target triple>[.exe]",
			Remedy:  "declare it in [projects." + p.Name + ".settings.tauri] sidecars, and smith builds and copies it with the right name",
		})
	}
	return findings
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
