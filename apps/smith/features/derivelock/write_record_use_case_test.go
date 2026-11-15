package derivelock

import "testing"

func TestRecordRoundTripsSortedAndFindsEntries(t *testing.T) {
	root := t.TempDir()
	r := Record{}
	r.Set(Entry{Path: "b/tsconfig.json", Owner: OwnerSmith, Renderer: "x", Version: 1, Mode: "keys", Hash: Hash([]byte("b"))})
	r.Set(Entry{Path: "a/tsconfig.json", Owner: OwnerYou, Renderer: "x", Version: 1, Mode: "file", Hash: Hash([]byte("a"))})
	r.Set(Entry{Path: "b/tsconfig.json", Owner: OwnerSmith, Renderer: "x", Version: 2, Mode: "keys", Hash: Hash([]byte("b2"))})
	if err := Write(root, r); err != nil {
		t.Fatal(err)
	}
	got, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Files) != 2 || got.Files[0].Path != "a/tsconfig.json" {
		t.Fatalf("entries not sorted or duplicated: %+v", got.Files)
	}
	if e := got.Find("b/tsconfig.json", "x"); e == nil || e.Version != 2 {
		t.Fatalf("Set did not replace the entry: %+v", e)
	}
	if owned := got.OwnedBySmith(); len(owned) != 1 || owned[0] != "b/tsconfig.json" {
		t.Fatalf("owned = %v", owned)
	}
}

func TestEmptyRecordWritesNothing(t *testing.T) {
	root := t.TempDir()
	r, err := Read(root)
	if err != nil || len(r.Files) != 0 || r.Schema != SchemaVersion {
		t.Fatalf("missing record: %+v %v", r, err)
	}
	if err := Write(root, r); err != nil {
		t.Fatal(err)
	}
	if _, statErr := Read(root); statErr != nil {
		t.Fatal(statErr)
	}
}
