"""What a declared `concurrency` means: one at a time by default, a count, or unlimited."""

from .concurrency_policy import (
    CONCURRENCY_UNLIMITED,
    UNLIMITED_IN_FLIGHT,
    UNLIMITED_WORKERS,
    concurrency_limit,
    in_flight_limit,
    is_valid_concurrency,
    worker_limit,
)

__all__ = [
    "CONCURRENCY_UNLIMITED",
    "UNLIMITED_IN_FLIGHT",
    "UNLIMITED_WORKERS",
    "concurrency_limit",
    "in_flight_limit",
    "is_valid_concurrency",
    "worker_limit",
]
