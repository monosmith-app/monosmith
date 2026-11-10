# ADR 0061: Teach by example drafts the facts, not the plugin

- Status: accepted (first slice)
- Date: 2026-11-01

## Context

Idea E2 of the design review: point smith at a sample project in a language it does not know, and
get a draft of what a plugin must say, confirmed by a conformance run, so adding a language takes an
afternoon.

## Challenge

The idea says "draft a pack". There is no pack format for languages: a language plugin is an
executable that answers the protocol (ADR 0053). A declarative description that smith interprets
would be a second plugin system and a second contract to keep equal to the first. The conformance
suite also checks a plugin that exists; it cannot check a description.

## Decision

`smith plugin draft <sample> --lang x` inspects the sample and writes the facts a plugin author
needs (manifests, the version line, tasks, publish candidates) as a TOML file with the source of
each value, and prints what is still to write. The author, or an assistant, writes the plugin with
the existing scaffold (`go-smith-plugin`) and runs `smith plugin check` on it. smith reads the draft
nowhere.

## Cost

The detection is heuristic (name and version keys, Makefile and justfile targets, `scripts`, CI
publish lines). It is wrong for a build system with no such files, and it says "guess" where it
guesses. Nothing is run, so nothing in the sample can harm the machine.

## Smallest first slice (done)

Manifest, task and publish detection; the draft file; the list of what is left; a test on a sample
file system. Not done: reading `Taskfile.yml`, `tox`, `build.gradle` targets; drafting the plugin
source from the facts (that is the assistant's work, with the plugin SDK guides); running the
conformance suite automatically.

## Sources

- `apps/smith/plugindraft/`
