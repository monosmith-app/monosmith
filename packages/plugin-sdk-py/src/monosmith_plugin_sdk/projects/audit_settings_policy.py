"""The workspace's audit policy choices."""

from __future__ import annotations

from ..protocol import AuditSettings


def audit_enabled(audit: AuditSettings | None, policy: str) -> bool:
    """Whether the ecosystem's (`node`, `go`, `python`, `rust`) audit policy checks are on."""
    return policy in ((audit or {}).get("policies") or [])


def audit_blocks(audit: AuditSettings | None, policy: str) -> bool:
    """Whether a failing check of the policy fails the audit: it does unless the workspace made the policy only warn."""
    return policy not in ((audit or {}).get("warn") or [])
