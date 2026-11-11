package discovery

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// DebugProfile is the cargo profile `cargo build` writes: the folder below the
// target folder (dist/cargo/debug) that an app's build ships from.
const DebugProfile = "debug"

// ReleaseProfile is the folder `cargo build --release` writes
// (dist/cargo/release): what `smith build --release` ships from.
const ReleaseProfile = "release"

// withReleaseProfile gives a cargo task its release profile: the same
// command with --release. The cache key holds the command, so each profile
// has its own entry.
//
// Parameters: t is the task to change. It returns the changed task.
func withReleaseProfile(t protocol.Task) protocol.Task {
	t.Release = &protocol.TaskProfile{Command: append(slices.Clone(t.Command), "--release")}
	return t
}

// withShippedBinaries makes an app's build task ship its programs: smith
// links dist/cargo/<profile>/<bin> (bin.exe on Windows) of each program into
// the project's shipped folder, which becomes the task's one output; the
// debug profile by default, the release profile with --release. The shared
// target folder dist/cargo is never an output: a cache restore would empty it
// for every crate.
//
// Parameters: build is the task to change, with its release profile set
// (withReleaseProfile) or not; root the crate's project folder; bins the
// programs' names. It returns the changed task.
func withShippedBinaries(build protocol.Task, root string, bins ...string) protocol.Task {
	build.Outputs = []string{protocol.BuildDirFromProject(root)}
	build.Ship = shippedBinaries(root, DebugProfile, bins)
	if build.Release == nil {
		build = withReleaseProfile(build)
	}
	release := *build.Release
	release.Ship = shippedBinaries(root, ReleaseProfile, bins)
	build.Release = &release
	return build
}

// shippedBinaries are the globs, relative to the project folder root, of each
// program in bins as cargo writes it into the profile folder.
func shippedBinaries(root, profile string, bins []string) []string {
	var ship []string
	for _, bin := range bins {
		file := protocol.BuildDir + "/cargo/" + profile + "/" + bin
		ship = append(ship, protocol.FromProject(root, file), protocol.FromProject(root, file+".exe"))
	}
	return ship
}
