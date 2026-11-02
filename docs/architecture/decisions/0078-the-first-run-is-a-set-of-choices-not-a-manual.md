# ADR 0078: The first run is a set of choices, not a manual

- Status: accepted
- Date: 2026-11-02

## Context

A review of the terminal interface played a junior developer who had read no documentation
(`junior-tui`, beta 16). The person wanted "a React app with a Python backend, as Azure
Functions". Smith asked for a kind, then for one of 18 languages, then for tools. A failed first
check left the person unsure that the project existed. The result of a change said "not committed"
and offered nothing to do about it. One key, `r`, meant "reload" on every view and "run again" on
the Run view, and a re-run of an add failed. A new workspace asked six questions to a person who
knew only the name.

The owner rules apply (ADR 0072): every decision is a choice with its consequence in plain words and
the default marked; smith never commits unasked; smith does the hard work so the person does not.

## Decision

1. **Add starts from the outcome.** The first question is "What are you building?": a website, an
   API, a function app, a desktop app, a library, a package. `compose.Outcomes` builds the list from
   the catalogue of the composer (ADR 0076), so the languages of a website, a function app or a
   desktop app are those that have a choice on its layer, and the words that find a line are the
   names, titles and aliases of those choices. The language is asked only when more than one fits. A
   website, a desktop app and a function app then ask their framework, shell or cloud; the answer
   starts ticked in the choices question. "I know the language" keeps the old path. `smith add` on
   the command line does not change.
2. **A run that ends with a decision to make ends in a dialog of choices.** The first line of the dialog
   is always the one that does nothing.
   - An add that could not be checked because a tool is missing says the project exists, why it could not be
     checked, and offers: open the install page, install it with smith (ADR 0077), check again, continue
     without checking.
   - A change that was not committed offers "Leave them uncommitted" (the default), "Commit only the
     files this operation changed" and "Commit everything git shows as changed". The box in the confirmation stays unticked.
3. **`smith commit <operation words>` commits after the fact.** It lists every changed file, asks, and
   makes one commit named from the operation and recorded for `smith undo`. smith writes down the files of an
   uncommitted operation (`.smith/uncommitted-operation.json`), so the commit takes only those by
   default and never sweeps up the person's own work; taking everything is an explicit choice
   (`--all`), and with no terminal the command refuses until a flag answers.
4. **A new workspace asks how to set it up first.** "Start with the defaults" lists what is chosen (derived from
   the questions of `smith new`) and how to change each later; "Customise" asks the questions. The
   questions that carry risk (first commit, secrets) are asked either way.
5. **A command that makes or removes something is not run again.** `add`, `new`, `remove`, `adopt` and
   `convert` are not offered by the re-run key or the action; the key says why. The footer names the two
   meanings: `r run again` on the Run view, `r reload view` elsewhere, and the help says what each does.
6. **A word is explained where it first appears.** Stack, slice (vertical-slices), plugin, recipe,
   framework, shell and host get one line under the question that uses them, from the glossary and the
   layers of the composer.

The VS Code extension has the same first question, the same notices and the same buttons; the JetBrains
plugin records them in the parity table.

## Consequences

- A test pins the first question, the dialogs and the footer words; changing any of them changes the tests
  in the same commit.
- `smith commit` is a new mutating command: it is in the action catalogue (with a plan flag) and has a
  manual page.
- The outcome list has six fixed entries. What they hold comes from the plugins, so a new framework, shell
  or host appears with no change in smith.
