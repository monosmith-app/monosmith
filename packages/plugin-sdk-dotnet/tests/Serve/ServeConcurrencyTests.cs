using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Serve;

public sealed class ServeConcurrencyTests
{
    private static void Discover(Conversation conversation, long id, string manifest) =>
        conversation.Request(id, "discover", $"{{\"manifests\":[\"{manifest}\"]}}");

    private static long IdOf(JsonMap answer) => ((JsonNumber)answer["id"]).Integer!.Value;

    [Fact]
    public void Handles_one_request_at_a_time_in_order_for_a_plugin_that_declares_nothing()
    {
        var lab = new LabPlugin();
        using var conversation = new Conversation(lab);
        var result = conversation.Handshake();
        Assert.False(result.ContainsKey("concurrency"));
        Assert.False(((JsonMap)result["capabilities"]).ContainsKey("cancel"));
        for (var id = 2; id < 8; id++)
        {
            Discover(conversation, id, $"quick-{id}");
        }

        Assert.Equal([2L, 3, 4, 5, 6, 7], Enumerable.Range(0, 6).Select(_ => IdOf(conversation.RecvJson())));
        Assert.Equal(1, lab.Peak);
    }

    [Fact]
    public void Runs_as_many_handlers_together_as_the_plugin_declared_and_no_more()
    {
        var lab = new LabPlugin(concurrency: 3);
        using var conversation = new Conversation(lab);
        Assert.Equal(new JsonNumber(3L), conversation.Handshake()["concurrency"]);
        for (var id = 2; id < 10; id++)
        {
            Discover(conversation, id, $"block-{id}");
        }

        lab.WaitStarted(3);
        Assert.Equal(3, lab.Peak);
        Assert.Equal(3, lab.Started.Count); // The other five wait for a place; the pool never starts a fourth.
        lab.Release.Set();
        var answered = Enumerable.Range(0, 8).Select(_ => IdOf(conversation.RecvJson())).ToHashSet();
        Assert.Equal(Enumerable.Range(2, 8).Select(id => (long)id).ToHashSet(), answered);
        Assert.Equal(3, lab.Peak);
    }

    [Fact]
    public void May_answer_out_of_order_each_answer_carrying_the_id_of_its_request()
    {
        var lab = new LabPlugin(concurrency: 2);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "stubborn-first");
        lab.WaitStarted(1);
        Discover(conversation, 3, "quick");
        var answer = conversation.RecvJson();
        Assert.Equal(3L, IdOf(answer));
        Assert.Contains("\"name\":\"quick\"", answer.Encode(), StringComparison.Ordinal);
        lab.Release.Set();
        Assert.Equal(2L, IdOf(conversation.RecvJson()));
    }

    [Fact]
    public void Refuses_a_request_id_that_is_already_in_flight_and_a_host_that_floods_it()
    {
        var lab = new LabPlugin(concurrency: 1);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "stubborn-a");
        lab.WaitStarted(1);
        Discover(conversation, 2, "stubborn-again");
        Assert.Equal(new JsonNumber(-32600L), ((JsonMap)conversation.RecvJson()["error"])["code"]);
        for (var id = 3; id < 3 + 1024 + 5; id++)
        {
            Discover(conversation, id, "stubborn-queued");
        }

        for (var refused = 0; refused < 5; refused++)
        {
            var error = (JsonMap)conversation.RecvJson()["error"];
            Assert.Equal(new JsonNumber(-32600L), error["code"]);
            Assert.Contains("too many requests", ((JsonText)error["message"]).Value, StringComparison.Ordinal);
        }

        lab.Release.Set();
        Assert.Equal(1 + 1024, Enumerable.Range(0, 1 + 1024).Select(_ => IdOf(conversation.RecvJson())).ToHashSet().Count);
    }
}
