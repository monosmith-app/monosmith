package declaredimports

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func lines(found []UndeclaredImport) []string {
	out := []string{}
	for _, u := range found {
		out = append(out, u.String())
	}
	return out
}

const manifest = `[project]
name = "acme-ocr"
version = "1.0.0"
dependencies = ["PyYAML>=6", "requests[socks]>=2"]

[project.optional-dependencies]
fast = ["orjson"]

[dependency-groups]
dev = ["pytest-cov"]
`

func workspace(t *testing.T) string {
	ws := t.TempDir()
	write(t, ws, "requirements-dev.txt", "pytest==8.0.0\n")
	write(t, ws, "packages/ocr/pyproject.toml", manifest)
	// pytest-cov is installed, so its modules are known; so are the others.
	site := ".venv/Lib/site-packages/"
	write(t, ws, site+"pytest_cov-5.0.0.dist-info/top_level.txt", "pytest_cov\n")
	write(t, ws, site+"pytest-8.0.0.dist-info/top_level.txt", "pytest\n_pytest\n")
	write(t, ws, site+"orjson-3.0.dist-info/RECORD", "orjson/__init__.py,sha256=x,1\norjson-3.0.dist-info/METADATA,,\n__pycache__/x.pyc,,\n")
	return ws
}

func TestCheckReportsOnlyWhatNoDeclarationProvides(t *testing.T) {
	ws := workspace(t)
	write(t, ws, "packages/ocr/src/ocr/__init__.py", "")
	write(t, ws, "packages/ocr/src/ocr/read.py", `"""Reads pages.

import docstring_only
"""
import os, sys
import yaml
import requests.adapters as adapters
from orjson import dumps
from . import sibling
from .sibling import thing
from ocr import read
import numpy
from helpers import shared
try:
    import ujson
except ImportError:
    import json as ujson
from typing import TYPE_CHECKING
if TYPE_CHECKING:
    import pandas


def lazy():
    import scipy
    return scipy
`)
	write(t, ws, "packages/ocr/src/ocr/helpers.py", "")
	write(t, ws, "packages/ocr/tests/test_read.py", "import pytest\nimport pytest_cov\nimport yaml\nimport faker\nimport numpy\n")

	got, err := Check(ws, "packages/ocr")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"src/ocr/read.py imports numpy, which pyproject.toml does not declare",
		"src/ocr/read.py imports scipy, which pyproject.toml does not declare",
		"tests/test_read.py imports faker, which pyproject.toml does not declare",
		"tests/test_read.py imports numpy, which pyproject.toml does not declare",
	}
	if !reflect.DeepEqual(lines(got), want) {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(lines(got), "\n"), strings.Join(want, "\n"))
	}
}

func TestCheckSkipsBuildFolders(t *testing.T) {
	ws := workspace(t)
	write(t, ws, "packages/ocr/src/ocr/__init__.py", "")
	for _, dir := range []string{"build/lib", "dist", "acme_ocr-1.0.0/src", "acme_ocr.egg-info"} {
		write(t, ws, "packages/ocr/"+dir+"/staged.py", "import numpy\n")
	}

	got, err := Check(ws, "packages/ocr")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("build folders must not be scanned, got\n%s", strings.Join(lines(got), "\n"))
	}
}

func TestCheckFlagsADevelopmentDependencyInShippedCode(t *testing.T) {
	ws := workspace(t)
	write(t, ws, "packages/ocr/src/ocr/read.py", "import requests, pytest_cov\n")
	got, err := Check(ws, "packages/ocr")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"src/ocr/read.py imports pytest_cov, declared only in development dependencies"}
	if !reflect.DeepEqual(lines(got), want) {
		t.Fatalf("got %v", lines(got))
	}
}

func TestCheckStaysSilentWhenADeclaredDistributionIsUnknown(t *testing.T) {
	ws := workspace(t)
	write(t, ws, "packages/ocr/pyproject.toml", strings.Replace(manifest, `"orjson"`, `"mystery-lib"`, 1))
	write(t, ws, "packages/ocr/src/ocr/read.py", "import numpy\n")
	got, err := Check(ws, "packages/ocr")
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v: an uninstalled, unknown distribution may provide numpy", got, err)
	}
}

func TestCheckStaysSilentWhenDependenciesAreDynamic(t *testing.T) {
	ws := workspace(t)
	write(t, ws, "packages/ocr/pyproject.toml", "[project]\nname = \"acme-ocr\"\ndynamic = [\"dependencies\"]\n")
	write(t, ws, "packages/ocr/read.py", "import numpy, requests\n")
	if got, err := Check(ws, "packages/ocr"); err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
	write(t, ws, "packages/ocr/requirements.txt", "# runtime\nrequests==2.0 --hash=sha256:0\n")
	got, err := Check(ws, "packages/ocr")
	if err != nil || len(got) != 1 || got[0].Module != "numpy" {
		t.Fatalf("a requirements.txt makes them known: %v, %v", got, err)
	}
}

func TestFileImportsLeavesOutWhatIsOptionalByIntent(t *testing.T) {
	src := "import a\nfrom b.c import d\nimport e as f, g.h\ntry:\n    import i\nexcept ImportError:\n    import j\n\nclass K:\n    def m(self):\n        import n\n"
	want := []string{"a", "b", "e", "g", "n"}
	if got := FileImports(src); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}
