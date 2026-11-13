// Package eslintpeer is the one source of the ESLint that smith's shared
// config (@monosmith/eslint-config) runs with. smith's conversion and the
// eslint toolchain plugin both raise a workspace's eslint to it: an older
// ESLint cannot load the config, and every lint crashes.
package eslintpeer
