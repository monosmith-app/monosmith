package dialect

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Request is what Resolve needs to find the dialect of a new project.
type Request struct {
	// Siblings are the shapes of the projects of the new project's kind and
	// language that report one.
	Siblings []map[string]string
	// Kind is the kind of the new project ("lib"); it words the reasons.
	Kind string
	// Follows are the shape keys a plugin of the new project's stack applies.
	Follows map[string]bool
	// Interactive says a person at a terminal can answer a question. When it
	// is false nothing is ever asked and nothing blocks.
	Interactive bool
	// In is where answers are read from when Interactive.
	In io.Reader
	// Out is where the report, the questions and the warnings go.
	Out io.Writer
}

// Dialect is the shape a new project follows and what to compare it with.
type Dialect struct {
	// Shape is sent to the plugins (GenerateParams.Shape): the keys to follow
	// with their values. Nil when there is nothing to follow.
	Shape map[string]string
	kind  string
	// expected are the values the siblings (or the person's answer) call for,
	// by key, to check the new project against.
	expected map[string]expectation
}

type expectation struct {
	value    string
	n, total int
}

// Resolve reads what the siblings agree on, prints it with the reason, and
// returns the shape to follow. A key the siblings split on, and a plugin of
// the stack follows, is asked when a person can answer (the built-in default
// is one of the answers); otherwise a warning says the default is used. With
// fewer than MinAgreeing siblings it prints nothing and follows nothing.
func Resolve(r Request) Dialect {
	d := Dialect{kind: r.Kind, expected: map[string]expectation{}}
	facts := CountFacts(r.Siblings)
	var agreed, splits []Fact
	for _, f := range facts {
		if _, ok := f.Dominant(); ok {
			agreed = append(agreed, f)
		} else if f.Split() {
			splits = append(splits, f)
		}
	}
	if len(agreed) == 0 && len(splits) == 0 {
		return d
	}
	d.Shape = map[string]string{}
	if len(agreed) > 0 {
		fmt.Fprintf(r.Out, "The %d sibling %ss agree on:\n", len(r.Siblings), r.Kind)
		for _, f := range agreed {
			top, _ := f.Dominant()
			fmt.Fprintf(r.Out, "  %s = %s (%d of %d %ss)\n", f.Key, top.Value, top.N, f.Total, r.Kind)
			d.expected[f.Key] = expectation{top.Value, top.N, f.Total}
			if r.Follows[f.Key] {
				d.Shape[f.Key] = top.Value
			}
		}
	}
	reader := bufio.NewReader(readerOrEmpty(r.In))
	for _, f := range splits {
		if !r.Follows[f.Key] {
			fmt.Fprintf(r.Out, "The sibling %ss differ on %s: %s\n", r.Kind, f.Key, describeCounts(f))
			continue
		}
		pick, asked := ask(r, reader, f)
		if !asked {
			fmt.Fprintf(r.Out, "Warning: the sibling %ss split on %s (%s); using the default. Run smith add in a terminal to choose.\n", r.Kind, f.Key, describeCounts(f))
			continue
		}
		if pick >= 0 {
			c := f.Counts[pick]
			d.Shape[f.Key] = c.Value
			d.expected[f.Key] = expectation{c.Value, c.N, f.Total}
		}
	}
	if len(d.Shape) == 0 {
		d.Shape = nil
	}
	return d
}

// Warnings compare the shape the new project has with what the siblings call
// for, and say, for each key that differs, why the project could not follow.
func (d Dialect) Warnings(shape map[string]string) []string {
	keys := make([]string, 0, len(d.expected))
	for k := range d.expected {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []string
	for _, k := range keys {
		e := d.expected[k]
		if got := shape[k]; got != e.value {
			out = append(out, fmt.Sprintf("Warning: %d of %d sibling %ss have %s = %s; the new %s has %s = %s. Its stack does not follow that key.", e.n, e.total, d.kind, k, e.value, d.kind, k, valueOrNone(got)))
		}
	}
	return out
}

func valueOrNone(v string) string {
	if v == "" {
		return "(nothing)"
	}
	return v
}

func describeCounts(f Fact) string {
	parts := make([]string, len(f.Counts))
	for i, c := range f.Counts {
		parts[i] = fmt.Sprintf("%s %d", c.Value, c.N)
	}
	return strings.Join(parts, ", ")
}

func readerOrEmpty(in io.Reader) io.Reader {
	if in == nil {
		return strings.NewReader("")
	}
	return in
}

// ask puts a split to the person. pick is the index of the chosen value, or
// -1 for the built-in default; asked is false when nobody can answer. An
// empty line, an answer that is not a number and the end of input all choose
// the default, so a question never blocks a run.
func ask(r Request, reader *bufio.Reader, f Fact) (pick int, asked bool) {
	if !r.Interactive || r.In == nil {
		return -1, false
	}
	fmt.Fprintf(r.Out, "The %d sibling %ss split on %s:\n", f.Total, r.Kind, f.Key)
	for i, c := range f.Counts {
		fmt.Fprintf(r.Out, "  %d) %s (%d %ss)\n", i+1, c.Value, c.N, r.Kind)
	}
	def := len(f.Counts) + 1
	fmt.Fprintf(r.Out, "  %d) the default of the stack\n", def)
	fmt.Fprintf(r.Out, "Which should the new %s follow [%d]: ", r.Kind, def)
	line, _ := reader.ReadString('\n')
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > len(f.Counts) {
		return -1, true
	}
	return n - 1, true
}
