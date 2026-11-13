package pipeline

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-jvm/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolpin"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Versions of the build tools installed when the workspace does not pin them;
// bump them deliberately.
const (
	MavenVersion  = "3.9.11"
	GradleVersion = "9.8.1"
)

// Version is the Java release the workspace uses, as toolpin decides it.
func Version(workspaceRoot string) string {
	return toolpin.Version(workspaceRoot, "java")
}

// buildTools says which build tools the projects use, and whether any is
// an Android module.
type buildTools struct{ maven, gradle, android bool }

func usedBuildTools(workspaceRoot string, projects []protocol.Project) buildTools {
	var used buildTools
	for _, p := range projects {
		dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
		if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
			used.maven = true
		}
		for _, script := range []string{"build.gradle.kts", "build.gradle"} {
			if src, err := os.ReadFile(filepath.Join(dir, script)); err == nil {
				used.gradle = true
				used.android = used.android || strings.Contains(string(src), "com.android.")
			}
		}
	}
	return used
}

func runtime(version string, used buildTools) protocol.Toolchain {
	cache := "maven"
	if !used.maven {
		cache = "gradle"
	}
	return protocol.Toolchain{
		Name: "java", Version: version,
		// GitHub's and Azure's hosted images carry Maven.
		GitHub: &protocol.ActionStep{Uses: "actions/setup-java@v5", Inputs: []protocol.Input{{Name: "distribution", Value: "temurin"}, {Name: "java-version", Value: version}, {Name: "cache", Value: cache}}},
		Azure:  &protocol.ActionStep{Uses: "JavaToolInstaller@0", Inputs: []protocol.Input{{Name: "versionSpec", Value: version}, {Name: "jdkArchitectureOption", Value: "x64"}, {Name: "jdkSourceOption", Value: "PreInstalled"}}},
		// Elsewhere (GitLab, on a Debian image): Temurin and Maven archives.
		Install: "mkdir -p /opt/jdk /opt/maven\n" +
			"curl -fsSL https://api.adoptium.net/v3/binary/latest/" + version + "/ga/linux/x64/jdk/hotspot/normal/eclipse | tar -xz -C /opt/jdk --strip-components=1\n" +
			"curl -fsSL https://archive.apache.org/dist/maven/maven-3/" + MavenVersion + "/binaries/apache-maven-" + MavenVersion + "-bin.tar.gz | tar -xz -C /opt/maven --strip-components=1\n" +
			"export JAVA_HOME=/opt/jdk PATH=/opt/jdk/bin:/opt/maven/bin:$PATH",
		Devcontainer: &protocol.DevcontainerFeature{ID: "ghcr.io/devcontainers/features/java:1", Options: map[string]string{"version": version, "installMaven": "true", "installGradle": strconv.FormatBool(used.gradle)}},
	}
}

// gradle installs the Gradle distribution at GradleVersion on every
// provider, and puts only its bin/ first on PATH: for the later steps on
// GitHub ($GITHUB_PATH) and Azure (a logging command), for the rest of the
// script elsewhere. setup-gradle is not used: when a Gradle is already on
// PATH (/usr/bin on GitHub's Ubuntu image) it puts that directory first,
// ahead of the Ruby, Python or .NET other steps installed.
func gradle() protocol.Toolchain {
	bin := "$HOME/gradle-" + GradleVersion + "/bin"
	return protocol.Toolchain{
		Name: "gradle", Version: GradleVersion,
		Install: "curl -fsSLo /tmp/gradle.zip https://services.gradle.org/distributions/gradle-" + GradleVersion + "-bin.zip\n" +
			"unzip -q -o -d \"$HOME\" /tmp/gradle.zip && rm /tmp/gradle.zip\n" +
			"export PATH=\"" + bin + ":$PATH\"\n" +
			"if [ -n \"$GITHUB_PATH\" ]; then echo \"" + bin + "\" >> \"$GITHUB_PATH\"; fi\n" +
			"echo \"##vso[task.prependpath]" + bin + "\"",
	}
}

// CI returns the JVM contribution: a JDK with Maven, and Gradle when a
// module builds with it, resolving every Maven module's dependencies, and
// the repository sign-in when a package is deployed: the job's token for
// GitHub Packages, else the Maven secrets.
func CI(workspaceRoot string, params protocol.CIParams) protocol.CIResult {
	if len(params.Projects) == 0 {
		return protocol.CIResult{}
	}
	used := usedBuildTools(workspaceRoot, params.Projects)
	result := protocol.CIResult{Toolchains: []protocol.Toolchain{runtime(Version(workspaceRoot), used)}}
	if used.gradle {
		result.Toolchains = append(result.Toolchains, gradle())
	}
	if used.android {
		result.Toolchains = append(result.Toolchains, androidSDK())
	}
	if used.maven {
		result.Setup = []protocol.Step{{Name: "Resolve Maven dependencies", Command: []string{"mvn", "-B", "-q", "dependency:go-offline"}, Blocking: true}}
	}
	signIn := registries.CI(params.Projects)
	result.Authenticate = signIn.Authenticate
	result.Credentials = signIn.Credentials
	result.Permissions = signIn.Permissions
	return result
}

