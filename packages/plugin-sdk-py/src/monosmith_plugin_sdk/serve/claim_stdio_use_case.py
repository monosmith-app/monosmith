"""Taking the process's standard input and output for the protocol and nothing else."""

from __future__ import annotations

import os
import sys

from .protocol_streams_contract import ServeStreams


def claim_stdio() -> ServeStreams:
    """Take standard input and output for the protocol, and leave nothing else a way to them.

    A `print`, a library's write to `sys.stdout` or a child process that inherits the
    plugin's standard output would corrupt the stream the host reads. So the protocol gets
    private copies of file descriptors 0 and 1, and from here on:

    - file descriptor 1, and `sys.stdout`, are standard error: whatever else writes to
      standard output (a `print`, a C extension, a child process) reaches the host's log, not
      the stream;
    - file descriptor 0 is the null device: a child process that reads standard input does not
      take the host's requests.

    The returned streams are the only way to the real ones.
    """
    for stream in (sys.stdout, sys.stderr):
        if stream is not None:
            stream.flush()
    protocol_in = os.fdopen(os.dup(0), "rb", buffering=0)
    protocol_out = os.fdopen(os.dup(1), "wb", buffering=0)
    os.dup2(2, 1)
    null = os.open(os.devnull, os.O_RDONLY)
    os.dup2(null, 0)
    os.close(null)
    sys.stdout = sys.stderr
    return ServeStreams(input=protocol_in, output=protocol_out)
