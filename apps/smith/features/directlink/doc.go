// Package directlink makes the smith command start the program itself, with no
// Node, Python or other launcher in front. The registry packages put a small
// entry on PATH (smith.cmd, a symlink, a script) that starts the launcher of their
// own language, which stays alive as long as smith runs: about 110 ms slower to
// start and about 40 MB more memory for every smith process. Link finds those
// entries and points each at a hard link of the program in ~/.smith/run, so the
// installed file is never the one running. A program that runs from there still
// runs the version the workspace pins: package versionhandoff hands the run on
// (ADR 0111).
//
// The first run of a version does this by itself, once, and says so. smith link
// does the same by hand, shows what runs now and undoes it.
package directlink
