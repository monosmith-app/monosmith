using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Serve;

public sealed class ServeHandshakeTests
{
    private sealed class Capable : WidgetPlugin, IDiscoverer, IVersioner, IDependencyManager
    {
        public Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new DiscoverResult { Projects = [] });

        public Task<VersionResult> WriteVersionAsync(VersionParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new VersionResult { Files = [] });

        public Task<DependencyResult> InstallAsync(InstallParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new DependencyResult());

        public Task<DependencyResult> UninstallAsync(UninstallParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new DependencyResult());
    }

    private sealed class Meddling : IPlugin
    {
        public PluginDescription Describe(InitializeParams parameters) => new()
        {
            Plugin = new Peer { Name = "w", Version = "1" },
            Kind = PluginKind.Language,
        };
    }

    [Fact]
    public void Answers_initialize_with_the_version_it_chose_and_the_capabilities_derived_from_its_interfaces()
    {
        using var conversation = new Conversation(new Capable());
        conversation.Request(1, "initialize", "{\"protocolVersions\":[1],\"host\":{\"name\":\"t\",\"version\":\"0\"},\"workspaceRoot\":\"/ws\"}");
        Assert.Equal(
            "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"protocolVersion\":1,\"plugin\":{\"name\":\"widget\",\"version\":\"0.1.0\"},"
            + "\"kind\":\"language\",\"capabilities\":{\"discover\":true,\"version\":true,\"dependencies\":true},"
            + "\"manifests\":[\"widget.json\"]}}",
            conversation.Recv());
    }

    [Fact]
    public void Writes_no_capabilities_for_a_plugin_that_implements_no_interface()
    {
        using var conversation = new Conversation(new Meddling());
        var result = conversation.Handshake();
        Assert.Equal("{}", ((JsonMap)result["capabilities"]).Encode());
        Assert.Equal(new JsonNumber(1L), result["protocolVersion"]);
    }

    [Fact]
    public void Advertises_cancel_for_a_plugin_that_says_it_is_cancellable()
    {
        using var conversation = new Conversation(new LabPlugin(cancellable: true));
        var capabilities = Assert.IsType<JsonMap>(conversation.Handshake()["capabilities"]);
        Assert.Equal(new JsonBool(true), capabilities["cancel"]);
    }

    [Fact]
    public void Refuses_a_host_that_does_not_speak_version_1_with_the_Go_SDK_message_and_code()
    {
        using var conversation = new Conversation(new WidgetPlugin());
        conversation.Request(1, "initialize", "{\"protocolVersions\":[2,3],\"host\":{\"name\":\"t\",\"version\":\"0\"},\"workspaceRoot\":\"/\"}");
        Assert.Equal(
            "{\"code\":-32000,\"message\":\"no common protocol version: host speaks [2 3], plugin speaks 1\"}",
            ((JsonMap)conversation.RecvJson()["error"]).Encode());
    }

    [Fact]
    public void Refuses_every_request_before_initialize_shutdown_included_and_goes_on()
    {
        using var conversation = new Conversation(new WidgetPlugin());
        conversation.Request(1, "shutdown", "{}");
        conversation.Request(2, "discover", "{}");
        var first = conversation.RecvJson();
        var second = conversation.RecvJson();
        Assert.Equal("{\"code\":-32600,\"message\":\"initialize must come first\"}", ((JsonMap)first["error"]).Encode());
        Assert.Equal(new JsonNumber(-32600L), ((JsonMap)second["error"])["code"]);
        Assert.Equal(new JsonNumber(1L), conversation.Handshake()["protocolVersion"]);
    }

    [Fact]
    public void Allows_a_second_initialize()
    {
        using var conversation = new Conversation(new WidgetPlugin());
        conversation.Handshake();
        conversation.Request(2, "initialize", "{\"protocolVersions\":[1],\"host\":{\"name\":\"t\",\"version\":\"0\"},\"workspaceRoot\":\"/\"}");
        Assert.Equal(new JsonNumber(1L), ((JsonMap)conversation.RecvJson()["result"])["protocolVersion"]);
    }

    [Fact]
    public void Answers_initialize_with_invalid_params_when_the_host_sends_nonsense()
    {
        using var conversation = new Conversation(new WidgetPlugin());
        conversation.Request(1, "initialize", "{\"protocolVersions\":\"1\"}");
        var error = (JsonMap)conversation.RecvJson()["error"];
        Assert.Equal(new JsonNumber(-32602L), error["code"]);
    }

    [Fact]
    public void Answers_a_plugin_that_describes_itself_with_an_error_as_a_plugin_failure()
    {
        using var conversation = new Conversation(new Broken());
        conversation.Request(1, "initialize", "{\"protocolVersions\":[1]}");
        Assert.Equal("{\"code\":-32000,\"message\":\"cannot start\"}", ((JsonMap)conversation.RecvJson()["error"]).Encode());
    }

    private sealed class Broken : IPlugin
    {
        public PluginDescription Describe(InitializeParams parameters) => throw new InvalidOperationException("cannot start");
    }
}
