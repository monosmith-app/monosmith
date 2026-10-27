// Package linting decides how a PHP package is analysed: PHPStan, with the
// package's own configuration when it has one (phpstan.neon, phpstan.neon.dist
// or phpstan.dist.neon in the package directory, which is where PHPStan looks),
// and smith's default otherwise.
package linting
