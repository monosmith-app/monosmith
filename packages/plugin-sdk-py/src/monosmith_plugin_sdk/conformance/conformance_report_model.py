"""Every check run against one plugin."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any

from ..protocol import Peer


@dataclass(frozen=True)
class CheckResult:
    """One check: its name, whether it held, and, when it did not, what is wrong."""

    check: str
    ok: bool
    detail: str | None = None


class ConformanceReport:
    """Every check run against one plugin.

    `to_dict()` gives the shape `smith plugin check --json` prints: `{"plugin": ..., "results": [...]}`.
    """

    def __init__(self, plugin: Peer) -> None:
        self.plugin: Peer = plugin
        self.results: list[CheckResult] = []

    @property
    def ok(self) -> bool:
        """Whether every check passed."""
        return all(result.ok for result in self.results)

    def failures(self) -> str:
        """The failed checks, one per line: `check: detail`."""
        return "\n".join(f"{r.check}: {r.detail or ''}" for r in self.results if not r.ok)

    def add(self, check: str, ok: bool, detail: str) -> None:
        """Record a check. The detail is kept only for a check that failed."""
        self.results.append(CheckResult(check, True) if ok else CheckResult(check, False, detail))

    def to_dict(self) -> dict[str, Any]:
        """The report as plain data."""
        return {
            "plugin": dict(self.plugin),
            "results": [
                {"check": r.check, "ok": r.ok}
                if r.detail is None
                else {"check": r.check, "ok": r.ok, "detail": r.detail}
                for r in self.results
            ],
        }
