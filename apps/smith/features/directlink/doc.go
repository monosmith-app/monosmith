// Package directlink makes the smith command start the program itself, with no
// Node, Python or other launcher in front. A package manager puts a small entry
// on PATH (smith.cmd, a symlink, a launcher program) that starts the launcher
// of its own language, which stays alive as long as smith runs: about 110 ms
// slower to start and about 40 MB more memory for every smith process. Link
// finds those entries, whatever the manager, and points each at the program.
//
// On Windows the entries call a hard link of the installed program in
// ~/.smith/run, so the installed file is never the one running and an update can
// replace it while smith runs in other terminals. On other systems the entries
// call the installed program, which the system lets an update replace.
//
// The first run of a version does this by itself, once, and says so. The package
// manager writes its entries again when it updates smith, so the next version's
// first run links again. smith link does the same by hand, shows what runs now
// and undoes it.
package directlink
