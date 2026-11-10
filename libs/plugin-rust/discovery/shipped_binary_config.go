package discovery

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// DebugProfile is the cargo profile `cargo build` writes: the folder below the
// target folder (dist/cargo/debug) that an app's build ships from.
const DebugProfile = "debug"

// withShippedBinary makes an app's build task ship its program: smith links
// dist/cargo/<profile>/<bin> (bin.exe on Windows) into the project's shipped
// folder, which becomes the task's one output. The shared target folder
// dist/cargo is never an output: a cache restore would empty it for every
// crate.
//
// Parameters: build is the task to change; root the crate's project folder;
// profile the cargo profile folder (DebugProfile); bin the program's name.
// It returns the changed task.
func withShippedBinary(build protocol.Task, root, profile, bin string) protocol.Task {
	file := protocol.BuildDir + "/cargo/" + profile + "/" + bin
	build.Outputs = []string{protocol.BuildDirFromProject(root)}
	build.Ship = []string{protocol.FromProject(root, file), protocol.FromProject(root, file+".exe")}
	return build
}
