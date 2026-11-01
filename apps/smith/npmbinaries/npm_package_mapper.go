package npmbinaries

import (
	"encoding/json"
	"fmt"
	"strings"
)

// npmOS and npmCPU spell Go's os and arch the way Node's process.platform
// and process.arch do, which npm's os and cpu fields match.
var (
	npmOS  = map[string]string{"linux": "linux", "darwin": "darwin", "windows": "win32", "freebsd": "freebsd"}
	npmCPU = map[string]string{"amd64": "x64", "arm64": "arm64", "386": "ia32", "arm": "arm"}
)

// PlatformPackage names the package holding one platform's binary: the
// first npm name with the platform appended (@acme/cli-linux-x64).
func PlatformPackage(base, goos, goarch string) (string, bool) {
	os, okOS := npmOS[goos]
	cpu, okCPU := npmCPU[goarch]
	if !okOS || !okCPU {
		return "", false
	}
	return base + "-" + os + "-" + cpu, true
}

// executable is the binary's file name on goos.
func executable(binary, goos string) string {
	if goos == "windows" {
		return binary + ".exe"
	}
	return binary
}

type repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// manifest is a package.json; the field order is the order npm shows.
type manifest struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Description          string            `json:"description,omitempty"`
	License              string            `json:"license,omitempty"`
	Homepage             string            `json:"homepage,omitempty"`
	Repository           *repository       `json:"repository,omitempty"`
	OS                   []string          `json:"os,omitempty"`
	CPU                  []string          `json:"cpu,omitempty"`
	PreferUnplugged      bool              `json:"preferUnplugged,omitempty"`
	Bin                  map[string]string `json:"bin,omitempty"`
	Files                []string          `json:"files"`
	Engines              map[string]string `json:"engines,omitempty"`
	OptionalDependencies map[string]string `json:"optionalDependencies,omitempty"`
}

func (r Release) repository() *repository {
	if r.Repository == "" {
		return nil
	}
	return &repository{Type: "git", URL: "git+" + r.Repository + ".git"}
}

func render(m manifest) string {
	raw, _ := json.MarshalIndent(m, "", "  ")
	return string(raw) + "\n"
}

// PlatformManifest is the package.json of one platform's package: npm
// installs it only on a matching os and cpu.
func PlatformManifest(r Release, name, goos, goarch string) string {
	return render(manifest{
		Name: name, Version: r.Version,
		Description: fmt.Sprintf("The %s %s binary of %s", npmOS[goos], npmCPU[goarch], r.Names[0]),
		License:     r.License, Homepage: r.Homepage, Repository: r.repository(),
		OS: []string{npmOS[goos]}, CPU: []string{npmCPU[goarch]}, PreferUnplugged: true,
		Files: []string{"bin"},
	})
}

// WrapperManifest is the package.json of a package people install: its bin
// runs the binary that one of its optional dependencies brought.
func WrapperManifest(r Release, name string, platforms map[string]string) string {
	deps := map[string]string{}
	for _, pkg := range platforms {
		deps[pkg] = r.Version
	}
	return render(manifest{
		Name: name, Version: r.Version, Description: r.Description,
		License: r.License, Homepage: r.Homepage, Repository: r.repository(),
		Bin:   map[string]string{r.Binary: "bin/" + r.Binary + ".js"},
		Files: []string{"bin"}, Engines: map[string]string{"node": ">=18"},
		OptionalDependencies: deps,
	})
}

// Shim is the wrapper's bin: it finds the platform package npm installed
// and runs its binary with the same arguments, exit code and signal.
func Shim(binary string, platforms map[string]string) string {
	raw, _ := json.MarshalIndent(platforms, "", "  ")
	return strings.NewReplacer("__BINARY__", binary, "__PLATFORMS__", string(raw)).Replace(`#!/usr/bin/env node
'use strict'
// Runs the __BINARY__ binary npm installed for this platform, from one of the
// optional dependencies; nothing is downloaded after install.
const { spawn } = require('node:child_process')
const fs = require('node:fs')

const platforms = __PLATFORMS__
const key = process.platform + '-' + process.arch
const pkg = platforms[key]
if (!pkg) {
  console.error('__BINARY__ has no build for ' + key)
  process.exit(1)
}
const file = process.platform === 'win32' ? '__BINARY__.exe' : '__BINARY__'
let binary
try {
  binary = require.resolve(pkg + '/bin/' + file)
} catch {
  console.error('__BINARY__: ' + pkg + ' is not installed; reinstall without --omit=optional or --no-optional')
  process.exit(1)
}
if (process.platform !== 'win32') {
  try {
    fs.accessSync(binary, fs.constants.X_OK)
  } catch {
    try { fs.chmodSync(binary, 0o755) } catch { /* a read-only install keeps its mode */ }
  }
}
// A stop request must reach the binary: forward it, never orphan the child.
const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' })
for (const sig of ['SIGINT', 'SIGTERM', 'SIGHUP']) {
  process.on(sig, () => { child.kill(sig) })
}
child.on('error', (error) => {
  console.error('__BINARY__: ' + error.message)
  process.exit(1)
})
child.on('exit', (code, signal) => {
  if (signal) {
    process.removeAllListeners(signal)
    process.kill(process.pid, signal)
    return
  }
  process.exit(code ?? 1)
})
`)
}

// Readme is the wrapper's README.md, shown on its npm page.
func Readme(r Release, name string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", name)
	if r.Description != "" {
		b.WriteString(r.Description + "\n\n")
	}
	fmt.Fprintf(&b, "```sh\nnpm install --global %s\n%s --version\n```\n\n", name, r.Binary)
	fmt.Fprintf(&b, "The package holds no code of its own: npm installs the prebuilt %s binary for your platform from one of its optional dependencies (%s-<os>-<cpu>), so nothing is downloaded after install and it works behind a proxy or offline from a cache.\n", r.Binary, r.Names[0])
	if r.Homepage != "" {
		fmt.Fprintf(&b, "\nMore: %s\n", r.Homepage)
	}
	return b.String()
}
