namespace MonoSmith.Plugin.Sdk.Tests.Codec;

public sealed class CodecTests
{
    [Fact]
    public void Writes_a_request_as_the_Go_SDK_writes_it()
    {
        Assert.Equal(
            "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"discover\",\"params\":{\"manifests\":[]}}",
            WireMessageMapper.WriteRequest(3, "discover", "{\"manifests\":[]}"));
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"shutdown\"}", WireMessageMapper.WriteRequest(4, "shutdown", null));
    }

    [Fact]
    public void Writes_a_notification_without_an_id()
    {
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"method\":\"cancel\",\"params\":{\"id\":2}}", WireMessageMapper.WriteNotification("cancel", "{\"id\":2}"));
    }

    [Fact]
    public void Writes_a_result_and_an_error_with_the_id_null_when_the_request_could_not_be_read()
    {
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{}}", WireMessageMapper.WriteResult(1, "{}"));
        Assert.Equal(
            "{\"jsonrpc\":\"2.0\",\"id\":null,\"error\":{\"code\":-32700,\"message\":\"bad json\"}}",
            WireMessageMapper.WriteFailure(null, -32700, "bad json"));
    }

    [Fact]
    public void Escapes_the_message_of_an_error_as_Go_does()
    {
        var written = WireMessageMapper.WriteFailure(2, -32000, "a<b");
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":2,\"error\":{\"code\":-32000,\"message\":\"a@Uu003cb\"}}".Replace("@U", "\\"), written);
    }

    [Fact]
    public void Reads_a_request_and_a_notification()
    {
        var request = WireMessageMapper.ParseRequest("{\"jsonrpc\":\"2.0\",\"id\":5,\"method\":\"doctor\",\"params\":{\"projects\":[]}}");
        Assert.Equal(5L, request.Id);
        Assert.Equal("doctor", request.Method);
        Assert.Equal("{\"projects\":[]}", request.Params!.Encode());

        var notification = WireMessageMapper.ParseRequest("{\"jsonrpc\":\"2.0\",\"method\":\"cancel\",\"params\":null}");
        Assert.Null(notification.Id);
        Assert.Null(notification.Params);
        Assert.Null(WireMessageMapper.ParseRequest("{\"id\":null,\"method\":\"x\"}").Id);
        Assert.Equal(string.Empty, WireMessageMapper.ParseRequest("{\"id\":1}").Method);
    }

    [Theory]
    [InlineData("{not json")]
    [InlineData("[]")]
    [InlineData("\"text\"")]
    [InlineData("{\"id\":\"1\",\"method\":\"x\"}")]
    [InlineData("{\"id\":1.5,\"method\":\"x\"}")]
    [InlineData("{\"id\":9223372036854775808,\"method\":\"x\"}")]
    [InlineData("{\"id\":1,\"method\":7}")]
    public void Refuses_a_request_it_cannot_read(string body) =>
        Assert.Throws<MessageParseError>(() => WireMessageMapper.ParseRequest(body));

    [Fact]
    public void Reads_a_result_and_an_error_response()
    {
        var result = WireMessageMapper.ParseResponse("{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"a\":1}}");
        Assert.Equal(1L, result.Id);
        Assert.Equal("{\"a\":1}", result.Result!.Encode());
        Assert.Null(result.Error);

        var failure = WireMessageMapper.ParseResponse("{\"jsonrpc\":\"2.0\",\"id\":null,\"error\":{\"code\":-32601,\"message\":\"nope\"}}");
        Assert.Null(failure.Id);
        Assert.Equal(new ReceivedError(-32601, "nope"), failure.Error);
    }

    [Theory]
    [InlineData("{\"id\":1,\"error\":\"x\"}")]
    [InlineData("{\"id\":1,\"error\":{\"code\":\"1\",\"message\":\"m\"}}")]
    [InlineData("{\"id\":1,\"error\":{\"code\":1}}")]
    public void Refuses_a_response_with_a_malformed_error(string body) =>
        Assert.Throws<MessageParseError>(() => WireMessageMapper.ParseResponse(body));
}
