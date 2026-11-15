"""Reads smith's release files from GitHub or a mirror (SMITH_DOWNLOAD_URL)."""

from __future__ import annotations

import json
import urllib.error
import urllib.request
from collections.abc import Mapping
from typing import Any
from urllib.parse import urlsplit

_GITHUB = "https://github.com/monosmith-app/monosmith/releases/download"
_API = "https://api.github.com/repos/monosmith-app/monosmith/releases?per_page=100"


def download_base(env: Mapping[str, str]) -> str:
    """Return the address a release's files are under: SMITH_DOWNLOAD_URL (a mirror), else GitHub.

    Args:
        env: The environment.

    Returns:
        The base, without a trailing slash.
    """
    return (env.get("SMITH_DOWNLOAD_URL") or _GITHUB).rstrip("/")


def _get(url: str, headers: Mapping[str, str] | None = None) -> bytes:
    request = urllib.request.Request(url, headers={"User-Agent": "monosmith-launcher", **(headers or {})})
    try:
        with urllib.request.urlopen(request, timeout=60) as res:  # noqa: S310 - https or the user's mirror
            body: bytes = res.read()
            return body
    except urllib.error.HTTPError as error:
        raise OSError(f"{url} answered {error.code}") from error
    except urllib.error.URLError as error:
        raise OSError(
            f"cannot reach {urlsplit(url).netloc} ({error.reason}); behind a firewall, "
            "set SMITH_DOWNLOAD_URL to a mirror of smith's release archives"
        ) from error


def release_file(env: Mapping[str, str], version: str, name: str) -> bytes:
    """Return one file of a release: <base>/smith@<version>/<name>.

    Args:
        env: The environment.
        version: An exact version.
        name: The file's name.

    Returns:
        Its bytes.

    Raises:
        OSError: When the file cannot be downloaded.
    """
    return _get(f"{download_base(env)}/smith%40{version}/{name}")


def newest_release(env: Mapping[str, str]) -> str:
    """Return the newest stable smith: the mirror's latest.txt, or the newest smith@ release on GitHub.

    GITHUB_TOKEN or GH_TOKEN, when set, raises GitHub's rate limit.

    Args:
        env: The environment.

    Returns:
        The version.

    Raises:
        OSError: When no release can be found.
    """
    if env.get("SMITH_DOWNLOAD_URL"):
        return _get(f"{download_base(env)}/latest.txt").decode().strip()
    headers = {"Accept": "application/vnd.github+json"}
    token = env.get("GITHUB_TOKEN") or env.get("GH_TOKEN")
    if token:
        headers["Authorization"] = f"Bearer {token}"
    releases: list[dict[str, Any]] = json.loads(_get(_API, headers))
    for release in releases:
        tag = str(release.get("tag_name") or "")
        if tag.startswith("smith@") and not release.get("prerelease") and not release.get("draft"):
            return tag[len("smith@") :]
    raise OSError("GitHub lists no smith release")
