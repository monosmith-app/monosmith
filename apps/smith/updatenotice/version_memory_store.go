package updatenotice

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// memoryFile is the file in the smith home that remembers versions.
const memoryFile = "version.json"

// memory is what smith keeps between runs.
type memory struct {
	// LastVersion is the smith that ran last.
	LastVersion string `json:"lastVersion,omitempty"`
	// CheckedAt is when the registry was last asked.
	CheckedAt time.Time `json:"checkedAt,omitzero"`
	// Latest is the newest version the registry reported.
	Latest string `json:"latest,omitempty"`
	// NoticeAt is when a newer version was last pointed out.
	NoticeAt time.Time `json:"noticeAt,omitzero"`
}

// load reads the memory; a missing or broken file is an empty memory.
func load(home string) memory {
	var m memory
	if home == "" {
		return m
	}
	if raw, err := os.ReadFile(filepath.Join(home, memoryFile)); err == nil {
		_ = json.Unmarshal(raw, &m)
	}
	return m
}

// update reads the memory, lets change edit it and writes it back. Failing to
// write is silent: a notice must never break a command.
func update(home string, change func(*memory)) {
	if home == "" {
		return
	}
	m := load(home)
	change(&m)
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil || os.MkdirAll(home, 0o755) != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(home, memoryFile), append(raw, '\n'), 0o644)
}
