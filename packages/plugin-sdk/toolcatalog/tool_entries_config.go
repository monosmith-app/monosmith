package toolcatalog

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// checked is the date the install entries below were checked against the
// package managers' own catalogues: winget with `winget search --exact`, brew
// against formulae.brew.sh, apt against packages.debian.org (Ubuntu uses the
// same names), pacman against archlinux.org. The dnf names are the binary
// package names of Fedora and were not looked up on a page.
const checked = "2026-11-02"

// install builds one install entry.
func install(manager, pkg string) protocol.ToolInstall {
	return protocol.ToolInstall{Manager: manager, Package: pkg}
}

// entries are the tools, by name. Every entry has a Source: the page its
// install entries were checked against.
var entries = []protocol.ToolNeed{
	{
		Name: "git", Title: "Git", Optional: false,
		Why:         "smith reads the history and the changed files of the workspace through git",
		Consequence: "smith cannot tell what changed, so affected, release and undo do not work",
		Program:     "git",
		Install: []protocol.ToolInstall{
			install("winget", "Git.Git"), install("brew", "git"),
			install("apt", "git"), install("dnf", "git"), install("pacman", "git"),
		},
		Download: "https://git-scm.com/downloads", Source: "https://git-scm.com/downloads", Checked: checked,
	},
	{
		Name: "gh", Title: "the GitHub CLI", Optional: true,
		Why:         "smith release uploads release files to GitHub through the GitHub CLI",
		Consequence: "smith release cannot attach files to a GitHub release",
		Program:     "gh",
		Install: []protocol.ToolInstall{
			install("winget", "GitHub.cli"), install("brew", "gh"),
			install("apt", "gh"), install("dnf", "gh"), install("pacman", "github-cli"),
		},
		Download: "https://cli.github.com", Source: "https://github.com/cli/cli#installation", Checked: checked,
	},
	{
		Name: "node", Title: "Node.js",
		Why:         "the node projects build, test and install their packages with Node.js and npm",
		Consequence: "smith install, build and test fail for the node projects",
		Program:     "node",
		Pins: []protocol.ToolVersionPin{
			{File: ".nvmrc", Format: "text"}, {File: ".node-version", Format: "text"},
		},
		Suggested: "22",
		Install: []protocol.ToolInstall{
			install("fnm", "{major}"),
			install("winget", "OpenJS.NodeJS.LTS"), install("brew", "node"),
			install("apt", "nodejs npm"), install("dnf", "nodejs npm"), install("pacman", "nodejs npm"),
		},
		Download: "https://nodejs.org/en/download", Source: "https://nodejs.org/en/download", Checked: checked,
	},
	{
		Name: "go", Title: "Go",
		Why:         "the go projects build and test with the Go toolchain",
		Consequence: "smith build, test and lint fail for the go projects",
		Program:     "go", VersionArgs: []string{"version"},
		Pins:      []protocol.ToolVersionPin{{File: "go.mod", Format: "go-mod"}},
		Suggested: "1.25",
		Install: []protocol.ToolInstall{
			install("winget", "GoLang.Go"), install("brew", "go"),
			install("apt", "golang-go"), install("dnf", "golang"), install("pacman", "go"),
		},
		Download: "https://go.dev/dl/", Source: "https://go.dev/doc/install", Checked: checked,
	},
	{
		Name: "python", Title: "Python",
		Why:         "the python projects install, build and test with Python and pip",
		Consequence: "smith install, build and test fail for the python projects",
		Program:     "python3", Alternatives: []string{"python", "py"},
		Pins: []protocol.ToolVersionPin{
			{File: ".python-version", Format: "text"}, {File: "pyproject.toml", Format: "pyproject"},
		},
		Suggested: "3.13",
		Install: []protocol.ToolInstall{
			install("uv", "{major.minor}"),
			install("winget", "Python.Python.{major.minor}"), install("brew", "python@{major.minor}"),
			install("apt", "python3 python3-pip python3-venv"), install("dnf", "python3 python3-pip"), install("pacman", "python python-pip"),
		},
		Download: "https://www.python.org/downloads/", Source: "https://www.python.org/downloads/", Checked: checked,
	},
	{
		Name: "dotnet", Title: "the .NET SDK",
		Why:         "the .NET projects restore, build and test with the .NET SDK",
		Consequence: "smith install, build and test fail for the .NET projects",
		Program:     "dotnet",
		Pins:        []protocol.ToolVersionPin{{File: "global.json", Format: "global-json"}},
		Suggested:   "10",
		Install: []protocol.ToolInstall{
			install("winget", "Microsoft.DotNet.SDK.{major}"), install("brew-cask", "dotnet-sdk"),
			install("apt", "dotnet-sdk-{major}.0"), install("dnf", "dotnet-sdk-{major}.0"), install("pacman", "dotnet-sdk"),
		},
		Download: "https://dotnet.microsoft.com/download", Source: "https://learn.microsoft.com/dotnet/core/install/", Checked: checked,
	},
	{
		Name: "rust", Title: "Rust (cargo)",
		Why:         "the rust projects build and test with cargo",
		Consequence: "smith build, test and lint fail for the rust projects",
		Program:     "cargo",
		Pins:        []protocol.ToolVersionPin{{File: "rust-toolchain.toml", Format: "rust-toolchain"}, {File: "rust-toolchain", Format: "rust-toolchain"}},
		Install: []protocol.ToolInstall{
			{Manager: "rustup", Package: "stable", Note: "rustup is already on this machine"},
			{Manager: "brew", Package: "rust", Note: "Homebrew's own build; rustup (the official page) keeps several versions"},
			{Manager: "apt", Package: "cargo", Note: "the distribution's own build; rustup (the official page) is newer"},
			{Manager: "dnf", Package: "cargo", Note: "the distribution's own build; rustup (the official page) is newer"},
			{Manager: "pacman", Package: "rust", Note: "Arch's own build; rustup (the official page) keeps several versions"},
		},
		Download: "https://rustup.rs", Source: "https://www.rust-lang.org/tools/install", Checked: checked,
	},
	{
		Name: "java", Title: "a JDK",
		Why:         "the JVM projects compile and test with the Java Development Kit",
		Consequence: "smith build and test fail for the java, kotlin and scala projects",
		Program:     "javac", VersionArgs: []string{"-version"},
		Suggested: "21",
		Install: []protocol.ToolInstall{
			install("winget", "EclipseAdoptium.Temurin.{major}.JDK"), install("brew-cask", "temurin@{major}"),
			install("apt", "openjdk-{major}-jdk"), install("dnf", "java-{major}-openjdk-devel"), install("pacman", "jdk{major}-openjdk"),
		},
		Download: "https://adoptium.net/temurin/releases/", Source: "https://adoptium.net/installation/", Checked: checked,
	},
	{
		Name: "dart", Title: "the Dart SDK",
		Why:         "the dart projects get packages, analyze and test with the Dart SDK",
		Consequence: "smith install, build and test fail for the dart projects",
		Program:     "dart",
		Install:     []protocol.ToolInstall{install("winget", "Google.DartSDK")},
		Download:    "https://dart.dev/get-dart", Source: "https://dart.dev/get-dart", Checked: checked,
	},
	{
		Name: "ruby", Title: "Ruby",
		Why:         "the ruby projects install gems and test with Ruby and Bundler",
		Consequence: "smith install, build and test fail for the ruby projects",
		Program:     "ruby",
		Pins:        []protocol.ToolVersionPin{{File: ".ruby-version", Format: "text"}},
		Suggested:   "3.4",
		Install: []protocol.ToolInstall{
			install("winget", "RubyInstallerTeam.RubyWithDevKit.{major.minor}"), install("brew", "ruby"),
			install("apt", "ruby-full"), install("dnf", "ruby"), install("pacman", "ruby"),
		},
		Download: "https://www.ruby-lang.org/en/downloads/", Source: "https://www.ruby-lang.org/en/documentation/installation/", Checked: checked,
	},
	{
		Name: "php", Title: "PHP",
		Why:         "the php projects install packages and test with PHP",
		Consequence: "smith install, build and test fail for the php projects",
		Program:     "php",
		Suggested:   "8.4",
		Install: []protocol.ToolInstall{
			install("winget", "PHP.PHP.{major.minor}"), install("brew", "php"),
			install("apt", "php-cli"), install("dnf", "php-cli"), install("pacman", "php"),
		},
		Download: "https://www.php.net/downloads", Source: "https://www.php.net/manual/en/install.php", Checked: checked,
	},
	{
		Name: "cmake", Title: "CMake",
		Why:         "the C and C++ projects configure and build with CMake",
		Consequence: "smith build and test fail for the C and C++ projects",
		Program:     "cmake",
		Install: []protocol.ToolInstall{
			install("winget", "Kitware.CMake"), install("brew", "cmake"),
			install("apt", "cmake"), install("dnf", "cmake"), install("pacman", "cmake"),
		},
		Download: "https://cmake.org/download/", Source: "https://cmake.org/download/", Checked: checked,
	},
	{
		Name: "docker", Title: "Docker",
		Why:         "the docker projects build their images with Docker",
		Consequence: "smith build fails for the docker projects",
		Program:     "docker",
		Install: []protocol.ToolInstall{
			install("winget", "Docker.DockerDesktop"), install("brew-cask", "docker-desktop"),
			{Manager: "apt", Package: "docker.io", Note: "start the docker service and add your user to the docker group afterwards"},
			{Manager: "pacman", Package: "docker", Note: "start the docker service and add your user to the docker group afterwards"},
		},
		Download: "https://docs.docker.com/get-docker/", Source: "https://docs.docker.com/get-docker/", Checked: checked,
	},
	{
		Name: "azure-functions-core-tools", Title: "Azure Functions Core Tools", Optional: true, Choice: "azure-functions",
		Why:         "the start task runs an Azure function app on this machine with the func command",
		Consequence: "smith start cannot run the Azure function app locally (tests and build still work)",
		Program:     "func",
		Install: []protocol.ToolInstall{
			install("winget", "Microsoft.Azure.FunctionsCoreTools"), install("npm", "azure-functions-core-tools@4"),
		},
		Download: "https://learn.microsoft.com/azure/azure-functions/functions-run-local#install-the-azure-functions-core-tools",
		Source:   "https://learn.microsoft.com/azure/azure-functions/functions-run-local", Checked: checked,
	},
	{
		Name: "aws-sam-cli", Title: "the AWS SAM CLI", Optional: true, Choice: "aws-lambda",
		Why:         "the start task runs an AWS Lambda function app on this machine with the sam command",
		Consequence: "smith start cannot run the AWS function app locally (tests and build still work)",
		Program:     "sam",
		Install: []protocol.ToolInstall{
			install("winget", "Amazon.SAM-CLI"), install("brew", "aws-sam-cli"),
		},
		Download: "https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html",
		Source:   "https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html", Checked: checked,
	},
}
