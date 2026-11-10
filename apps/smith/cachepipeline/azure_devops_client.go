package cachepipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Runner starts a program with text on its standard input and returns what it
// wrote. It has the signature of cachesetup.Runner, so one fake serves both.
type Runner func(stdin string, argv ...string) (stdout, stderr string, err error)

// ErrExists is returned when a variable exists and the caller did not allow a
// replace.
var ErrExists = errors.New("the variable already exists")

// Pipeline is one pipeline of a project.
type Pipeline struct {
	// ID is the pipeline's number.
	ID int
	// Name is what the person reads.
	Name string
	// Path is the folder of the pipeline, "\" for the top.
	Path string
}

// Target is the pipeline that gets the variables.
type Target struct {
	// Organization and Project name where the pipeline lives.
	Organization, Project string
	// PipelineID is the pipeline's number; PipelineName is shown to the person.
	PipelineID   int
	PipelineName string
}

// OrganizationURL is the address az takes for --organization.
func (t Target) OrganizationURL() string {
	return AzureRepo{Organization: t.Organization}.OrganizationURL()
}

// Client reads and changes pipelines through the az program.
type Client struct {
	// Run starts az.
	Run Runner
	// TempDir makes a fresh private directory for a request body; nil uses
	// os.MkdirTemp, which makes it readable by its owner only.
	TempDir func() (string, error)
}

const apiVersion = "7.1"

// PipelinesArgv is the read-only command that lists the pipelines of a repository.
func PipelinesArgv(r AzureRepo) []string {
	return []string{"az", "pipelines", "list", "--organization", r.OrganizationURL(), "--project", r.Project,
		"--repository", r.Repository, "--repository-type", "tfsgit", "--output", "json"}
}

// ParsePipelines reads the output of PipelinesArgv.
//
// Errors: the output is not the JSON list az prints.
func ParsePipelines(out string) ([]Pipeline, error) {
	var rows []struct {
		ID   int
		Name string
		Path string
	}
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, fmt.Errorf("could not read the pipeline list: %w", err)
	}
	pipelines := make([]Pipeline, 0, len(rows))
	for _, r := range rows {
		pipelines = append(pipelines, Pipeline{ID: r.ID, Name: r.Name, Path: r.Path})
	}
	return pipelines, nil
}

// definitionArgv is the az call that reads a pipeline's definition (the
// Build Definitions REST API). The call has no secret: a secret variable comes
// back with no value.
func definitionArgv(t Target) []string {
	return append([]string{"az", "devops", "invoke", "--area", "build", "--resource", "definitions"}, invokeTail(t, "get")...)
}

// putArgv is the az call that writes a definition from the request body in file.
func putArgv(t Target, file string) []string {
	return append(append([]string{"az", "devops", "invoke", "--area", "build", "--resource", "definitions"}, invokeTail(t, "put")...), "--in-file", file)
}

func invokeTail(t Target, method string) []string {
	return []string{"--route-parameters", "project=" + t.Project, "definitionId=" + strconv.Itoa(t.PipelineID),
		"--organization", t.OrganizationURL(), "--api-version", apiVersion, "--http-method", method, "--output", "json"}
}

// Variables reads the names of a pipeline's variables.
//
// Returns: each name, with true when the variable is a secret.
// Errors: az failed (the error text is az's) or its output is not a definition.
func (c Client) Variables(t Target) (map[string]bool, error) {
	def, err := c.definition(t)
	if err != nil {
		return nil, err
	}
	vars, err := variablesOf(def)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for name, raw := range vars {
		var v struct {
			IsSecret bool `json:"isSecret"`
		}
		_ = json.Unmarshal(raw, &v)
		names[name] = v.IsSecret
	}
	return names, nil
}

