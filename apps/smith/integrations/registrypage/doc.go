// Package registrypage tells the address of a package's page on its registry
// (npmjs.com, pypi.org, nuget.org, crates.io, Maven Central, pub.dev,
// rubygems.org, packagist.org, pkg.go.dev, or the page of an Azure Artifacts
// feed named in smith.toml), for the Deps view and `smith up --json`. When the
// registry is private and its web page cannot be derived, the answer is none.
package registrypage
