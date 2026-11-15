package verifyfacts

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Coverage is a coverage report that is already in place.
type Coverage struct {
	// Path is the report, workspace-relative with forward slashes.
	Path string
	// Percent is the share of lines (or statements) covered, 0 to 100.
	Percent float64
	// Modified is when the file was written: a report older than the run is shown
	// with its age, not passed off as the result of this run.
	Modified time.Time
}

// coverageFiles are the reports smith knows, in the order it looks, each with the
// way to read the percentage from it. Paths are relative to the project folder.
var coverageFiles = []struct {
	path  string
	parse func(path string) (float64, bool)
}{
	{"coverage/coverage-summary.json", readIstanbulSummary},
	{"coverage/lcov.info", readLcov},
	{"lcov.info", readLcov},
	{"coverage.lcov", readLcov},
	{"coverage/cobertura-coverage.xml", readCobertura},
	{"coverage.xml", readCobertura},
	{"cobertura.xml", readCobertura},
	{"coverage.out", readGoProfile},
	{"cover.out", readGoProfile},
	{"coverage.txt", readGoProfile},
}

// FindCoverage looks for a coverage report in the project folder, and for a
// cobertura report a .NET test run wrote under TestResults. It reads nothing it
// was not asked to: smith never turns coverage on, so a project without a report
// has none to show.
//
// Parameters: workspaceRoot is the workspace folder; projectRoot is the project
// folder, workspace-relative.
// Returns: the report and true when one is in place and could be read.
func FindCoverage(workspaceRoot, projectRoot string) (Coverage, bool) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot))
	for _, f := range coverageFiles {
		path := filepath.Join(dir, filepath.FromSlash(f.path))
		if c, ok := coverageAt(workspaceRoot, path, f.parse); ok {
			return c, true
		}
	}
	// dotnet test --collect "XPlat Code Coverage" writes TestResults/<guid>/coverage.cobertura.xml.
	matches, _ := filepath.Glob(filepath.Join(dir, "TestResults", "*", "coverage.cobertura.xml"))
	for _, path := range matches {
		if c, ok := coverageAt(workspaceRoot, path, readCobertura); ok {
			return c, true
		}
	}
	return Coverage{}, false
}

func coverageAt(workspaceRoot, path string, parse func(string) (float64, bool)) (Coverage, bool) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return Coverage{}, false
	}
	percent, ok := parse(path)
	if !ok {
		return Coverage{}, false
	}
	rel, err := filepath.Rel(workspaceRoot, path)
	if err != nil {
		rel = path
	}
	return Coverage{Path: filepath.ToSlash(rel), Percent: percent, Modified: info.ModTime()}, true
}

// readIstanbulSummary reads total.lines.pct of the json-summary reporter of istanbul (jest, vitest, c8).
func readIstanbulSummary(path string) (float64, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	var doc struct {
		Total struct {
			Lines struct {
				Pct float64 `json:"pct"`
			} `json:"lines"`
		} `json:"total"`
	}
	if json.Unmarshal(raw, &doc) != nil {
		return 0, false
	}
	return doc.Total.Lines.Pct, true
}

// readLcov adds the LF (lines found) and LH (lines hit) of every record.
func readLcov(path string) (float64, bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, false
	}
	defer func() { _ = f.Close() }()
	var found, hit int
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 16<<20)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "LF:"):
			n, _ := strconv.Atoi(strings.TrimPrefix(line, "LF:"))
			found += n
		case strings.HasPrefix(line, "LH:"):
			n, _ := strconv.Atoi(strings.TrimPrefix(line, "LH:"))
			hit += n
		}
	}
	if found == 0 {
		return 0, false
	}
	return 100 * float64(hit) / float64(found), true
}

// readCobertura reads the line-rate of the root element (pytest-cov, coverlet, tarpaulin).
func readCobertura(path string) (float64, bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, false
	}
	defer func() { _ = f.Close() }()
	var doc struct {
		LineRate string `xml:"line-rate,attr"`
	}
	if xml.NewDecoder(f).Decode(&doc) != nil || doc.LineRate == "" {
		return 0, false
	}
	rate, err := strconv.ParseFloat(doc.LineRate, 64)
	if err != nil {
		return 0, false
	}
	return 100 * rate, true
}

// goProfileLine reads "file.go:10.2,12.3 2 1": the statements of a block and how many times it ran.
var goProfileLine = regexp.MustCompile(`^\S+:\d+\.\d+,\d+\.\d+ (\d+) (\d+)$`)

// readGoProfile adds the statements of every block of a go test -coverprofile file.
func readGoProfile(path string) (float64, bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, false
	}
	defer func() { _ = f.Close() }()
	var total, covered int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		m := goProfileLine.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		stmts, _ := strconv.Atoi(m[1])
		runs, _ := strconv.Atoi(m[2])
		total += stmts
		if runs > 0 {
			covered += stmts
		}
	}
	if total == 0 {
		return 0, false
	}
	return 100 * float64(covered) / float64(total), true
}
