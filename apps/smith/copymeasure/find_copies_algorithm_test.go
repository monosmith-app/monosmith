package copymeasure

import (
	"errors"
	"testing"
)

func TestMeasureCountsExactCopiesAtTheProjectRoot(t *testing.T) {
	files := map[string]string{
		"libs/a/tsconfig.json": "{\n  \"x\": 1\n}\n",
		"libs/b/tsconfig.json": "{\r\n  \"x\": 1\r\n}  \r\n", // line ends and trailing space do not matter
		"libs/c/tsconfig.json": "{\n  \"x\": 1\n}\n",
		"libs/d/tsconfig.json": "{\n  \"x\": 2\n}\n",
		"libs/a/package.json":  "a", "libs/b/package.json": "b", "libs/c/package.json": "c",
		"libs/a/src/tsconfig.json": "{\n  \"x\": 1\n}\n", // not at the root
	}
	var projects []Project
	for _, n := range []string{"a", "b", "c", "d"} {
		p := Project{Name: n, Root: "libs/" + n}
		for f := range files {
			if len(f) > len(p.Root) && f[:len(p.Root)+1] == p.Root+"/" {
				p.Files = append(p.Files, f)
			}
		}
		projects = append(projects, p)
	}
	read := func(f string) ([]byte, error) {
		s, ok := files[f]
		if !ok {
			return nil, errors.New("missing")
		}
		return []byte(s), nil
	}
	r := Measure(projects, read)
	if len(r.Copies) != 1 {
		t.Fatalf("got %+v", r.Copies)
	}
	c := r.Copies[0]
	if c.File != "tsconfig.json" || len(c.Projects) != 3 || c.Lines != 3 || len(c.Different) != 1 || c.Different[0] != "d" {
		t.Errorf("copy: %+v", c)
	}
	if r.LinesNow != 9 || r.LinesOnce != 3 {
		t.Errorf("lines now %d once %d", r.LinesNow, r.LinesOnce)
	}
}

func TestTwoCopiesAreNotEnough(t *testing.T) {
	read := func(string) ([]byte, error) { return []byte("x\n"), nil }
	r := Measure([]Project{{Name: "a", Root: "a", Files: []string{"a/f"}}, {Name: "b", Root: "b", Files: []string{"b/f"}}}, read)
	if len(r.Copies) != 0 {
		t.Errorf("%+v", r.Copies)
	}
}
