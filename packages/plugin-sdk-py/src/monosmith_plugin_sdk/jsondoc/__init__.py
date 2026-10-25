"""JSON editing that keeps key order, and the text of what you do not touch."""

from .json_document_error import JsonDocumentError
from .json_document_model import JsonDocument

__all__ = ["JsonDocument", "JsonDocumentError"]
