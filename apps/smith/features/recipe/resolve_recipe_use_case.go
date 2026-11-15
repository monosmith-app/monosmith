package recipe

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// declaration is a choice as the chain so far has declared it.
type declaration struct {
	choice Choice
	origin Origin
}

// walker resolves a recipe chain for one set of answers.
type walker struct {
	res      Resolution
	answers  map[string]string
	used     map[string]bool
	declared map[string]*declaration
	resolved map[string]string
	// order lists the declared choices in the order they were declared.
	order []string
}

// Resolve turns a recipe chain and the answers to its choices into one
// Resolution:
//
//  1. The choices of every recipe in the chain are merged, key by key, the
//     child's over the parent's, and each is answered: by the answers given,
//     else by its default. A choice with neither is a MissingAnswerError that
//     names every such choice. A given answer must be one of the options.
//  2. The recipes are applied oldest first. Within one, the unconditional
//     use and settings come first, then the when blocks that match, in file
//     order; a block's own choices are asked only when it matches, and its
//     nested when blocks follow its use and settings.
//  3. Later values replace earlier ones key by key; settings tables merge
//     recursively. Every value records the recipe, file and line it came from.
//
// A when block that names a choice no recipe declares, or an answer that is
// not one of the choice's options, is an error. So is an answer given to a
// choice that is never asked. On error the Resolution holds what was asked so
// far.
func Resolve(loaded Loaded, answers map[string]string) (Resolution, error) {
	if len(loaded.Chain) == 0 {
		return Resolution{}, fmt.Errorf("no recipe to resolve")
	}
	last := loaded.Chain[len(loaded.Chain)-1].Recipe
	w := &walker{
		res:      Resolution{Name: last.Name, Version: last.Version},
		answers:  answers,
		used:     map[string]bool{},
		declared: map[string]*declaration{},
		resolved: map[string]string{},
	}
	for _, doc := range loaded.Chain {
		w.res.Chain = append(w.res.Chain, Link{Name: doc.Recipe.Name, Version: doc.Recipe.Version, Source: doc.Source.Ref})
		for _, name := range byLine(doc, "choices.", doc.Recipe.Choices) {
			w.declare(doc, name, doc.Recipe.Choices[name], "")
		}
	}
	if err := w.answerPending(); err != nil {
		return w.res, err
	}
	for _, doc := range loaded.Chain {
		if err := w.apply(doc, doc.Recipe.Block, ""); err != nil {
			return w.res, err
		}
	}
	for _, name := range sortedKeys(w.answers) {
		if !w.used[name] {
			return w.res, fmt.Errorf("--choice %s=%s: this recipe asks no choice named %q with these answers (it asks: %s)", name, w.answers[name], name, strings.Join(w.order, ", "))
		}
	}
	return w.res, nil
}

// byLine lists the names of a table of a document in file order.
func byLine[V any](doc Document, table string, m map[string]V) []string {
	names := sortedKeys(m)
	sort.SliceStable(names, func(i, j int) bool {
		return doc.Line(table+names[i]) < doc.Line(table+names[j])
	})
	return names
}

// declare records a choice, merging it over an earlier declaration of the
// same name key by key.
func (w *walker) declare(doc Document, name string, c Choice, prefix string) {
	origin := Origin{Recipe: doc.Recipe.Name, File: doc.Source.Ref, Line: doc.Line(prefix + "choices." + name)}
	d, ok := w.declared[name]
	if !ok {
		w.declared[name] = &declaration{choice: c, origin: origin}
		w.order = append(w.order, name)
		return
	}
	if c.Ask != "" {
		d.choice.Ask = c.Ask
	}
	if len(c.Options) > 0 {
		d.choice.Options = c.Options
	}
	if c.Default != "" {
		d.choice.Default = c.Default
	}
	d.origin = origin
}

// answerPending answers every declared choice not answered yet. Choices with
// no answer and no default are reported together.
func (w *walker) answerPending() error {
	var missing []Question
	for _, name := range w.order {
		if _, done := w.resolved[name]; done {
			continue
		}
		d := w.declared[name]
		q := Question{Name: name, Ask: d.choice.Ask, Options: d.choice.Options, Default: d.choice.Default, Origin: d.origin}
		given, ok := w.answers[name]
		switch {
		case ok:
			if len(q.Options) > 0 && !slices.Contains(q.Options, given) {
				return fmt.Errorf("--choice %s=%s: %q is not one of the options of choice %q (%s), declared at %s", name, given, given, name, strings.Join(q.Options, ", "), d.origin)
			}
			q.Answer, q.Given = given, true
			w.used[name] = true
		case q.Default != "":
			q.Answer = q.Default
		default:
			missing = append(missing, q)
			w.res.Asked = append(w.res.Asked, q)
			continue
		}
		w.resolved[name] = q.Answer
		w.res.Asked = append(w.res.Asked, q)
	}
	if len(missing) > 0 {
		return &MissingAnswerError{Choices: missing}
	}
	return nil
}

// apply puts a block's use and settings on the resolution, then the nested
// when blocks that match the answers.
func (w *walker) apply(doc Document, b Block, prefix string) error {
	name, file := doc.Recipe.Name, doc.Source.Ref
	for _, key := range sortedKeys(b.Use) {
		plugins, _ := UseList(b.Use[key])
		w.res.setUse(key, plugins, Origin{Recipe: name, File: file, Line: doc.Line(prefix + "use." + key)})
	}
	for _, plugin := range sortedKeys(b.Settings) {
		w.res.mergeSettings(plugin, b.Settings[plugin], name, file, func(path string) int {
			return doc.Line(prefix + "settings." + plugin + "." + path)
		})
	}
	type hit struct {
		choice, answer string
		block          Block
		line           int
	}
	var hits []hit
	for _, choice := range sortedKeys(b.When) {
		d, declared := w.declared[choice]
		_, answered := w.resolved[choice]
		at := fmt.Sprintf("%s:%d", file, doc.Line(prefix+"when."+choice))
		if !declared || !answered {
			return fmt.Errorf("%s: when.%s refers to a choice no recipe declares at this point; declare it with [choices.%s]", at, choice, choice)
		}
		for _, answer := range sortedKeys(b.When[choice]) {
			if len(d.choice.Options) > 0 && !slices.Contains(d.choice.Options, answer) {
				return fmt.Errorf("%s:%d: when.%s.%s: %q is not one of the options of choice %q (%s)", file, doc.Line(prefix+"when."+choice+"."+answer), choice, answer, answer, choice, strings.Join(d.choice.Options, ", "))
			}
			if answer == w.resolved[choice] {
				hits = append(hits, hit{choice, answer, b.When[choice][answer], doc.Line(prefix + "when." + choice + "." + answer)})
			}
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].line < hits[j].line })
	for _, h := range hits {
		path := prefix + "when." + h.choice + "." + h.answer + "."
		for _, cname := range byLine(doc, path+"choices.", h.block.Choices) {
			if _, asked := w.resolved[cname]; asked {
				return fmt.Errorf("%s:%d: choice %q was already asked and cannot be declared again here", file, doc.Line(path+"choices."+cname), cname)
			}
			w.declare(doc, cname, h.block.Choices[cname], path)
		}
		if err := w.answerPending(); err != nil {
			return err
		}
		if err := w.apply(doc, h.block, path); err != nil {
			return err
		}
	}
	return nil
}
