// Package linting decides how a Python project is linted: ruff, with the
// repository's own configuration when it has one (ruff.toml, .ruff.toml or a
// [tool.ruff] table in a pyproject.toml), and smith's default otherwise.
package linting
