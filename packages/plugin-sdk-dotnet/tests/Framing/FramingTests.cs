using System.Text;

namespace MonoSmith.Plugin.Sdk.Tests.Framing;

public sealed class FramingTests
{
    [Fact]
    public void Frames_a_body_with_the_length_in_bytes_not_characters()
    {
        var frame = Encoding.UTF8.GetString(Sdk.Framing.Frame("{\"a\":\"é\"}"));
        Assert.Equal("Content-Length: 10\r\n\r\n{\"a\":\"é\"}", frame);
    }

    [Fact]
    public void Reads_the_bodies_of_frames_that_arrive_whole_or_in_pieces_of_any_size()
    {
        var data = Sdk.Framing.Frame("first").Concat(Sdk.Framing.Frame("sécond")).ToArray();
        foreach (var size in new[] { 1, 2, 7, 13, data.Length })
        {
            var reader = new FrameReader();
            var bodies = new List<string>();
            for (var start = 0; start < data.Length; start += size)
            {
                bodies.AddRange(reader.Push(data.AsSpan(start, Math.Min(size, data.Length - start))));
            }

            Assert.Equal(["first", "sécond"], bodies);
            Assert.Equal(0, reader.Pending);
        }
    }

    [Fact]
    public void Accepts_headers_ended_by_a_bare_line_feed_in_any_case_and_other_headers_before_it()
    {
        var reader = new FrameReader();
        var bodies = reader.Push(Encoding.ASCII.GetBytes("Content-Type: x\ncontent-LENGTH:  2 \n\nhi"));
        Assert.Equal(["hi"], bodies);
    }

    [Fact]
    public void Holds_what_is_not_a_whole_frame_yet()
    {
        var reader = new FrameReader();
        Assert.Empty(reader.Push(Encoding.ASCII.GetBytes("Content-Length: 5\r\n\r\nab")));
        Assert.Equal(23, reader.Pending);
        Assert.Equal(["abcde"], reader.Push(Encoding.ASCII.GetBytes("cde")));
    }

    [Fact]
    public void Reads_an_empty_body_and_replaces_bytes_that_are_not_UTF8()
    {
        var reader = new FrameReader();
        Assert.Equal([string.Empty], reader.Push(Encoding.ASCII.GetBytes("Content-Length: 0\r\n\r\n")));
        var bodies = reader.Push([.. Encoding.ASCII.GetBytes("Content-Length: 1\r\n\r\n"), 0xFF]);
        Assert.Equal([char.ConvertFromUtf32(0xFFFD)], bodies);
    }

    [Theory]
    [InlineData("hello\r\n\r\n", "malformed header")]
    [InlineData("X: y\r\n\r\n", "message without Content-Length")]
    [InlineData("Content-Length: abc\r\n\r\n", "invalid Content-Length")]
    [InlineData("Content-Length: -1\r\n\r\n", "invalid Content-Length")]
    [InlineData("Content-Length: +1\r\n\r\n", "invalid Content-Length")]
    [InlineData("Content-Length: \r\n\r\n", "invalid Content-Length")]
    [InlineData("Content-Length: 67108865\r\n\r\n", "invalid Content-Length")]
    [InlineData("Content-Length: 99999999999999999999\r\n\r\n", "invalid Content-Length")]
    public void Refuses_a_stream_that_is_not_made_of_frames(string stream, string message)
    {
        var error = Assert.Throws<FramingError>(() => new FrameReader().Push(Encoding.ASCII.GetBytes(stream)));
        Assert.Contains(message, error.Message, StringComparison.Ordinal);
    }

    [Fact]
    public void Accepts_the_largest_length_the_protocol_allows()
    {
        var reader = new FrameReader();
        Assert.Empty(reader.Push(Encoding.ASCII.GetBytes("Content-Length: 67108864\r\n\r\n")));
    }
}
