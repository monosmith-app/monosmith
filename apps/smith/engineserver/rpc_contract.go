package engineserver

import (
	"encoding/json"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipecatalog"
)

// ReleaseParams are the parameters of the release method.
type ReleaseParams struct {
	// NoPush creates the tags locally only.
	NoPush bool `json:"noPush,omitempty"`
}

// SettingParams are the parameters of the methods that name one setting of
// smith.toml, such as "ci.devcontainer".
type SettingParams struct {
	Setting string `json:"setting"`
}

// ConfigSetParams are the parameters of config.set: Value for a setting that
// holds one value, Values for a list.
type ConfigSetParams struct {
	Setting string   `json:"setting"`
	Value   string   `json:"value,omitempty"`
	Values  []string `json:"values,omitempty"`
}

// ConfigValue is what config.get answers. Value is as written in smith.toml;
// when Set is false the setting has its default.
type ConfigValue struct {
	Setting     string `json:"setting"`
	Value       string `json:"value,omitempty"`
	Set         bool   `json:"set"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// RecipeParams are the parameters of recipe.describe and recipe.roles.
type RecipeParams struct {
	// Recipe is a recipe file in the workspace, or an https URL.
	Recipe string `json:"recipe"`
	// Choices answer the recipe's choices for recipe.roles, by name.
	Choices map[string]string `json:"choices,omitempty"`
}

// RecipeList is what recipe.list answers.
type RecipeList struct {
	Recipes []recipecatalog.Listed `json:"recipes"`
}

// RecipeRoles is what recipe.roles answers: the choices as answered and the
// plugins each role gets.
type RecipeRoles struct {
	Name    string            `json:"name"`
	Version string            `json:"version,omitempty"`
	Choices []recipe.Question `json:"choices"`
	Roles   []recipe.Use      `json:"roles"`
}

// ProtocolVersion is the version of the methods and their parameters. Additions
// are optional fields and new methods, so a client written for a version keeps
// working with a newer smith.
const ProtocolVersion = 1

// EventNotification is the method of the notifications that carry run events.
const EventNotification = "event"

// JSON-RPC error codes.
const (
	codeParse          = -32700
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeFailed         = -32000
)

type request struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return fmt.Sprintf("%d: %s", e.Code, e.Message) }

// RunParams are the parameters of the run method.
type RunParams struct {
	// Run names the run in its events; empty means smith chooses.
	Run      string   `json:"run,omitempty"`
	Tasks    []string `json:"tasks"`
	Projects []string `json:"projects,omitempty"`
	Affected bool     `json:"affected,omitempty"`
	Base     string   `json:"base,omitempty"`
	Parallel int      `json:"parallel,omitempty"`
	Bail     bool     `json:"bail,omitempty"`
	NoCache  bool     `json:"noCache,omitempty"`
}

// CancelParams are the parameters of the cancel method.
type CancelParams struct {
	Run string `json:"run"`
}

// RunResult is what run answers once the run has finished.
type RunResult struct {
	Run     string `json:"run"`
	OK      bool   `json:"ok"`
	Summary any    `json:"summary"`
}

// CommandResult is what a method that runs a smith command answers.
type CommandResult struct {
	Run      string `json:"run"`
	OK       bool   `json:"ok"`
	ExitCode int    `json:"exitCode"`
}

// InitializeResult is what initialize answers.
type InitializeResult struct {
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	ProtocolVersion int      `json:"protocolVersion"`
	EventVersion    int      `json:"eventVersion"`
	Workspace       string   `json:"workspace"`
	Methods         []string `json:"methods"`
}
