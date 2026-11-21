package outsidefiles

import "path"

// The kinds of a file outside the sub-folder.
const (
	// KindFile is a plain file a project file names by a path (..\LICENSE):
	// it can come along into the project folder.
	KindFile = "file"
	// KindProject is another project the build references
	// (<ProjectReference Include="..\Other\Other.csproj">): it cannot come
	// along as one file.
	KindProject = "project"
	// KindFolder is a folder a project file names by a path: it does not
	// come along as one file.
	KindFolder = "folder"
	// KindAbove is a file MSBuild reads from a folder above the project
	// (Directory.Build.props): no path names it, so it cannot come along by
	// a change of path.
	KindAbove = "above"
)

// File is one file the project uses outside its sub-folder.
type File struct {
	// Ref is the path as the manifest writes it (..\LICENSE); "" for KindAbove.
	Ref string
	// Path is the file's path from the top of the source, with forward slashes.
	Path string
	// Manifest is the file of the project that names it, from the project
	// folder (Roy-T.AStar.csproj); "" for KindAbove.
	Manifest string
	// Kind is KindFile, KindProject, KindFolder or KindAbove.
	Kind string
	// Missing is true when the source does not hold the file either.
	Missing bool
	// Content is the file's content, read before the import; nil when Missing
	// or not KindFile.
	Content []byte
}

// Movable reports whether the file can come along into the project folder:
// a plain file that the source holds.
//
// Returns: true for a KindFile that is not Missing.
func (f File) Movable() bool { return f.Kind == KindFile && !f.Missing }

// Name is the file's name in the project folder once it came along: its base
// name.
//
// Returns: the base name of Path.
func (f File) Name() string { return path.Base(f.Path) }
