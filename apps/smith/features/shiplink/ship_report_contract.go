package shiplink

// Report is what Ship did, for the task runner to print.
type Report struct {
	// Folder is the shipped folder, workspace-relative (dist/apps/api).
	Folder string
	// Shipped counts the files the shipped folder holds from this run:
	// linked, copied, or already the same file as the tool's.
	Shipped int
	// Copied counts the shipped files that are copies, not links.
	Copied int
	// InUse are the files, workspace-relative, that a running program holds:
	// the shipped folder keeps their old copy.
	InUse []string
	// Failed are the files that could not be shipped, each with the reason.
	Failed []Failure
	// NothingMatched is true when no glob matched a file: the tool wrote
	// somewhere else, or built nothing.
	NothingMatched bool
}

// Failure is one file Ship could not put into the shipped folder.
type Failure struct {
	// File is the workspace-relative path in the shipped folder.
	File string
	// Reason is the error, in the operating system's words.
	Reason string
}
