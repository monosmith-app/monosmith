// Package cachewizard is the remote cache setup as screens of the interactive
// shell: it opens from Enter on `cache.remote` in the Settings tab and from the
// palette. The decisions come from the same engine as `smith cache setup`
// (package cachesetup): the engine runs in the background behind the Dialogue
// port (Bridge), and the model draws each question as a list, a typed line, a
// confirmation that shows the exact command first, a working screen with a
// spinner for each cloud listing, a page of steps, or the key screen with its
// clickable buttons. The model runs nothing and holds no secret.
package cachewizard
