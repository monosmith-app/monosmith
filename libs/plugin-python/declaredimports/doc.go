// Package declaredimports checks that a Python package declares what it
// imports. A package's imports are resolved from whatever installs it, so one
// its pyproject.toml does not declare works in the workspace (every project
// shares one venv) and breaks for whoever installs it from PyPI. The check
// reports only what it can be sure about: an import that no declared
// distribution provides, by name, by a short table of well-known differences
// between a distribution's name and its import name (PyYAML is imported as
// yaml), or by the distribution's metadata in the workspace venv. When a
// declared distribution is neither known nor installed, smith cannot tell
// what it provides and reports nothing for that package.
package declaredimports
