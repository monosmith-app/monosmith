using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Serve;

public sealed class ServeCancelTests
{
    private static void Discover(Conversation conversation, long id, string manifest) =>
        conversation.Request(id, "discover", $"{{\"manifests\":[\"{manifest}\"]}}");

    private static long IdOf(JsonMap answer) => ((JsonNumber)answer["id"]).Integer!.Value;

    [Fact]
    public void Sets_the_token_of_a_running_request_and_answers_it_with_the_cancelled_error()
    {
        var lab = new LabPlugin(cancellable: true);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "block-a");
        lab.WaitStarted(1);
        conversation.Notify("cancel", "{\"id\":2}");
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":2,\"error\":{\"code\":-32800,\"message\":\"request cancelled: aborted\"}}", conversation.Recv());
        Discover(conversation, 3, "quick");
        Assert.Contains("\"name\":\"quick\"", conversation.Recv(), StringComparison.Ordinal);
    }

    [Fact]
    public void Ignores_a_cancel_for_an_unknown_or_finished_request_or_with_bad_params()
    {
        var lab = new LabPlugin(cancellable: true);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "quick");
        conversation.Recv();
        conversation.Notify("cancel", "{\"id\":2}");
        conversation.Notify("cancel", "{\"id\":424242}");
        conversation.Notify("cancel", "\"not params\"");
        Discover(conversation, 3, "quick");
        Assert.Equal(3L, IdOf(conversation.RecvJson())); // Nothing else was written: the next frame is this answer.
        Assert.Empty(conversation.Logs);
    }

    [Fact]
    public void Answers_with_the_result_a_request_produced_despite_the_cancel()
    {
        var lab = new LabPlugin(cancellable: true);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "stubborn");
        lab.WaitStarted(1);
        conversation.Notify("cancel", "{\"id\":2}");
        lab.Release.Set();
        Assert.Contains("\"name\":\"stubborn\"", conversation.Recv(), StringComparison.Ordinal);
    }

    [Fact]
    public void Keeps_the_plugin_failure_code_for_an_error_nobody_asked_for()
    {
        using var conversation = new Conversation(new LabPlugin(cancellable: true));
        conversation.Handshake();
        Discover(conversation, 2, "fail-by-itself");
        Assert.Equal("{\"code\":-32000,\"message\":\"the manifest is broken\"}", ((JsonMap)conversation.RecvJson()["error"]).Encode());
    }

    [Fact]
    public void Answers_a_request_that_had_not_started_at_once_and_never_starts_it()
    {
        var lab = new LabPlugin(concurrency: 1, cancellable: true);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "stubborn-running");
        Discover(conversation, 3, "stubborn-waiting");
        lab.WaitStarted(1);
        conversation.Notify("cancel", "{\"id\":3}");
        var waiting = conversation.RecvJson();
        Assert.Equal(3L, IdOf(waiting));
        Assert.Equal("{\"code\":-32800,\"message\":\"request cancelled before it started\"}", ((JsonMap)waiting["error"]).Encode());
        lab.Release.Set();
        Assert.Equal(2L, IdOf(conversation.RecvJson()));
        Assert.Equal(["stubborn-running"], lab.Started);
    }

    [Fact]
    public void Is_ignored_by_a_plugin_that_did_not_say_it_is_cancellable_and_its_token_never_cancels()
    {
        var lab = new LabPlugin();
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "stubborn");
        lab.WaitStarted(1);
        conversation.Notify("cancel", "{\"id\":2}");
        lab.Release.Set();
        Assert.Contains("\"name\":\"stubborn\"", conversation.Recv(), StringComparison.Ordinal);
    }

    [Theory]
    [InlineData(true)]
    [InlineData(false)]
    public void Answers_what_is_in_flight_and_what_is_waiting_with_their_results_before_shutdown(bool cancellable)
    {
        var lab = new LabPlugin(concurrency: 1, cancellable: cancellable);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "stubborn-running");
        Discover(conversation, 3, "stubborn-waiting");
        conversation.Request(4, "shutdown");
        lab.WaitStarted(1);
        lab.Release.Set();
        var answers = Enumerable.Range(0, 3).Select(_ => conversation.RecvJson()).ToList();
        Assert.Equal([2L, 3, 4], answers.Select(IdOf));
        Assert.Contains("\"name\":\"stubborn-running\"", answers[0].Encode(), StringComparison.Ordinal);
        Assert.Contains("\"name\":\"stubborn-waiting\"", answers[1].Encode(), StringComparison.Ordinal);
    }

    [Fact]
    public void Cancels_what_is_in_flight_when_the_input_ends_the_host_being_gone()
    {
        var lab = new LabPlugin(cancellable: true);
        using var conversation = new Conversation(lab);
        conversation.Handshake();
        Discover(conversation, 2, "block-a");
        lab.WaitStarted(1);
        conversation.CloseInput();
        conversation.Finish();
        Assert.Equal(1, lab.Aborted);
    }

    [Fact]
    public async Task Answers_each_of_a_stress_of_concurrent_and_cancelled_requests_exactly_once()
    {
        const int total = 300;
        var lab = new LabPlugin(concurrency: 8, cancellable: true);
        using var conversation = new Conversation(lab);
        conversation.Handshake();

        // The answers are read while the requests are written: a full pipe in either direction would stall both ends.
        var sender = Task.Run(() =>
        {
            for (var id = 2; id < total + 2; id++)
            {
                Discover(conversation, id, id % 3 == 0 ? "block-x" : "quick-x");
                if (id % 6 == 0)
                {
                    conversation.Notify("cancel", $"{{\"id\":{id}}}");
                }
            }

            lab.Release.Set(); // Whatever was not cancelled finishes.
        });
        var counts = new Dictionary<long, int>();
        for (var received = 0; received < total; received++)
        {
            var answer = conversation.RecvJson();
            counts[IdOf(answer)] = counts.GetValueOrDefault(IdOf(answer)) + 1;
            if (answer.ContainsKey("error"))
            {
                Assert.Equal(new JsonNumber(-32800L), ((JsonMap)answer["error"])["code"]);
            }
        }

        Assert.Equal(Enumerable.Range(2, total).Select(id => (long)id).Order(), counts.Keys.Order());
        Assert.All(counts.Values, count => Assert.Equal(1, count));
        await sender;
        Assert.True(lab.Peak <= 8);
    }
}
