package flagvocabulary

// Spelling is one extra way to write a flag.
type Spelling struct {
	// Name is the long form without dashes.
	Name string `json:"name"`
	// Short is the one-letter form without the dash; empty when there is none.
	Short string `json:"short,omitempty"`
}

// Concept is one idea that commands share and the flag that carries it.
type Concept struct {
	// ID names the concept: "project".
	ID string `json:"id"`
	// Flag is the canonical long name without dashes: "project".
	Flag string `json:"flag"`
	// Short is the canonical one-letter form without the dash; empty when none.
	Short string `json:"short"`
	// Type is the value kind: bool, string or stringSlice (repeatable and comma
	// separated).
	Type string `json:"type"`
	// Synonyms are other spellings every command with the flag also accepts.
	Synonyms []Spelling `json:"synonyms"`
	// Near are names a command must not define for this concept: they are the
	// near-synonyms people reach for.
	Near []string `json:"near"`
	// Meaning says in plain words what the flag is for.
	Meaning string `json:"meaning"`
	// Commands are the command paths under smith that take the flag. Empty
	// means "many": a test cannot list them and does not check.
	Commands []string `json:"commands"`
}

// Exception excuses one flag from the vocabulary rules, with the reason.
type Exception struct {
	// Command is the command path under smith ("graph").
	Command string `json:"command"`
	// Flag is the flag's long name without dashes.
	Flag string `json:"flag"`
	// Why says why the flag is not a spelling of a shared concept.
	Why string `json:"why"`
}

// Document is the content of flag_vocabulary.json.
type Document struct {
	Concepts   []Concept   `json:"concepts"`
	Exceptions []Exception `json:"exceptions"`
}
