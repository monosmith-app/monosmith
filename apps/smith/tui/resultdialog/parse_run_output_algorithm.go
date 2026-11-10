package resultdialog

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

var (
	filesWritten  = regexp.MustCompile(`Wrote (\d+) file`)                      // words:literal
	creatingLine  = regexp.MustCompile(`^Creating \S+ (\S+) in (\S+?):`)        // words:literal
	missingTool   = regexp.MustCompile(`(\S+) is not installed or not on PATH`) // words:literal
	installFrom   = regexp.MustCompile(`Install it from (https?://[^\s,]+)`)    // words:literal
	verifyingLine = regexp.MustCompile(`^Verifying: smith (run .+ -p \S+)$`)    // words:literal
)

// toolTitles are the names a person knows for the programs smith checks.
var toolTitles = map[string]string{
	"dotnet": "the .NET SDK", "go": "Go", "node": "Node.js", "npm": "Node.js", "npx": "Node.js", // words:literal
	"python": "Python", "python3": "Python", "pip": "Python", "cargo": "Rust", "rustc": "Rust", // words:literal
	"java": "Java (JDK)", "javac": "Java (JDK)", "gradle": "Gradle", "mvn": "Maven", "flutter": "Flutter", // words:literal
	"dart": "Dart", "git": "Git", "docker": "Docker", "cmake": "CMake", // words:literal
}

// MissingTool is what a `smith add` told when its first check could not run
// because a tool is missing.
type MissingTool struct {
	// Project is the new project; Dir is where it is, "" when unknown.
	Project, Dir string
	// Files is how many files smith wrote, 0 when unknown.
	Files int
	// Program is the program that was not found ("dotnet"); Title is its name for a person.
	Program, Title string
	// URL is the page that tells how to get it, "" when smith printed none.
	URL string
	// Check is the smith command line that runs the first check again, nil when unknown.
	Check []string
}

// ParseMissingTool reads the output of a failed `smith add` and tells whether it
// failed because a tool is missing, which leaves the project on disk.
//
// Parameters: argv is the command line that ran; lines is its output.
// Returns: what was created and what is missing; false for any other failure.
func ParseMissingTool(argv, lines []string) (MissingTool, bool) {
	if len(argv) < 3 || argv[0] != "add" || strings.HasPrefix(argv[2], "-") {
		return MissingTool{}, false
	}
	t := MissingTool{Project: argv[2]}
	found := false
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if m := filesWritten.FindStringSubmatch(l); m != nil {
			t.Files, _ = strconv.Atoi(m[1])
		}
		if m := creatingLine.FindStringSubmatch(l); m != nil {
			t.Dir = strings.TrimSuffix(m[2], "/")
		}
		if m := missingTool.FindStringSubmatch(l); m != nil && !found {
			found, t.Program = true, strings.TrimSuffix(strings.ToLower(m[1]), ".exe")
		}
		if m := installFrom.FindStringSubmatch(l); m != nil && t.URL == "" {
			t.URL = m[1]
		}
		if m := verifyingLine.FindStringSubmatch(l); m != nil {
			t.Check = strings.Fields(m[1])
		}
	}
	if !found {
		return MissingTool{}, false
	}
	if t.Title = toolTitles[t.Program]; t.Title == "" {
		t.Title = t.Program
	}
	return t, true
}

// Created is what a finished `smith add` made, read from its output.
type Created struct {
	// Name is the project and Kind its kind (app, lib, package); Dir is where it is.
	Name, Kind, Dir string
	// Files is how many files smith wrote, 0 when unknown.
	Files int
}

// ParseCreated reads the output of a `smith add` that finished.
//
// Parameters: argv is the command line that ran; lines is its output.
// Returns: the project; false when argv is not an add.
func ParseCreated(argv, lines []string) (Created, bool) {
	if len(argv) < 3 || argv[0] != "add" || strings.HasPrefix(argv[2], "-") {
		return Created{}, false
	}
	c := Created{Kind: argv[1], Name: argv[2]}
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if m := filesWritten.FindStringSubmatch(l); m != nil {
			c.Files, _ = strconv.Atoi(m[1])
		}
		if m := creatingLine.FindStringSubmatch(l); m != nil {
			c.Dir = strings.TrimSuffix(m[2], "/")
		}
	}
	return c, true
}

// OperationText is the words that name an operation in a commit: the command
// and its arguments up to the first flag ("add app web"), as smith names it.
//
// Parameters: argv is the smith command line.
// Returns: the text, "" for an empty line.
func OperationText(argv []string) string {
	var words []string
	for _, a := range argv {
		if strings.HasPrefix(a, "-") {
			break
		}
		words = append(words, a)
	}
	return strings.Join(words, " ")
}

// FileCount is "1 file" or "12 files".
//
// Parameters: n is the number of files.
// Returns: the count with the noun in the form that fits n.
func FileCount(n int) string { return words.Count("resultdialog.files", n) }
