"""The error a method raises to say it stopped because the request was cancelled."""

from __future__ import annotations


class RequestCancelledError(Exception):
    """Raised by a method that noticed its request was cancelled (`CancelToken.raise_if_cancelled`).

    The serve loop answers a cancelled request that fails with any error with the code
    `REQUEST_CANCELLED`, so raising this is a convenience, not a requirement.
    """
