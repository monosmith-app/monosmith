using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Serve;

public sealed class ServeDispatchTests
{
    private sealed class Dispatching : WidgetPlugin, IDiscoverer, IDoctorContributor, IPublisher, IRemover
    {
        private int _count;

        public Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new DiscoverResult
            {
                Projects = [.. parameters.Manifests.Select(manifest => new Project { Name = manifest, Root = manifest, Language = "widget", Kind = ProjectKind.Lib })],
            });

        public Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation)
        {
            if (parameters.Projects.Count == 0)
            {
                throw new InvalidOperationException("the toolchain is missing: install dotnet 8");
            }

            var call = Interlocked.Increment(ref _count);
            return Task.FromResult(new DoctorResult { Findings = [new Finding { Check = $"call {call}", Ok = true }] });
        }

        public Task<PublishResult> PublishAsync(PublishParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new PublishResult { Published = true });

        public Task<RemoveResult> RemoveAsync(RemoveParams parameters, CancellationToken cancellation) =>
            Task.FromResult<RemoveResult>(null!);
    }

    private static Conversation Started(IPlugin plugin)
    {
        var conversation = new Conversation(plugin);
        conversation.Handshake();
        return conversation;
    }

    [Fact]
    public void Calls_the_method_with_the_params_shaped_as_the_types_promise()
    {
        using var conversation = Started(new Dispatching());
        conversation.Request(2, "discover", "{\"manifests\":null}");
        Assert.Equal("{\"projects\":[]}", ((JsonMap)conversation.RecvJson()["result"]).Encode());
        conversation.Request(3, "discover", "{\"manifests\":[\"a\"]}");
        Assert.Equal(
            "{\"jsonrpc\":\"2.0\",\"id\":3,\"result\":{\"projects\":[{\"name\":\"a\",\"root\":\"a\",\"language\":\"widget\",\"kind\":\"lib\"}]}}",
            conversation.Recv());
        conversation.Request(4, "publish", "{\"project\":{\"name\":\"a\",\"root\":\"a\",\"language\":\"w\",\"kind\":\"lib\"},\"version\":\"1.0.0\"}");
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":4,\"result\":{\"published\":true}}", conversation.Recv());
    }

    [Fact]
    public void Answers_a_method_the_plugin_lacks_with_method_not_found_whether_or_not_the_protocol_knows_it()
    {
        using var conversation = Started(new Dispatching());
        conversation.Request(2, "contribute", "{}");
        Assert.Equal("{\"code\":-32601,\"message\":\"method not supported: contribute\"}", ((JsonMap)conversation.RecvJson()["error"]).Encode());
        conversation.Request(3, "nope", "{}");
        Assert.Equal("{\"code\":-32601,\"message\":\"method not supported: nope\"}", ((JsonMap)conversation.RecvJson()["error"]).Encode());
    }

    [Fact]
    public void Turns_what_a_method_throws_into_a_plugin_failure_with_its_message()
    {
        using var conversation = Started(new Dispatching());
        conversation.Request(2, "doctor", "{\"projects\":[]}");
        Assert.Equal(
            "{\"code\":-32000,\"message\":\"the toolchain is missing: install dotnet 8\"}",
            ((JsonMap)conversation.RecvJson()["error"]).Encode());
    }

    [Fact]
    public void Answers_params_of_the_wrong_shape_with_invalid_params_naming_the_member()
    {
        using var conversation = Started(new Dispatching());
        conversation.Request(2, "discover", "{\"manifests\":\"a\"}");
        Assert.Equal(
            "{\"code\":-32602,\"message\":\"params.manifests must be an array, not a string\"}",
            ((JsonMap)conversation.RecvJson()["error"]).Encode());
    }

    [Fact]
    public void Writes_an_empty_object_for_a_method_that_returned_nothing()
    {
        using var conversation = Started(new Dispatching());
        conversation.Request(2, "remove", "{\"project\":{\"name\":\"a\",\"root\":\"a\",\"language\":\"w\",\"kind\":\"lib\"}}");
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{}}", conversation.Recv());
    }

    [Fact]
    public void Answers_shutdown_with_an_empty_object_and_stops_reading()
    {
        using var conversation = Started(new WidgetPlugin());
        conversation.Request(2, "shutdown", "{}");
        conversation.Request(3, "discover", "{}");
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{}}", conversation.Recv());
        Assert.Empty(conversation.RecvAll());
    }

    [Fact]
    public void Answers_no_notification_and_an_unreadable_body_with_parse_error_and_an_id_of_null()
    {
        using var conversation = Started(new WidgetPlugin());
        conversation.Notify("discover", "{}");
        conversation.SendRaw("{not json");
        conversation.SendRaw("[]");
        conversation.Request(4, "shutdown");
        var answers = conversation.RecvAll().Select(body => (JsonMap)JsonAny.Parse(body)).ToList();
        Assert.Equal([JsonNull.Instance, JsonNull.Instance, new JsonNumber(4L)], answers.Select(answer => answer["id"]));
        Assert.Equal(new JsonNumber(-32700L), ((JsonMap)answers[0]["error"])["code"]);
        Assert.Equal(new JsonNumber(-32700L), ((JsonMap)answers[1]["error"])["code"]);
    }

    [Fact]
    public void Calls_a_method_on_the_plugin_object_so_it_can_keep_state()
    {
        using var conversation = Started(new Dispatching());
        conversation.Request(2, "doctor", "{\"projects\":[{\"name\":\"a\",\"root\":\"a\",\"language\":\"w\",\"kind\":\"lib\"}]}");
        conversation.Request(3, "doctor", "{\"projects\":[{\"name\":\"a\",\"root\":\"a\",\"language\":\"w\",\"kind\":\"lib\"}]}");
        Assert.Contains("\"check\":\"call 1\"", conversation.Recv(), StringComparison.Ordinal);
        Assert.Contains("\"check\":\"call 2\"", conversation.Recv(), StringComparison.Ordinal);
    }
}
