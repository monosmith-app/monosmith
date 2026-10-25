"""What answering one request came to."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class RequestOutcome:
    """The result already written as JSON, or an error object."""

    result_json: str | None = None
    code: int = 0
    message: str = ""

    @property
    def ok(self) -> bool:
        """Whether the request has a result."""
        return self.result_json is not None


def succeeded(result_json: str) -> RequestOutcome:
    """An outcome with a result."""
    return RequestOutcome(result_json=result_json)


def failed(code: int, message: str) -> RequestOutcome:
    """An outcome with an error."""
    return RequestOutcome(code=code, message=message)
