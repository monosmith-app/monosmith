package pipeline

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-cpp/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// toolchain installs CMake, Ninja, a compiler and cppcheck: from the
// distribution on Debian and Ubuntu (the CI images), with Homebrew on a Mac
// runner (the compiler is Xcode's). A Windows runner comes with CMake, Ninja
// and MinGW's compiler, so nothing is installed there. The script runs in
// bash on every runner.
var toolchain = protocol.Toolchain{
	Name: "cpp", Version: "system",
	Install: "case \"$(uname -s)\" in\n" +
		"  Linux)\n" +
		"    SUDO=$(command -v sudo || true)\n" +
		"    $SUDO apt-get update -qq && $SUDO apt-get install -y -qq cmake ninja-build g++ cppcheck git curl zip unzip tar pkg-config ;;\n" +
		"  Darwin)\n" +
		"    brew install cmake ninja cppcheck pkg-config ;;\n" +
		"esac",
	Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers-extra/features/apt-get-packages:1", Options: map[string]string{"packages": "cmake,ninja-build,g++,cppcheck"}},
}

// vcpkg clones and bootstraps vcpkg, and sets VCPKG_ROOT for the steps
// after it (GitHub's and Azure's ways).
var vcpkg = protocol.Toolchain{
	Name: "vcpkg", Version: "latest",
	Install: "git clone --depth 1 https://github.com/microsoft/vcpkg.git \"$HOME/vcpkg\"\n" +
		"\"$HOME/vcpkg/bootstrap-vcpkg.sh\" -disableMetrics\n" +
		"export VCPKG_ROOT=\"$HOME/vcpkg\"\n" +
		"if [ -n \"$GITHUB_ENV\" ]; then echo \"VCPKG_ROOT=$HOME/vcpkg\" >> \"$GITHUB_ENV\"; fi\n" +
		"echo \"##vso[task.setvariable variable=VCPKG_ROOT]$HOME/vcpkg\"",
}

// usesVcpkg reports whether a project has a vcpkg.json.
func usesVcpkg(workspaceRoot string, projects []protocol.Project) bool {
	for _, p := range projects {
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "vcpkg.json")); err == nil {
			return true
		}
	}
	return false
}

// CI returns the C++ contribution: the toolchain, and vcpkg when a project
// uses it. A C++ release needs no secret: it is the tag.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	result := protocol.CIResult{Toolchains: []protocol.Toolchain{toolchain}}
	if usesVcpkg(workspaceRoot, params.Projects) {
		result.Toolchains = append(result.Toolchains, vcpkg)
	}
	return result
}

// probe runs a tool for its output; tests replace it.
var probe = func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) }

// compilers are the C++ compilers CMake finds on its own.
var compilers = []string{"c++", "g++", "clang++", "cl"}

// Doctor checks for CMake, Ninja, a C++ compiler and cppcheck, and for
// VCPKG_ROOT when a project has a vcpkg.json.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	has := func(argv ...string) bool { _, err := probe(workspaceRoot, argv); return err == nil }
	compiler := false
	for _, c := range compilers {
		if toolrun.Available(c) {
			compiler = true
			break
		}
	}
	findings := []protocol.Finding{
		{Check: "CMake is installed", OK: has("cmake", "--version"), Detail: "cmake is not on PATH", Remedy: "install CMake 3.25+ (https://cmake.org/download)"},
		{Check: "Ninja is installed", OK: has("ninja", "--version"), Detail: "ninja is not on PATH", Remedy: "install Ninja (https://ninja-build.org)"},
		{Check: "a C++ compiler is installed", OK: compiler, Detail: "none of c++, g++, clang++ or cl is on PATH", Remedy: "install GCC, Clang or MSVC (from a developer prompt)"},
		{Check: "cppcheck is installed", OK: has("cppcheck", "--version"), Detail: "cppcheck is not on PATH", Remedy: "install cppcheck (https://cppcheck.sourceforge.io)"},
	}
	for _, p := range params.Projects {
		findings = append(findings, releasing.CheckReleasePlatforms(workspaceRoot, p)...)
	}
	if usesVcpkg(workspaceRoot, params.Projects) {
		findings = append(findings, protocol.Finding{Check: "vcpkg is set up", OK: os.Getenv("VCPKG_ROOT") != "", Detail: "a project has a vcpkg.json, but VCPKG_ROOT is not set", Remedy: "clone and bootstrap vcpkg, then set VCPKG_ROOT (https://learn.microsoft.com/vcpkg/get_started)"})
	}
	return protocol.DoctorResult{Findings: findings}
}
