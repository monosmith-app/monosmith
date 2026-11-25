// Package projectstandard offers smith's standard configuration files to the
// projects of a converted repository. It asks the plugins (protocol method
// project/standard) for the standard version of each project's files, compares
// them with the files on disk, and either writes them (smith adopt --init and
// smith convert, with --project-config replace) or turns the differences into
// doctor findings whose fix writes them. A whole file it replaces is kept
// beside it as <path>.pre-smith; a manifest only has the listed keys set.
package projectstandard
