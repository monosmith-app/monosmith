using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Serve;

public sealed class ServeStreamsTests
{
    [Fact]
    public void Ends_without_an_answer_when_the_host_closes_the_stream()
    {
        using var conversation = new Conversation(new WidgetPlugin());
        conversation.CloseInput();
        Assert.Empty(conversation.RecvAll());
    }

    [Fact]
    public async Task Throws_when_the_header_is_not_a_frame_header()
    {
        var streams = new ServeStreams(new MemoryStream("hello\r\n\r\n"u8.ToArray()), new MemoryStream(), static _ => { });
        var error = await Assert.ThrowsAsync<FramingError>(() => PluginServer.ServeAsync(new WidgetPlugin(), streams));
        Assert.Contains("malformed header", error.Message, StringComparison.Ordinal);
    }

    [Fact]
    public void Ends_when_the_host_is_gone_and_answers_what_is_still_running()
    {
        var lab = new LabPlugin();
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        conversation.Request(2, "discover", "{\"manifests\":[\"stubborn\"]}");
        lab.WaitStarted(1);
        conversation.CloseInput(); // The host is gone, but a plugin that cannot be cancelled finishes what it does.
        lab.Release.Set();
        Assert.Equal(new JsonNumber(2L), ((JsonMap)JsonAny.Parse(conversation.RecvAll()[0]))["id"]);
    }

    [Fact]
    public void Reads_a_request_in_pieces()
    {
        using var conversation = new Conversation(new WidgetPlugin());
        var data = Sdk.Framing.Frame("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersions\":[1]}}");
        for (var start = 0; start < data.Length; start += 7)
        {
            conversation.SendBytes(data[start..Math.Min(start + 7, data.Length)]);
        }

        Assert.Equal(new JsonNumber(1L), ((JsonMap)conversation.RecvJson()["result"])["protocolVersion"]);
    }

    [Fact]
    public void Writes_a_non_ASCII_answer_with_the_length_in_bytes()
    {
        using var conversation = new Conversation(new Lab());
        conversation.Handshake();
        conversation.Request(2, "discover", "{\"manifests\":[\"é-ü\"]}");
        Assert.Contains("é-ü", conversation.Recv(), StringComparison.Ordinal);
    }

    private sealed class Lab : WidgetPlugin, IDiscoverer
    {
        public Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new DiscoverResult { Projects = [new Project { Name = parameters.Manifests[0], Root = "r", Language = "l", Kind = ProjectKind.Lib }] });
    }
}
