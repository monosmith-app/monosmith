// Package splitrepo mirrors a released project into a repository of its own:
// a registry that reads a package's manifest at a repository root (Packagist
// reads composer.json there) cannot list a package that lives in a monorepo's
// subdirectory, so each release pushes that directory's history, split out,
// to the mirror and tags it there.
package splitrepo
