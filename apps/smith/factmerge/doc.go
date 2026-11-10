// Package factmerge merges the shared facts of one project: the layers that
// smith.toml and plugins publish, narrowest scope first, into one value per
// key that remembers where it came from. The merge is pure; callers collect
// the layers. See ADR 0042.
package factmerge
