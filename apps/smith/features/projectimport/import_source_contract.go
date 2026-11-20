package projectimport

// The history modes of --history.
const (
	// HistoryKeep merges every commit of the source, written again under the
	// project folder, so git log and git blame follow them.
	HistoryKeep = "keep"
	// HistorySquash is git subtree add --squash: one commit with the source
	// and its commit in the message.
	HistorySquash = "squash"
	// HistoryCopy copies the files only. It is the only mode for a folder
	// that is not a git repository.
	HistoryCopy = "copy"
)

// Source is what smith import reads the project from.
type Source struct {
	// Arg is the source as the person typed it.
	Arg string
	// URL is the git URL, or "" for a local folder.
	URL string
	// Dir is the folder smith reads: the local folder, or the temporary clone.
	Dir string
	// Git is true when Dir is the top of a git repository with a commit, so
	// the import can take its history.
	Git bool
	// Ref is the branch, tag or commit that is imported. It is "" when Git is false.
	Ref string
	// Commit is the short hash Ref points at.
	Commit string
	// Commits is the number of commits Ref holds.
	Commits int
	// Dirty is true when a local git folder has changes that are not
	// committed: the import of its history does not take them.
	Dirty bool
	// Temp is the temporary clone that Release removes, or "".
	Temp string
	// Path is the sub-folder of the source the project is in (--path), with
	// forward slashes, or "" for the top. Dir is then that folder's content.
	Path string
	// PathRef is the ref the history of Path was split from; Ref is then
	// the temporary branch that holds it.
	PathRef string
}

// Plan is what smith import will do, settled before anything is written.
type Plan struct {
	// Name is the project's name.
	Name string
	// Kind is app, lib or package.
	Kind string
	// Folder is the project's folder, from the workspace root, with forward slashes.
	Folder string
	// Ecosystem is the language the plugins identified ("node", "python", ...).
	Ecosystem string
	// Manifest is the file that identified the project.
	Manifest string
	// Source is where the project comes from.
	Source Source
	// History is HistoryKeep, HistorySquash or HistoryCopy.
	History string
	// WriteManifest is the language whose manifest the import writes into
	// the folder after the place step (--manifest), or "" when the source has one.
	WriteManifest string
	// Sources is the number of source files of a source without a manifest.
	Sources int
	// Lockfiles are the project's own lockfiles that the import drops,
	// because the workspace has one lockfile for every project.
	Lockfiles []string
	// Tooling is the answer of the tooling question for a Node project
	// (projecttooling.Keep or projecttooling.Smith); "" for a project of
	// another language, which keeps its tools.
	Tooling string
	// Legacy names the project's legacy tools in words, the reason smith
	// recommends its own tooling; empty when its tools are current.
	Legacy []string
}
