// Package testrunning decides how a Python project's tests run: pytest, with the
// repository's own configuration when it has one (pytest.ini, .pytest.ini, a
// [tool.pytest] or [tool.pytest.ini_options] table in a pyproject.toml, a
// [pytest] section in tox.ini or a [tool:pytest] section in setup.cfg), and
// smith's default otherwise.
package testrunning
