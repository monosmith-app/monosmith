using System.Globalization;
using System.Text;

namespace MonoSmith.Plugin.Sdk;

/// <summary><c>Content-Length</c> framing over a byte stream, as the Language Server Protocol frames.</summary>
public static class Framing
{
    private static readonly UTF8Encoding Utf8 = new(encoderShouldEmitUTF8Identifier: false, throwOnInvalidBytes: false);

    /// <summary>
    /// One message as the protocol frames it: <c>Content-Length: &lt;bytes&gt;</c>, a blank line (CRLF CRLF), the body.
    /// The length counts bytes of UTF-8, not characters.
    /// </summary>
    /// <param name="body">The JSON text of the message.</param>
    /// <returns>The bytes to write.</returns>
    public static byte[] Frame(string body)
    {
        ArgumentNullException.ThrowIfNull(body);
        var data = Utf8.GetBytes(body);
        var header = Encoding.ASCII.GetBytes($"Content-Length: {data.Length.ToString(CultureInfo.InvariantCulture)}\r\n\r\n");
        var frame = new byte[header.Length + data.Length];
        header.CopyTo(frame, 0);
        data.CopyTo(frame, header.Length);
        return frame;
    }

    /// <summary>The text of a body: UTF-8, with a replacement character for bytes that are not.</summary>
    /// <param name="bytes">The bytes of a body.</param>
    /// <returns>The text.</returns>
    internal static string Decode(ReadOnlySpan<byte> bytes) => Utf8.GetString(bytes);
}
