using System.Text;
using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// Cuts a byte stream into the bodies of its frames. Bytes arrive in chunks of any size; <see cref="Push"/>
/// returns every body a chunk completed.
/// </summary>
/// <remarks>
/// Headers end at a blank line (CRLF or LF alone, as the Go SDK reads them), <c>Content-Length</c> is matched
/// without regard to case, and a header that is not <c>Name: value</c>, a length that is not a number from 0
/// to 64 MiB, or a header block without a length is a <see cref="FramingError"/>.
/// </remarks>
public sealed class FrameReader
{
    private byte[] _buffer = new byte[4096];
    private int _length;
    private int _bodyStart = -1;
    private int _bodyLength;

    /// <summary>Bytes received that are not part of a complete frame yet.</summary>
    public int Pending => _length;

    /// <summary>Adds the bytes that arrived and returns the body of every frame they completed, in order.</summary>
    /// <param name="chunk">The bytes that arrived.</param>
    /// <returns>The bodies, as text.</returns>
    /// <exception cref="FramingError">The bytes are not frames.</exception>
    public IReadOnlyList<string> Push(ReadOnlySpan<byte> chunk)
    {
        if (_length + chunk.Length > _buffer.Length)
        {
            Array.Resize(ref _buffer, Math.Max(_buffer.Length * 2, _length + chunk.Length));
        }

        chunk.CopyTo(_buffer.AsSpan(_length));
        _length += chunk.Length;

        var bodies = new List<string>();
        var consumed = 0;
        while (true)
        {
            if (_bodyStart < 0 && !ReadHeader(consumed))
            {
                break;
            }

            var end = _bodyStart + _bodyLength;
            if (_length < end)
            {
                break;
            }

            bodies.Add(Framing.Decode(_buffer.AsSpan(_bodyStart, _bodyLength)));
            consumed = end;
            _bodyStart = -1;
        }

        if (consumed > 0)
        {
            Buffer.BlockCopy(_buffer, consumed, _buffer, 0, _length - consumed);
            _length -= consumed;
            if (_bodyStart >= 0)
            {
                _bodyStart -= consumed;
            }
        }

        return bodies;
    }

    /// <summary>Reads the header that starts at <paramref name="from"/>. False while it is incomplete.</summary>
    private bool ReadHeader(int from)
    {
        long length = -1;
        var position = from;
        while (true)
        {
            var newline = Array.IndexOf(_buffer, (byte)'\n', position, _length - position);
            if (newline < 0)
            {
                return false;
            }

            var line = Encoding.Latin1.GetString(_buffer, position, newline - position).TrimEnd('\r', '\n');
            position = newline + 1;
            if (line.Length == 0)
            {
                break;
            }

            var colon = line.IndexOf(':', StringComparison.Ordinal);
            if (colon < 0)
            {
                throw new FramingError($"malformed header \"{line}\"");
            }

            if (!string.Equals(line[..colon].Trim(), "content-length", StringComparison.OrdinalIgnoreCase))
            {
                continue;
            }

            var value = line[(colon + 1)..];
            var text = value.Trim();
            if (text.Length == 0 || !text.All(char.IsAsciiDigit) || !long.TryParse(text, out length) || length > ProtocolInfo.MaxMessageBytes)
            {
                throw new FramingError($"invalid Content-Length \"{value}\"");
            }
        }

        if (length < 0)
        {
            throw new FramingError("message without Content-Length");
        }

        _bodyStart = position;
        _bodyLength = (int)length;
        return true;
    }
}
