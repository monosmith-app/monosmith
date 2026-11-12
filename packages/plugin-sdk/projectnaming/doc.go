// Package projectnaming builds the default package, crate, gem, module and
// namespace name of a new project, for every language plugin: one prefix (the
// workspace namespace), one skip-prefix rule and one sanitizer per target
// syntax (ADR 0098). A plugin never builds such a name by hand.
package projectnaming
