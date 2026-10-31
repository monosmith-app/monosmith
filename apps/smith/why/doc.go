// Package why is `smith why <file>`: it says who owns a file, what produced
// it, the rule or renderer behind it, and how to change it. It reads the
// ownership record smith.lock, the derive plan, the project graph and the
// stacks of smith.toml. It is not `smith why-miss`, which explains a cache
// miss (see ADR 0048).
package why
