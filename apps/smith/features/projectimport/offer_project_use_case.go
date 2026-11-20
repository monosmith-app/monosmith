package projectimport

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/apps/smith/features/manifestoffer"
	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/shared/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// offer settles a source with no manifest at its top. A project whose
// manifest is below the top is offered (--path answers): the import then
// takes that sub-folder, with its own history. A folder of sources is
// offered a manifest of its language's default stack (--manifest answers).
// Nothing is changed except the source's temporary clone.
//
// Parameters: src is the source, narrowed in place when a sub-folder is
// chosen; claims are the plugins' claims; o are the options.
// Returns: the candidate, the language whose manifest the import writes ("" for
// none), or a problem that says what stopped it and the flag that answers.
func offer(src *Source, claims []adoptmove.Claim, o Options) (adoptmove.Candidate, string, error) {
	ask := operationplan.ShouldAsk(o.Env, o.Yes) && !o.DryRun
	in, _ := o.In.(*bufio.Reader)
	if nested := manifestoffer.FindNested(src.Dir, claims); len(nested) > 0 {
		sub, err := chooseNested(nested, o, ask, in)
		if err != nil {
			return adoptmove.Candidate{}, "", err
		}
		if *src, err = Narrow(*src, sub, o.Out); err != nil {
			return adoptmove.Candidate{}, "", err
		}
		c, err := adoptmove.Detect(src.Dir, claims)
		if err != nil {
			return c, "", problem.New("smith finds no project at the top of "+sub+" after all", "import it with --path and the folder that holds the manifest")
		}
		return c, "", nil
	}
	finds := manifestoffer.Sources(src.Dir, claims)
	if len(finds) == 0 {
		return adoptmove.Candidate{}, "", problem.New("smith finds no project at the top of "+o.Source,
			"smith knows a project by one of these manifests at the top of its folder: "+ManifestNames(claims),
			"import the sub-folder that holds the manifest (--path), or add the manifest to the project first")
	}
	answer := o.Manifest
	if answer == "" && o.DryRun {
		// A dry run changes nothing: it shows the suggested manifest, and the real run asks.
		answer = finds[0].Language
	}
	f, ok, err := manifestoffer.Settle(manifestoffer.Question{
		Folder: o.Source, Finds: finds, Answer: answer, Ask: ask, Command: "smith import " + o.Source, In: in, Out: o.Out,
	})
	if err != nil {
		return adoptmove.Candidate{}, "", err
	}
	if !ok {
		return adoptmove.Candidate{}, "", problem.New("nothing was imported: "+o.Source+" has no manifest, and none is written",
			"write the manifest yourself, then run smith import again", "or run smith import "+o.Source+" --manifest "+finds[0].Language)
	}
	c := adoptmove.Candidate{Ecosystem: f.Language, Manifest: f.Manifest, Name: adoptmove.DeriveName("", FolderName(*src)), App: manifestoffer.LooksLikeProgram(f.Files)}
	return c, f.Language, nil
}

// chooseNested settles which project below the top of the source the import
// takes: a terminal asks, else smith refuses with the --path that answers.
// Several projects make the source a workspace, which the text says.
func chooseNested(nested []manifestoffer.Nested, o Options, ask bool, in *bufio.Reader) (string, error) {
	cmd := "smith import " + o.Source + " --path "
	if len(nested) == 1 && o.DryRun {
		return nested[0].Path, nil
	}
	if !ask || in == nil {
		n := nested[0]
		if len(nested) == 1 {
			return "", problem.New(fmt.Sprintf("the manifest of %s is not at its top: it is %s/%s", o.Source, n.Path, n.Manifest),
				cmd+n.Path+" imports that project (a git source keeps the commits of that folder)")
		}
		return "", problem.New(fmt.Sprintf("%s holds %d projects below its top (%s), so it is a workspace rather than one project", o.Source, len(nested), nestedList(nested)),
			"smith adopt makes it a workspace: clone it and run smith adopt in it",
			"or import one project: "+cmd+n.Path)
	}
	options := make([]choiceprompt.Option, 0, len(nested)+1)
	for _, n := range nested {
		options = append(options, choiceprompt.Option{Label: n.Path, Value: n.Path, Hint: "a " + n.Language + " project (" + n.Manifest + ")"})
	}
	options = append(options, choiceprompt.Option{Label: "none", Value: manifestoffer.None, Hint: "stop: import nothing"})
	prompt := fmt.Sprintf("The manifest of %s is not at its top. Import the project below it? (--path <folder> answers)", o.Source)
	if len(nested) > 1 {
		prompt = fmt.Sprintf("%s holds %d projects below its top, so it is a workspace (smith adopt takes all of them). Import one of them? (--path <folder> answers)", o.Source, len(nested))
	}
	answer := choiceprompt.Ask(prompt, options, nested[0].Path, in, o.Out)
	if answer == manifestoffer.None {
		return "", problem.New("nothing was imported", "smith adopt takes a repository with several projects as a workspace", "or "+cmd+"<folder> imports one")
	}
	return answer, nil
}

func nestedList(nested []manifestoffer.Nested) string {
	names := make([]string, 0, len(nested))
	for i, n := range nested {
		if i == 5 {
			names = append(names, "...")
			break
		}
		names = append(names, n.Path)
	}
	return strings.Join(names, ", ")
}
