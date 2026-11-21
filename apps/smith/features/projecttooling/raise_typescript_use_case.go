package projecttooling

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TypeScriptRange is the TypeScript smith's tooling raises a TypeScript 4 (or
// older) to. smith's ESLint config brings the newest @types/node into the
// workspace's node_modules, and TypeScript loads every @types package it
// finds there; TypeScript 4 cannot read its syntax, so tsc fails. TypeScript
// 5.9 reads it and keeps the compiler options of a TypeScript 4 project: the
// 6.x that smith's stacks write (toolchainversions) turns some of them
// (moduleResolution "node") into errors.
const TypeScriptRange = "^5.9.3"

// typeScriptSections are the package.json sections whose typescript range is
// raised; peerDependencies are a package's published contract.
var typeScriptSections = []string{"dependencies", "devDependencies", "optionalDependencies"}

// PlanTypeScript returns the fix that raises a TypeScript older than 5 to
// TypeScriptRange in the root package.json and in the project's package.json,
// or nil when neither names one.
//
// Parameters: root is the workspace folder (absolute); dir the project folder,
// workspace-relative.
// Returns: the fix, safe and without questions; an error for a package.json
// that cannot be read or is not JSON.
func PlanTypeScript(root, dir string) (*doctorfix.Planned, error) {
	files := []string{"package.json"}
	if dir != "." {
		files = append(files, path.Join(dir, "package.json"))
	}
	var edits []protocol.FileEdit
	var old []string
	for _, rel := range files {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		doc, err := jsondoc.ParseDocument(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
		changed := false
		for _, section := range typeScriptSections {
			deps := doc.Object(section, false)
			if deps == nil {
				continue
			}
			current, ok := deps.String("typescript")
			if !ok || !olderThanFive(current) {
				continue
			}
			deps.SetString("typescript", TypeScriptRange)
			old = append(old, current)
			changed = true
		}
		if changed {
			edits = append(edits, protocol.FileEdit{Path: rel, Content: string(doc.Bytes())})
		}
	}
	if len(edits) == 0 {
		return nil, nil
	}
	return &doctorfix.Planned{
		Fix: protocol.Fix{
			Description: fmt.Sprintf("raise TypeScript from %s to %s: smith's ESLint config installs a @types/node that TypeScript 4 cannot read, so tsc would fail", old[0], TypeScriptRange),
			Safe:        true,
			Files:       edits,
		},
		Checks: []string{"tooling: TypeScript reads the types smith's tools install"},
	}, nil
}

// raiseTypeScript shows the TypeScript raise (PlanTypeScript) and applies it
// when the person agrees, or with --yes.
//
// Parameters: ctx stops the walk; in is the decision; name the project; o
// the run options.
// Returns: the step left when the raise was not applied; an error for a file
// that cannot be read or written.
func raiseTypeScript(ctx context.Context, in Input, name string, o RunOptions) ([]string, error) {
	planned, err := PlanTypeScript(in.Root, in.Dir)
	if err != nil || planned == nil {
		return nil, err
	}
	opts := doctorfix.Options{Root: in.Root, Out: o.Out, Interactive: o.Interactive, Yes: o.Yes}
	if o.In != nil {
		opts.In = o.In
	}
	fmt.Fprintln(o.Out)
	walked, err := doctorfix.Walk(ctx, []doctorfix.Planned{*planned}, opts)
	if err != nil {
		return nil, err
	}
	if len(walked.Applied) > 0 {
		return nil, nil
	}
	return []string{name + " keeps TypeScript 4, which cannot read the @types/node that smith's ESLint config installs, so tsc fails. Set typescript to " + TypeScriptRange + " in package.json, then run smith install"}, nil
}

// olderThanFive reports whether a typescript range accepts only versions
// before 5: its lowest version has a major under 5.
func olderThanFive(r string) bool {
	floor, ok := eslintpeer.RangeFloor(r)
	if !ok {
		return false
	}
	major, _, _ := strings.Cut(floor, ".")
	n, err := strconv.Atoi(major)
	return err == nil && n < 5
}
