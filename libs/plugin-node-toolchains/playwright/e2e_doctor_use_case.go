package playwright

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lookPath finds a command on PATH; tests replace it.
var lookPath = exec.LookPath

// pipelineFiles are where `smith pipeline e2e` writes the e2e pipeline.
var pipelineFiles = []string{".github/workflows/e2e.yml", "azure-pipelines/e2e.yml"}

// Doctor checks what the e2e tasks need from this machine and the workspace:
// Playwright installed, a virtual display for the suites that open windows,
// and a pipeline that runs them, since smith ci verify leaves e2e out.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	var display []string
	suites, installed := 0, exists(workspaceRoot, "node_modules/playwright-core/package.json")
	for _, p := range params.Projects {
		_, hasScript := p.Tasks[TaskName]
		suite := Detect(workspaceRoot, p.Root, hasScript)
		if suite.Flavour == None {
			continue
		}
		suites++
		// npm hoists playwright-core to the root, unless versions conflict.
		installed = installed || exists(workspaceRoot, p.Root+"/node_modules/playwright-core/package.json")
		if suite.Options.Display {
			display = append(display, p.Name)
		}
	}
	findings := []protocol.Finding{}
	if suites == 0 {
		return protocol.DoctorResult{Findings: findings}
	}
	findings = append(findings, protocol.Finding{
		Check:  "Playwright is installed",
		OK:     installed,
		Detail: "node_modules has no playwright-core, so the e2e tasks cannot start a browser",
		Remedy: "npm ci (or npm install -D @playwright/test in the project)",
	})
	if len(display) > 0 && platform == "linux" {
		sort.Strings(display)
		_, err := lookPath("xvfb-run")
		findings = append(findings, protocol.Finding{
			Check:  "a virtual display for e2e suites that open windows",
			OK:     err == nil,
			Detail: "xvfb-run is not on PATH; " + strings.Join(display, ", ") + " run their e2e task under it",
			Remedy: "sudo apt-get install -y xvfb",
		})
	}
	pipeline := false
	for _, f := range pipelineFiles {
		pipeline = pipeline || exists(workspaceRoot, f)
	}
	findings = append(findings, protocol.Finding{
		Check:   "a pipeline runs the e2e tasks",
		OK:      pipeline,
		Warning: true,
		Detail:  "smith ci verify leaves e2e out, and no e2e pipeline is committed",
		Remedy:  "smith pipeline e2e",
		Fix: &protocol.Fix{Description: "Write the e2e pipeline", Steps: []protocol.FixStep{{
			Description: "smith pipeline e2e writes the pipeline that runs the e2e tasks",
			Argv:        []string{"smith", "pipeline", "e2e"},
		}}},
	})
	return protocol.DoctorResult{Findings: findings}
}

func exists(workspaceRoot, rel string) bool {
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	return err == nil
}