// SetSecret adds a secret variable to a pipeline, or replaces one. The rest of
// the definition stays as Azure returned it. The value goes to az only in a
// request body in a private temporary file; the file is removed when the call
// ends, also on an error or a panic. When Azure says the definition changed
// since it was read, the call reads it again and tries once more.
//
// Parameters: t is the pipeline; name and value are the variable; replace
// says that an existing variable may be replaced.
// Returns: true when the variable was created, false when it was replaced.
// Errors: ErrExists when the variable exists and replace is false; otherwise
// az's error, with the value removed from its text.
func (c Client) SetSecret(t Target, name, value string, replace bool) (created bool, err error) {
	for attempt := 0; ; attempt++ {
		def, err := c.definition(t)
		if err != nil {
			return false, redact(err, value)
		}
		body, existed, err := withSecret(def, name, value)
		if err != nil {
			return false, err
		}
		if existed && !replace {
			return false, ErrExists
		}
		stderr, err := c.put(t, body, value)
		if err == nil {
			return !existed, nil
		}
		if attempt == 1 || !strings.Contains(strings.ToLower(stderr), "revision") && !strings.Contains(stderr, "409") && !strings.Contains(strings.ToLower(stderr), "conflict") {
			return false, err
		}
	}
}

// definition reads the pipeline's definition as az prints it.
func (c Client) definition(t Target) ([]byte, error) {
	out, stderr, err := c.Run("", definitionArgv(t)...)
	if err != nil {
		return nil, fmt.Errorf("az could not read the pipeline: %s", strings.TrimSpace(stderr))
	}
	return []byte(out), nil
}

// put writes body to a private temporary file and hands the file to az. It
// returns az's error text with the value removed.
func (c Client) put(t Target, body []byte, value string) (stderr string, err error) {
	dir, err := c.tempDir()
	if err != nil {
		return "", fmt.Errorf("could not make a private temporary folder: %w", err)
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "definition.json")
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("could not make the request file: %w", err)
	}
	_, werr := f.Write(body)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return "", fmt.Errorf("could not write the request file: %w", werr)
	}
	_, stderr, runErr := c.Run("", putArgv(t, file)...)
	if runErr != nil {
		stderr = strings.ReplaceAll(stderr, value, "***")
		return stderr, fmt.Errorf("az could not write the pipeline: %s", strings.TrimSpace(stderr))
	}
	return "", nil
}

func (c Client) tempDir() (string, error) {
	if c.TempDir != nil {
		return c.TempDir()
	}
	return os.MkdirTemp("", "smith-az-")
}

// redact removes a secret value from an error's text.
func redact(err error, value string) error {
	if value == "" || !strings.Contains(err.Error(), value) {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), value, "***"))
}

// variablesOf is the "variables" object of a definition.
func variablesOf(def []byte) (map[string]json.RawMessage, error) {
	var doc struct {
		Variables map[string]json.RawMessage `json:"variables"`
	}
	if err := json.Unmarshal(def, &doc); err != nil {
		return nil, fmt.Errorf("could not read the pipeline definition: %w", err)
	}
	return doc.Variables, nil
}

// withSecret returns the definition with one secret variable added or
// replaced. Every other field keeps its value; a replaced variable keeps its
// other settings, such as allowOverride.
func withSecret(def []byte, name, value string) (body []byte, existed bool, err error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(def, &doc); err != nil {
		return nil, false, fmt.Errorf("could not read the pipeline definition: %w", err)
	}
	vars := map[string]json.RawMessage{}
	if raw, ok := doc["variables"]; ok && string(raw) != "null" {
		if err := json.Unmarshal(raw, &vars); err != nil {
			return nil, false, fmt.Errorf("could not read the pipeline variables: %w", err)
		}
	}
	fields := map[string]json.RawMessage{}
	if raw, ok := vars[name]; ok {
		existed = true
		_ = json.Unmarshal(raw, &fields)
	}
	val, _ := json.Marshal(value)
	fields["value"], fields["isSecret"] = val, json.RawMessage("true")
	if vars[name], err = json.Marshal(fields); err != nil {
		return nil, false, err
	}
	if doc["variables"], err = json.Marshal(vars); err != nil {
		return nil, false, err
	}
	body, err = json.Marshal(doc)
	return body, existed, err
}