// probe runs a tool for its output, stdout then stderr (java prints its
// version to stderr); tests replace it.
var probe = func(dir string, argv []string) (string, error) {
	out, errOut, err := toolrun.Capture(dir, argv)
	return out + errOut, err
}

var javaVersion = regexp.MustCompile(`version "(\d+(?:\.\d+)*)`)

// javaVersionFinding compares the JDK on PATH with the release the generated
// CI and dev container install; it says nothing when java prints no version.
func javaVersionFinding(output, ci string) (protocol.Finding, bool) {
	m := javaVersion.FindStringSubmatch(output)
	if m == nil {
		return protocol.Finding{}, false
	}
	have := m[1]
	if rest, ok := strings.CutPrefix(have, "1."); ok { // Java 8 prints 1.8.0
		have = rest
	}
	major, _, _ := strings.Cut(have, ".")
	f, ok := toolrun.CIPinFinding("JDK", major, ci)
	if ok && !f.OK {
		f.Remedy = "install JDK " + ci + " (https://adoptium.net), or set the release in .java-version and run `smith ci generate` to move CI and the dev container"
		// Installing a JDK is the person's own step (no smith command does it), so
		// the fix is the other way: follow this machine, then move CI to match.
		f.Fix = &protocol.Fix{
			Description: "pin JDK " + major + " in .java-version and move CI and the dev container to it",
			Files:       []protocol.FileEdit{{Path: ".java-version", Content: major + "\n"}},
			Steps:       []protocol.FixStep{{Description: "write CI and the dev container again so they follow the pin", Argv: []string{"smith", "ci", "generate", "--yes"}}},
		}
	}
	return f, ok
}

// Doctor checks that a JDK is installed, Maven or Gradle for the modules
// that build with them, and the Android SDK for an Android module.
func Doctor(workspaceRoot string, params protocol.DoctorParams) protocol.DoctorResult {
	if len(params.Projects) == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}
	}
	version := Version(workspaceRoot)
	javaOut, javaErr := probe(workspaceRoot, []string{"java", "-version"})
	findings := []protocol.Finding{
		{Check: "a JDK is installed", Area: "toolchain", OK: javaErr == nil, Problem: toolrun.ProblemUnless(javaErr == nil, "java is not on PATH"), Detail: "java is not on PATH", Remedy: "install JDK " + version + " (https://adoptium.net)"},
	}
	if f, ok := javaVersionFinding(javaOut, version); ok && javaErr == nil {
		findings = append(findings, f)
	}
	used := usedBuildTools(workspaceRoot, params.Projects)
	if used.maven {
		_, err := probe(workspaceRoot, []string{"mvn", "-v"})
		findings = append(findings, protocol.Finding{Check: "Maven is installed", Area: "toolchain", OK: err == nil, Problem: toolrun.ProblemUnless(err == nil, "mvn is not on PATH"), Detail: "mvn is not on PATH", Remedy: "install Maven " + MavenVersion + " (https://maven.apache.org)"})
	}
	if used.gradle {
		_, err := probe(workspaceRoot, []string{"gradle", "--version"})
		findings = append(findings, protocol.Finding{Check: "Gradle is installed", Area: "toolchain", OK: err == nil, Problem: toolrun.ProblemUnless(err == nil, "gradle is not on PATH"), Detail: "gradle is not on PATH", Remedy: "install Gradle " + GradleVersion + " (https://gradle.org/install)"})
	}
	if used.android {
		sdkFound := androidSDKFound(workspaceRoot)
		findings = append(findings, protocol.Finding{Check: "the Android SDK is installed", Area: "toolchain", OK: sdkFound, Problem: toolrun.ProblemUnless(sdkFound, "no Android SDK is found: neither ANDROID_HOME nor local.properties' sdk.dir names one"), Detail: "neither ANDROID_HOME nor local.properties' sdk.dir names an SDK", Remedy: "install the Android SDK command-line tools (https://developer.android.com/tools) and set ANDROID_HOME"})
	}
	if found, ok := CheckBuildDir(workspaceRoot, params.Projects); ok {
		findings = append(findings, found...)
	}
	return protocol.DoctorResult{Findings: findings}
}
