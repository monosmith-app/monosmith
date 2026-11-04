package flagvocabulary

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed flag_vocabulary.json
var source []byte

// JSON returns the text of flag_vocabulary.json, the one declaration, for the
// generator that writes the clients' constants.
func JSON() []byte { return source }

var (
	once sync.Once
	doc  Document
)

// load parses the embedded file once. A broken file is a mistake in the
// repository, so it panics and any test that touches the vocabulary shows it.
func load() Document {
	once.Do(func() {
		if err := json.Unmarshal(source, &doc); err != nil {
			panic(fmt.Sprintf("flagvocabulary: flag_vocabulary.json is not valid: %v", err))
		}
	})
	return doc
}

// All returns every concept, in the order the file lists them.
func All() []Concept { return load().Concepts }

// Exceptions returns the flags excused from the rules, each with its reason.
func Exceptions() []Exception { return load().Exceptions }

// Find returns the concept with this id.
//
// It panics when the id is not in the file: a command or client that names a
// concept the vocabulary lacks is a mistake found by any test that runs it.
func Find(id string) Concept {
	for _, c := range load().Concepts {
		if c.ID == id {
			return c
		}
	}
	panic(fmt.Sprintf("flagvocabulary: no concept %q: add it to flag_vocabulary.json first", id))
}
