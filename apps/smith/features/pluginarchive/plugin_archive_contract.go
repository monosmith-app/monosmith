package pluginarchive

import "io"

// File names and the suffix that make up the format.
const (
	// ManifestFile is the plugin's manifest, which must be at the archive root.
	ManifestFile = "plugin.json"
	// ChecksumsFile lists the SHA-256 of every other file in the archive.
	ChecksumsFile = "checksums.sha256"
	// Suffix ends the name of a packaged plugin.
	Suffix = ".smith-plugin.zip"
)

// Entry is one file of a plugin, read lazily.
type Entry struct {
	// Name is the slash-separated path inside the archive.
	Name string
	// Executable is whether the file is installed with its execute bits.
	Executable bool
	// Size is the size the source declares, in bytes.
	Size int64
	// Open returns the file's content.
	Open func() (io.ReadCloser, error)
}

// Limits bound what an extraction may write, so a hostile archive cannot fill
// the disk. Zero fields mean DefaultLimits.
type Limits struct {
	MaxFiles      int
	MaxFileBytes  int64
	MaxTotalBytes int64
}

// DefaultLimits are generous for a plugin (a few binaries) and small for a bomb.
var DefaultLimits = Limits{MaxFiles: 5000, MaxFileBytes: 512 << 20, MaxTotalBytes: 1 << 30}

// Extraction is what Extract wrote.
type Extraction struct {
	Files int
	Bytes int64
	// Checksummed is whether a checksums file was found and every file matched it.
	Checksummed bool
}

func (l Limits) orDefault() Limits {
	if l.MaxFiles <= 0 {
		l.MaxFiles = DefaultLimits.MaxFiles
	}
	if l.MaxFileBytes <= 0 {
		l.MaxFileBytes = DefaultLimits.MaxFileBytes
	}
	if l.MaxTotalBytes <= 0 {
		l.MaxTotalBytes = DefaultLimits.MaxTotalBytes
	}
	return l
}
