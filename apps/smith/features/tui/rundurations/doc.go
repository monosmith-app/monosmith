// Package rundurations keeps how long the tasks of the workspace took when they
// ran, a few runs per task, and estimates from them how long a run still takes
// ("about 20s left"). The history is a file in the workspace's local cache, which
// git ignores; it never goes to the remote cache. A task without history gives
// no estimate.
package rundurations
