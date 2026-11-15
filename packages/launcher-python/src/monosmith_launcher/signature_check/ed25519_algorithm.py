"""Ed25519 signature checking (RFC 8032, section 5.1.7), without a dependency.

Python's standard library has no Ed25519, and the launcher installs with nothing else. It only
checks public signatures, so it holds no secret that its timing could reveal.
"""

from __future__ import annotations

import hashlib

_P = 2**255 - 19
_L = 2**252 + 27742317777372353535851937790883648493
_D = -121665 * pow(121666, _P - 2, _P) % _P
_SQRT_M1 = pow(2, (_P - 1) // 4, _P)

_Point = tuple[int, int, int, int]


def _add(a: _Point, b: _Point) -> _Point:
    x1, y1, z1, t1 = a
    x2, y2, z2, t2 = b
    k = (y1 - x1) * (y2 - x2) % _P
    m = (y1 + x1) * (y2 + x2) % _P
    c = t1 * 2 * _D * t2 % _P
    d = z1 * 2 * z2 % _P
    e, f, g, h = m - k, d - c, d + c, m + k
    return e * f % _P, g * h % _P, f * g % _P, e * h % _P


def _multiply(s: int, point: _Point) -> _Point:
    result: _Point = (0, 1, 1, 0)
    while s > 0:
        if s & 1:
            result = _add(result, point)
        point = _add(point, point)
        s >>= 1
    return result


def _same(a: _Point, b: _Point) -> bool:
    return (a[0] * b[2] - b[0] * a[2]) % _P == 0 and (a[1] * b[2] - b[1] * a[2]) % _P == 0


def _decode(data: bytes) -> _Point | None:
    if len(data) != 32:
        return None
    y = int.from_bytes(data, "little")
    sign = y >> 255
    y &= (1 << 255) - 1
    if y >= _P:
        return None
    x2 = (y * y - 1) * pow(_D * y * y + 1, _P - 2, _P) % _P
    if x2 == 0:
        return None if sign else (0, y, 1, 0)
    x = pow(x2, (_P + 3) // 8, _P)
    if (x * x - x2) % _P != 0:
        x = x * _SQRT_M1 % _P
    if (x * x - x2) % _P != 0:
        return None
    if x & 1 != sign:
        x = _P - x
    return x, y, 1, x * y % _P


_BASE = _decode(bytes.fromhex("5866666666666666666666666666666666666666666666666666666666666666"))


def verify(public_key: bytes, message: bytes, signature: bytes) -> bool:
    """Check an Ed25519 signature.

    Args:
        public_key: The 32 key bytes.
        message: The signed bytes.
        signature: The 64 signature bytes.

    Returns:
        True only when the signature is the key's signature of the message.
    """
    if len(signature) != 64 or _BASE is None:
        return False
    a = _decode(public_key)
    r = _decode(signature[:32])
    s = int.from_bytes(signature[32:], "little")
    if a is None or r is None or s >= _L:
        return False
    h = int.from_bytes(hashlib.sha512(signature[:32] + public_key + message).digest(), "little") % _L
    return _same(_multiply(s, _BASE), _add(r, _multiply(h, a)))
