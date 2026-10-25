"""Cooperative cancellation: the token a method watches, and the error it raises when it stops."""

from .cancel_token_model import CancelToken
from .request_cancelled_error import RequestCancelledError

__all__ = ["CancelToken", "RequestCancelledError"]
