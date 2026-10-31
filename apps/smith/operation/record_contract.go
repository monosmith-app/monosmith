package operation

// SchemaVersion is the .smith/operations.json schema this build reads and writes.
const SchemaVersion = 1

// Record is one operation smith committed.
type Record struct {
	// ID counts up from 1.
	ID int `json:"id"`
	// Command is the human text of the operation: "add package billing".
	Command string `json:"command"`
	// Commit is the hash of the commit that holds the operation.
	Commit string `json:"commit"`
	// Files are the paths the commit changed, workspace-relative.
	Files []string `json:"files"`
	// Time is the commit time in RFC 3339.
	Time string `json:"time"`
	// Undone is true once smith undo reverted the commit.
	Undone bool `json:"undone,omitempty"`
	// UndoneBy is the hash of the revert commit.
	UndoneBy string `json:"undoneBy,omitempty"`
}

// Log is the list of recorded operations, oldest first.
type Log struct {
	Schema     int      `json:"schema"`
	Operations []Record `json:"operations"`
}

// Last returns the newest operation that is not undone, or nil.
func (l *Log) Last() *Record {
	for i := len(l.Operations) - 1; i >= 0; i-- {
		if !l.Operations[i].Undone {
			return &l.Operations[i]
		}
	}
	return nil
}
