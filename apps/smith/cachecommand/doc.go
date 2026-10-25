// Package cachecommand is `smith cache`: what the local task cache holds
// (status), trimming it to its limit (prune), deleting it (clear) and checking
// it for damage (verify). It acts on the local store only and never touches the
// remote cache; the work itself is in cachemaintenance.
package cachecommand
