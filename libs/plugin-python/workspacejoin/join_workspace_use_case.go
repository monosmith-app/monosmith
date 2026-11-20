package workspacejoin

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// rootManifest is the workspace's pyproject.toml, which holds a uv workspace.
const rootManifest = "pyproject.toml"

// Join says, for each Python project, whether the uv workspace of the root
// pyproject.toml lists it in [tool.uv.workspace] members. A workspace without
// that table has no uv workspace: smith installs every project into the
// workspace venv, so each project is reached and Join reports none. The
// projects the members leave out share one fix: the root pyproject.toml with
// their parent folder's glob added, the change smith kind makes when a
// project moves. It writes nothing.
//
// Parameters: workspaceRoot is the workspace folder; params are the projects.
// Returns: one state for each project when there is a uv workspace, else none.
func Join(workspaceRoot string, params protocol.JoinParams) (protocol.JoinResult, error) {
	res := protocol.JoinResult{Projects: []protocol.JoinState{}}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, rootManifest))
	if os.IsNotExist(err) {
		return res, nil
	}
	if err != nil {
		return res, err
	}
	src := string(raw)
	members, ok := Members(src)
	if !ok {
		return res, nil
	}
	next := src
	var names []string
	for _, p := range params.Projects {
		if p.Root == "." {
			continue
		}
		joined := Covered(members, p.Root)
		state := protocol.JoinState{Project: p.Name, Joined: joined, Workspace: "the uv workspace members of the root pyproject.toml"}
		if !joined {
			state.Problem = "is not in the uv workspace members of the root pyproject.toml: uv sync and uv lock leave it out"
			next = EnsureMember(next, p.Root)
			names = append(names, p.Name)
		}
		res.Projects = append(res.Projects, state)
	}
	if next == src {
		return res, nil
	}
	fix := &protocol.Fix{
		Description: "add the projects' folders to the uv workspace members of the root pyproject.toml",
		Safe:        true,
		Files:       []protocol.FileEdit{{Path: rootManifest, Content: next}},
		AppliesTo:   names,
	}
	for i := range res.Projects {
		if !res.Projects[i].Joined {
			res.Projects[i].Fix = fix
		}
	}
	return res, nil
}
