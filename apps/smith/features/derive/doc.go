// Package derive is `smith derive`: files that smith writes from facts and
// the project graph, in the ownership mode the renderer chose, and an
// ownership record (smith.lock) that lets smith rewrite a file when its inputs
// change and never overwrite a hand edit. It is opt-in ([derive] enabled in
// smith.toml). Plugins render; this package decides what to write. See ADR
// 0046.
package derive
