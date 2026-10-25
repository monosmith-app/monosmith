// Package cachemaintenance looks after the local task cache: it keeps the store
// under [cache] max_size_mb after a run, reports what is in it, checks it for
// damage and clears it. It works on a cache directory, never on the remote
// tier, and it is safe beside other smith processes: it only removes files
// through taskcache.Storage, tolerates files that vanish, and leaves alone
// anything written or used in the last few minutes.
package cachemaintenance
