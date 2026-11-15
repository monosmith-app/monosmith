// Package tauriprepare is `smith tauri-prepare`, the tauri plugin's prepare and
// verify tasks: it builds a Tauri app's sidecar programs and copies them and
// its resources where `tauri build` looks, or checks the built app holds them.
// Hidden: the tasks call it through {smith}, people run smith run tauri-build.
package tauriprepare
