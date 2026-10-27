// Package dependencies lists and changes the gems a Ruby workspace declares:
// each gem's gemspec and the workspace's Gemfile, so `smith sync`, `smith up`
// and the one-version check cover Ruby. Installing the bundle and adding or
// removing a gem's dependencies go through Bundler and the gemspec.
package dependencies
