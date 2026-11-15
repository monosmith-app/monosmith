package plugininstall

import "context"

// Kind says where a plugin comes from.
type Kind string

// The kinds of source `smith plugin install` accepts.
const (
	// KindLocal is an archive file or a plugin directory on this machine.
	KindLocal Kind = "local"
	// KindURL is an archive at an https URL.
	KindURL Kind = "url"
	// KindGitHub is an asset of a GitHub release: github:owner/repo@tag.
	KindGitHub Kind = "github"
)

// Source is a parsed install argument.
type Source struct {
	Kind Kind
	// Location is the path (KindLocal) or the URL (KindURL).
	Location string
	// Owner, Repo, Tag and Asset describe a GitHub release (KindGitHub); an
	// empty Asset means the one asset ending in .smith-plugin.zip.
	Owner, Repo, Tag, Asset string
}

// String is how the source is shown to the person.
func (s Source) String() string {
	if s.Kind == KindGitHub {
		out := "github:" + s.Owner + "/" + s.Repo + "@" + s.Tag
		if s.Asset != "" {
			out += "#" + s.Asset
		}
		return out
	}
	return s.Location
}

// Plugin is what the caller learns of the staged files, from their manifest.
type Plugin struct {
	Name    string
	Version string
}

// Request is one install.
type Request struct {
	// Source is the argument as typed.
	Source string
	// PluginsDir is where installed plugins live; the plugin goes into a
	// directory named after it.
	PluginsDir string
	// SHA256 is the digest the person expects the archive to have; empty when
	// they gave none. A GitHub release's own .sha256 asset is used when this is empty.
	SHA256 string
	// Replace allows overwriting an installed plugin of the same name.
	Replace bool
	// Inspect reads the staged files (nothing has been run) and returns the
	// plugin they hold, or refuses them. It runs before anything is installed.
	Inspect func(stagedDir string) (Plugin, error)
}

// Result is a finished install.
type Result struct {
	Plugin
	// Dir is where the plugin now is.
	Dir string
	// Source is what was installed from.
	Source Source
	// ArchiveSHA256 is the digest of the archive that was downloaded or read;
	// empty for a directory.
	ArchiveSHA256 string
	// ArchiveVerified says the archive matched a digest the person or the
	// release published, not just its own checksums file.
	ArchiveVerified bool
	// Checksummed says every file matched the archive's checksums file.
	Checksummed bool
	Files       int
	Bytes       int64
	// Replaced is true when an installed plugin of that name was overwritten.
	Replaced bool
}

// Described is what the plugin host reads from an extracted plugin directory.
type Described struct {
	Plugin
	// Argv is the manifest's command resolved for this machine, and Command0
	// its first element as written.
	Argv     []string
	Command0 string
}

// Resolved is a marketplace name turned into something Install accepts.
type Resolved struct {
	// Name is the plugin and version as found, such as "widget@1.2.0".
	Name string
	// Origin is the index the name was found in.
	Origin string
	// Source is the archive's https URL and SHA256 the digest the index lists for it.
	Source, SHA256 string
}

// Resolver turns an install argument that names a marketplace plugin into the
// archive to fetch. It returns false, and no error, for an argument that is
// not a marketplace name (a path, a URL, github:owner/repo@tag), which Install
// handles as before.
type Resolver func(ctx context.Context, arg string) (Resolved, bool, error)
