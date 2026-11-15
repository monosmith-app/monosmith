"""Checks a downloaded archive against smith's signed checksums.txt (minisign)."""

from __future__ import annotations

import base64
import binascii
import hashlib
import re

from .ed25519_algorithm import verify

RELEASE_PUBLIC_KEY = "RWQB+dlN6CxrAMNDX47n/BQK3mOTgrJ4yVIasbrbP6xRRZTQeB5yncav"
"""The public key smith's releases are signed with (minisign, key ID 6B2CE84DD9F901). Its private
half is the repository secret SMITH_SIGNING_KEY (ADR 0101)."""


def verify_minisign(message: bytes, signature_text: str, public_key_text: str) -> bool:
    """Check a minisign signature (the plain Ed25519 kind smith release writes) of a message.

    The key ID, the signature of the message and the signature of the trusted comment must match.

    Args:
        message: The signed file's bytes (checksums.txt).
        signature_text: The .minisig file's text.
        public_key_text: The public key in base64, as minisign prints it.

    Returns:
        True only when every check passes.
    """
    lines = [line for line in signature_text.splitlines() if line]
    signature_line = next((x for x in lines if not x.startswith(("untrusted comment:", "trusted comment:"))), None)
    trusted = next((x for x in lines if x.startswith("trusted comment: ")), None)
    if not signature_line or not trusted or lines[-1] == signature_line:
        return False
    try:
        sig = base64.b64decode(signature_line, validate=True)
        pub = base64.b64decode(public_key_text.strip(), validate=True)
        global_sig = base64.b64decode(lines[-1], validate=True)
    except binascii.Error:
        return False
    # Ed + key ID (8) + signature (64); Ed + key ID (8) + key (32).
    if len(sig) != 74 or len(pub) != 42 or sig[:2] != b"Ed" or sig[2:10] != pub[2:10]:
        return False
    key, message_signature = pub[10:], sig[10:]
    if not verify(key, message, message_signature):
        return False
    comment = trusted[len("trusted comment: ") :].encode()
    return verify(key, message_signature + comment, global_sig)


def listed_digest(checksums: str, name: str) -> str | None:
    """Return the SHA-256 a checksums.txt lists for a file.

    Args:
        checksums: The text of checksums.txt (``<sha256>  <name>`` lines).
        name: The file name.

    Returns:
        The hex digest, or None when the file is not listed.
    """
    for line in checksums.splitlines():
        parts = re.split(r"\s+\*?", line.strip(), maxsplit=1)
        if len(parts) == 2 and parts[1] == name:
            return parts[0].lower()
    return None


def check_archive(
    archive: bytes, name: str, checksums: bytes, signature: str, public_key: str = RELEASE_PUBLIC_KEY
) -> str | None:
    """Check a downloaded archive against a signed checksums.txt.

    Args:
        archive: The archive's bytes.
        name: The archive's file name.
        checksums: checksums.txt.
        signature: checksums.txt.minisig.
        public_key: The public key; default the release key.

    Returns:
        Why the archive is not trusted, or None when it is.
    """
    if not verify_minisign(checksums, signature, public_key):
        return "the signature of checksums.txt does not match smith's release key"
    listed = listed_digest(checksums.decode(), name)
    if listed is None:
        return f"checksums.txt does not list {name}"
    actual = hashlib.sha256(archive).hexdigest()
    return None if actual == listed else f"{name} does not match its checksum (got {actual}, listed {listed})"
