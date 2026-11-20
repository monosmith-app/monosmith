// Package lockcarry keeps the versions a project's own lockfile locked when
// the project joins a workspace that has one lockfile. smith import drops the
// project's lockfile, so without this the first install resolves every range
// again and can move a package to a newer version that breaks the project.
// For npm, the entries of the project's package-lock.json go into the
// workspace's package-lock.json under the project's folder, and npm keeps
// them because they satisfy the ranges.
package lockcarry
