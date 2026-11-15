// Package remotecache shares the task cache between machines: CI fills it and
// developers read it. A remote store (an HTTP object store, S3, Azure Blob)
// sits behind the local cache. Every remote object is signed with a key only
// the team holds (HMAC-SHA256), and an object that fails verification, like
// an unreachable store, is a miss: a remote cache can make a run slower,
// never wrong.
package remotecache
