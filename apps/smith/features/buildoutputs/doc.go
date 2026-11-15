// Package buildoutputs compares the directories a Node project's build script
// writes (--outdir, --outfile, --declarationDir, -o) with the outputs its
// build task declares. A cache hit restores only the declared directories, so
// a build that writes somewhere else leaves it empty on a warm cache. The
// package finds the gap and declares the missing outputs in smith.toml; smith
// convert, smith upgrade and smith doctor all use it.
package buildoutputs
