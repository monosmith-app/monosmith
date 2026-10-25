// Package conformance checks that a plugin keeps the protocol's promises:
// it identifies itself, its roles match what it implements, the stacks it
// publishes are complete, every method it advertises answers, requests sent
// together are answered as its declared concurrency allows, a plugin that
// advertises cancel stops a cancelled request, and, for a language plugin,
// the projects it generates are discovered as generated.
//
// Plugin authors run it in their tests through an in-process connection;
// `smith plugin check` runs it against any plugin binary.
package conformance
