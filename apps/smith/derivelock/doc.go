// Package derivelock is the ownership record of derived files, smith.lock:
// for each file smith derives, who owns it (smith or the person), which
// renderer wrote it at which version, and a hash of the content smith last
// wrote. A file whose content no longer matches its hash was edited by hand.
// See ADR 0046.
package derivelock
