"""Cutting a byte stream into frames."""

from __future__ import annotations

import pytest

from monosmith_plugin_sdk import FrameReader, FramingError, frame_message


def test_frame_counts_bytes_not_characters() -> None:
    assert frame_message('{"a":"é€"}') == b'Content-Length: 13\r\n\r\n{"a":"\xc3\xa9\xe2\x82\xac"}'


def test_cuts_frames_that_arrive_in_one_chunk() -> None:
    reader = FrameReader()
    assert reader.push(frame_message('{"a":1}') + frame_message('{"b":2}')) == ['{"a":1}', '{"b":2}']
    assert reader.pending == 0


def test_waits_for_the_rest_when_a_frame_arrives_byte_by_byte_even_inside_a_multi_byte_character() -> None:
    reader = FrameReader()
    bodies: list[str] = []
    for byte in frame_message('{"a":"é€"}'):
        bodies.extend(reader.push(bytes([byte])))
    assert bodies == ['{"a":"é€"}']


def test_matches_content_length_without_regard_to_case_and_accepts_bare_lf_and_other_headers() -> None:
    assert FrameReader().push(b"content-length:  2\nX-Other: yes\n\n{}") == ["{}"]


@pytest.mark.parametrize(
    ("data", "message"),
    [
        (b"hello world\r\n\r\n", 'malformed header "hello world"'),
        (b"X: 1\r\n\r\n", "message without Content-Length"),
        (b"Content-Length: abc\r\n\r\n", "invalid Content-Length"),
        (b"Content-Length: -1\r\n\r\n", "invalid Content-Length"),
        (f"Content-Length: {64 * 1024 * 1024 + 1}\r\n\r\n".encode(), "invalid Content-Length"),
    ],
)
def test_refuses_a_malformed_header_a_missing_or_invalid_length_and_a_length_beyond_64_mib(
    data: bytes, message: str
) -> None:
    with pytest.raises(FramingError, match=message):
        FrameReader().push(data)
