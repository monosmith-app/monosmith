package languageserver

// documentStore holds the text of the documents the editor has open, by URI.
// The server handles one message at a time, so it needs no lock.
type documentStore struct {
	texts map[string]string
}

func newDocumentStore() *documentStore {
	return &documentStore{texts: map[string]string{}}
}

func (d *documentStore) set(uri, text string) { d.texts[uri] = text }

func (d *documentStore) get(uri string) (string, bool) {
	t, ok := d.texts[uri]
	return t, ok
}

func (d *documentStore) remove(uri string) { delete(d.texts, uri) }
