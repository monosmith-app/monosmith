// Package typechecking decides how a Python project is type-checked: mypy, with
// the repository's own configuration when it has one (mypy.ini, .mypy.ini or a
// [tool.mypy] table in a pyproject.toml, or a [mypy] section in setup.cfg), and
// smith's default otherwise.
package typechecking
