// Package cachekey makes the secret that signs the remote cache's objects
// (SMITH_CACHE_KEY) and hands it to its owner without printing it: it
// generates 32 random bytes as 64 hex characters, puts them on the clipboard,
// clears the clipboard again if it still holds the key, and offers to set the
// key as a user environment variable. The key lives in memory only; nothing
// here writes it to a file, a log or smith.toml.
package cachekey
