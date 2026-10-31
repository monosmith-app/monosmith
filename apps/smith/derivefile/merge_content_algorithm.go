package derivefile

import (
	"errors"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

// ErrNoMarkers means a region file lacks the smith:begin and smith:end lines
// of its renderer, so smith cannot tell which lines it owns.
var ErrNoMarkers = errors.New("the file has no smith:begin and smith:end markers for this renderer")

// Merge returns the new content of the whole file with the renderer's content
// put in, and everything the renderer does not own kept as it is. exists tells
// whether the file is on disk; for a missing file, a whole-file or keys render
// creates it and a region render fails with ErrNoMarkers. It returns an error
// when the file does not parse in its format.
func Merge(f protocol.DerivedFile, disk []byte, exists bool) ([]byte, error) {
	wanted, err := Wanted(f)
	if err != nil {
		return nil, err
	}
	switch f.Mode {
	case protocol.DeriveFile:
		return []byte(lineendings.Match(string(disk), string(wanted))), nil
	case protocol.DeriveRegion:
		if !exists {
			return nil, ErrNoMarkers
		}
		before, _, after, ok := splitRegion(norm(string(disk)), f)
		if !ok {
			return nil, ErrNoMarkers
		}
		return []byte(lineendings.Match(string(disk), before+string(wanted)+after)), nil
	}
	if f.Format == "toml" {
		src := string(disk)
		for _, k := range sortedKeys(f.Keys) {
			raw, err := tomlValue(f.Keys[k])
			if err != nil {
				return nil, fmt.Errorf("%s: key %s: %w", f.Path, k, err)
			}
			src = tomltext.SetRaw(src, f.Table, k, raw)
		}
		return []byte(src), nil
	}
	doc, err := jsondoc.ParseDocument([]byte("{}"))
	if err != nil {
		return nil, err
	}
	if exists {
		if doc, err = jsondoc.ParseDocument(disk); err != nil {
			return nil, fmt.Errorf("%s is not a JSON object smith can edit (comments are not supported): %w", f.Path, err)
		}
	}
	for _, k := range sortedKeys(f.Keys) {
		if err := doc.SetJSON(k, f.Keys[k]); err != nil {
			return nil, err
		}
	}
	return doc.Bytes(), nil
}

// splitRegion cuts text (LF line ends) at the markers of the renderer: before
// holds the lines up to and including the begin marker, body the lines
// between the markers, after the lines from the end marker on.
func splitRegion(text string, f protocol.DerivedFile) (before, body, after string, ok bool) {
	lines := strings.SplitAfter(text, "\n")
	begin, end := -1, -1
	for i, line := range lines {
		switch {
		case begin < 0 && strings.Contains(line, "smith:begin "+f.Renderer):
			begin = i
		case begin >= 0 && strings.Contains(line, "smith:end "+f.Renderer):
			end = i
		}
		if end >= 0 {
			break
		}
	}
	if begin < 0 || end < 0 {
		return "", "", "", false
	}
	return strings.Join(lines[:begin+1], ""), strings.Join(lines[begin+1:end], ""), strings.Join(lines[end:], ""), true
}
