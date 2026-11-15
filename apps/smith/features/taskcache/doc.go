// Package taskcache decides whether a task can be skipped and replays it when
// it can. A task's key is a SHA-256 over a versioned, canonical encoding of
// everything that can change its result; its outputs and log are stored as
// content-addressed blobs (CAS) indexed by an action-cache entry (AC), the
// same split Bazel's remote cache uses, behind a Storage interface so a remote
// backend can be added without touching the rest of smith.
package taskcache
