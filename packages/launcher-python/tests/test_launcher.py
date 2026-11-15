"""The launcher against a release signed by smith release's own minisign code, with a test key."""

from __future__ import annotations

import http.server
import threading
from collections.abc import Iterator
from pathlib import Path

import pytest

from monosmith_launcher.signature_check import check_archive, listed_digest
from monosmith_launcher.smith_run import resolve_program
from monosmith_launcher.version_choice import choose_version, compare_versions, pin_in

FIXTURES = Path(__file__).parent / "fixtures"
ARCHIVE = "monosmith_1.2.3_linux_amd64.zip"
PUBLIC_KEY = (FIXTURES / "test-key.pub").read_text().splitlines()[1]


def fixture(name: str) -> bytes:
    return (FIXTURES / name).read_bytes()


def test_an_archive_listed_in_a_correctly_signed_list_is_trusted() -> None:
    sig = fixture("checksums.txt.minisig").decode()
    assert check_archive(fixture(ARCHIVE), ARCHIVE, fixture("checksums.txt"), sig, PUBLIC_KEY) is None


def test_a_changed_list_another_key_or_a_changed_archive_is_not_trusted() -> None:
    sig = fixture("checksums.txt.minisig").decode()
    sums = fixture("checksums.txt")
    assert "signature" in (check_archive(fixture(ARCHIVE), ARCHIVE, sums + b"x", sig, PUBLIC_KEY) or "")
    assert "signature" in (check_archive(fixture(ARCHIVE), ARCHIVE, sums, sig) or "")
    assert "does not match its checksum" in (check_archive(b"other", ARCHIVE, sums, sig, PUBLIC_KEY) or "")
    other = "monosmith_1.2.3_darwin_arm64.zip"
    assert "does not list" in (check_archive(fixture(ARCHIVE), other, sums, sig, PUBLIC_KEY) or "")


def test_a_digest_is_found_by_the_file_name_of_its_line() -> None:
    assert listed_digest("ab12  one.zip\ncd34  two.zip\n", "two.zip") == "cd34"
    assert listed_digest("ab12  one.zip\n", "three.zip") is None


def test_the_version_comes_from_smith_version_then_the_workspace_pin(tmp_path: Path) -> None:
    (tmp_path / "smith.toml").write_text('[ci]\nsmith = "9.9.9"\n[workspace]\nsmith = "2.50.0" # pinned\n')
    deep = tmp_path / "apps" / "web"
    deep.mkdir(parents=True)
    assert choose_version({}, deep) == "2.50.0"
    assert choose_version({"SMITH_VERSION": "3.0.0-beta.1"}, deep) == "3.0.0-beta.1"
    assert choose_version({"SMITH_VERSION": "latest"}, deep) is None
    assert pin_in('[workspace]\nsmith = "latest"\n') is None
    with pytest.raises(ValueError, match="SMITH_VERSION"):
        choose_version({"SMITH_VERSION": "^3"}, deep)


def test_a_prerelease_sorts_before_its_release() -> None:
    assert compare_versions("3.0.0-beta.1", "3.0.0") < 0
    assert compare_versions("2.10.0", "2.9.0") > 0


@pytest.fixture
def mirror() -> Iterator[tuple[str, list[str]]]:
    """A mirror of one release (SMITH_DOWNLOAD_URL), serving the files smith release attaches."""
    requests: list[str] = []
    files = {"/latest.txt": b"1.2.3\n"}
    for name in (ARCHIVE, "checksums.txt", "checksums.txt.minisig"):
        files[f"/smith%401.2.3/{name}"] = fixture(name)

    class Handler(http.server.BaseHTTPRequestHandler):
        def do_GET(self) -> None:  # noqa: N802 - the name http.server calls
            requests.append(self.path)
            body = files.get(self.path)
            self.send_response(200 if body else 404)
            self.end_headers()
            if body:
                self.wfile.write(body)

        def log_message(self, *_: object) -> None:
            pass

    server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    yield f"http://127.0.0.1:{server.server_address[1]}", requests
    server.shutdown()


def test_the_newest_release_is_downloaded_once_checked_and_kept(tmp_path: Path, mirror: tuple[str, list[str]]) -> None:
    base, requests = mirror
    env = {"SMITH_DOWNLOAD_URL": base, "SMITH_HOME": str(tmp_path)}
    program = resolve_program(env, tmp_path, "linux", "x86_64", lambda _: None, PUBLIC_KEY)
    assert program == tmp_path / "versions" / "1.2.3" / "smith"
    assert program.read_text() == "#!/bin/sh\necho smith 1.2.3\n"
    requests.clear()
    assert resolve_program(env, tmp_path, "linux", "x86_64", lambda _: None, PUBLIC_KEY) == program
    assert requests == []


def test_a_release_signed_with_another_key_is_not_installed(tmp_path: Path, mirror: tuple[str, list[str]]) -> None:
    base, _ = mirror
    env = {"SMITH_DOWNLOAD_URL": base, "SMITH_HOME": str(tmp_path), "SMITH_VERSION": "1.2.3"}
    with pytest.raises(OSError, match="was not installed"):
        resolve_program(env, tmp_path, "linux", "x86_64", lambda _: None)
    assert not (tmp_path / "versions" / "1.2.3").exists()
