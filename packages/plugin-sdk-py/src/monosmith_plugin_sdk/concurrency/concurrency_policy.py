"""What a declared `concurrency` means."""

from __future__ import annotations

CONCURRENCY_UNLIMITED = -1
"""The `concurrency` value that leaves the number of simultaneous requests to the host."""

UNLIMITED_IN_FLIGHT = 32
"""How many requests the host keeps in flight to a plugin that declared unlimited concurrency."""

UNLIMITED_WORKERS = 1024
"""The most workers a plugin that declared unlimited concurrency runs: the pool stays bounded."""


def is_valid_concurrency(declared: int) -> bool:
    """Whether a declared concurrency is one the protocol defines.

    Absent (0), a count (1 and above) or unlimited (-1).
    """
    return declared >= CONCURRENCY_UNLIMITED


def concurrency_limit(declared: int | None) -> tuple[int, bool]:
    """Read a declared concurrency: `(limit, unlimited)`.

    An absent, zero or invalid value is 1, the default of every plugin that says nothing.
    """
    if declared == CONCURRENCY_UNLIMITED:
        return 0, True
    if declared is None or isinstance(declared, bool) or declared < 1:
        return 1, False
    return declared, False


def in_flight_limit(declared: int | None) -> int:
    """How many requests the host keeps in flight to a plugin that declared `declared`."""
    limit, unlimited = concurrency_limit(declared)
    return UNLIMITED_IN_FLIGHT if unlimited else limit


def worker_limit(declared: int | None) -> int:
    """How many requests a plugin that declared `declared` works on at once (its pool size)."""
    limit, unlimited = concurrency_limit(declared)
    return UNLIMITED_WORKERS if unlimited else limit
