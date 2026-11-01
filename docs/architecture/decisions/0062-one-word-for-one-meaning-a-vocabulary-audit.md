# ADR 0062: One word for one meaning, a vocabulary audit

- Status: accepted (one rename done, the five-noun model rejected)
- Date: 2026-11-01

## Context

Ideas A1 and A3 of the design review: a newcomer meets stack, recipe, convention, task group, default
stack, kind, plugin role, slot, level, toolchain, scaffold, structure and deploy. The proposal cut
them to five nouns (Intent, Pack, Profile, Exemplar, Owner) and put verbs (test, lint, build) in front
of tools. The owner's rule for this batch: do not rename core concepts blindly.

## Audit

Counts are files of Go code that print or define the word, then lines of user documentation
(`docs/manual`, `docs/reference`, `docs/guides`), on 2026-11-01.

| Word | Go files | Doc lines | Meanings | Verdict |
|---|---|---|---|---|
| stack | 94 | 178 | the set of plugins one project is built with | keep: one meaning |
| recipe | 99 | 329 | a team's data file that picks the stack and asks questions | keep: different from a stack (it chooses, a stack is chosen) |
| convention | 26 | 42 | a plugin applied to many projects | keep |
| task group | 2 | 1 | `[[task_groups]]` overrides for selected projects | keep ([ADR 0058](0058-folder-rules-are-task-groups-with-a-path-selector.md)) |
| default stack | 6 | 5 | the stack of a language when a project records none | keep: the plain words say what it is |
| kind | 162 | 174 | app, lib or package; also the `node-app` names `smith kinds` lists | keep; the glossary already separates the two |
| role | - | - | the part a plugin plays | the one word, see below |
| level | 78 | 102 | (1) the same as role, (2) `--level` depth of adopt and convert, (3) patch/minor/major in `smith up` | **role** replaces meaning 1; meanings 2 and 3 stay, they are not confusable once 1 is gone |
| slot | 32 | 41 | (1) a task name a toolchain plugin owns, (2) a `smith:slot` marker in a pipeline | keep: two meanings, glossary explains; renaming the marker would break every committed pipeline |
| toolchain | 54 | 107 | (1) a plugin role, (2) a runtime CI installs | keep: glossary explains; the role is the older meaning |
| scaffold | 57 | 77 | a plugin role that writes starter files | keep |
| structure | 24 | 44 | a plugin role that checks layout | keep |
| deploy | 47 | 59 | a plugin role that adds deploy steps | keep |

## Decision

1. One rename where two words meant one thing: "level" as the part a plugin plays becomes "role".
   `smith plugin list` shows `ROLES`, `smith recipe new` takes `--fixed role=plugin`, and the help
   of `explain`, `recipe check` and the MCP tools, the validation message and the docs say role.
   Code identifiers (`Levels`, `LevelPlan`), the recipe file's `[levels.<role>]` table and the
   protocol's `requires.level` keep their names: they are formats of files that exist.
2. The glossary (`docs/reference/glossary.md`) is the map from the old word to the new one and says
   which words carry two meanings and why.
3. The five-noun model is rejected. Pack, Profile and Intent would replace stack, recipe and
   convention, which have three different jobs (what a project is made of, who chooses it, what is
   applied to many), and would rename about 240 files and 700 doc lines for a word count of five
   that the product does not have in the first place: a newcomer meets stack, recipe, kind and
   project in the first hour and the rest only when writing a plugin. "Exemplar" (B1) is a feature,
   not a word, and "owner" already names who maintains a derived file (`smith.lock`).
4. A3 (verbs in front of tools) is already how the commands read: `smith test`, `lint`, `build`,
   `typecheck` are the verbs, and the tool behind a verb is the toolchain plugin of a stack.

## Cost

Anyone with a recipe script that passes `--fixed level=plugin` must write `role=plugin`. smith is in
beta and has no users, so there is no alias.
