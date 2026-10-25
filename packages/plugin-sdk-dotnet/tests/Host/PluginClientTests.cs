using MonoSmith.Plugin.Sdk.Tests.Testing;
using MonoSmith.Plugin.Sdk.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Host;

public sealed class PluginClientTests
{
    private static readonly HostOptions Options = new() { WorkspaceRoot = "/w", CancelGrace = TimeSpan.FromSeconds(10) };

    private static DiscoverParams Manifest(string name) => new() { Manifests = [name] };

    [Fact]
    public async Task Connects_in_process_with_the_real_handshake_and_calls_a_typed_method()
    {
        await using var client = await PluginClient.ConnectInProcessAsync(new LabPlugin(), Options);
        Assert.Equal("lab", client.Info.Plugin.Name);
        Assert.True(client.Supports("discover"));
        Assert.False(client.Supports("publish"));
        Assert.True(client.Supports("initialize"));
        var result = await client.DiscoverAsync(Manifest("quick"));
        Assert.Equal("quick", result.Projects[0].Name);
    }

    [Fact]
    public async Task Throws_the_error_the_plugin_answered_with_its_code()
    {
        await using var client = await PluginClient.ConnectInProcessAsync(new LabPlugin(), Options);
        var failure = await Assert.ThrowsAsync<RpcError>(() => client.DiscoverAsync(Manifest("fail")));
        Assert.Equal(ErrorCodes.PluginFailure, failure.Code);
        Assert.Equal("the manifest is broken", failure.Message);

        var unknown = await Assert.ThrowsAsync<RpcError>(() => client.CallRawAsync("nope", "{}"));
        Assert.Equal(ErrorCodes.MethodNotFound, unknown.Code);
    }

    [Fact]
    public async Task Never_has_more_requests_in_flight_than_the_plugin_declared()
    {
        var lab = new LabPlugin();
        await using var client = await PluginClient.ConnectInProcessAsync(lab, Options);
        var calls = Enumerable.Range(0, 5).Select(i => client.DiscoverAsync(Manifest($"quick-{i}"))).ToList();
        await Task.WhenAll(calls);
        Assert.Equal(1, client.PeakInFlight);
        Assert.Equal(1, lab.Peak);
    }

    [Fact]
    public async Task Sends_as_many_requests_together_as_the_plugin_declared()
    {
        var lab = new LabPlugin(concurrency: 3);
        await using var client = await PluginClient.ConnectInProcessAsync(lab, Options);
        var calls = Enumerable.Range(0, 6).Select(i => client.DiscoverAsync(Manifest($"block-{i}"))).ToList();
        lab.WaitStarted(3);
        Assert.Equal(3, lab.Started.Count);
        lab.Release.Set();
        await Task.WhenAll(calls);
        Assert.Equal(3, client.PeakInFlight);
    }

    [Fact]
    public async Task Gives_up_on_a_call_whose_time_ran_out_and_kills_a_plugin_that_cannot_be_asked_to_stop()
    {
        var lab = new LabPlugin();
        var client = await PluginClient.ConnectInProcessAsync(lab, Options);
        var interrupted = await Assert.ThrowsAsync<CallInterruptedError>(
            () => client.DiscoverAsync(Manifest("stubborn"), timeout: TimeSpan.FromMilliseconds(50)));
        Assert.Equal(InterruptionReason.Timeout, interrupted.Reason);
        Assert.True(interrupted.Killed);
        lab.Release.Set();
        _ = await Record.ExceptionAsync(client.CloseAsync);
    }

    [Fact]
    public async Task Asks_a_cancellable_plugin_to_stop_and_reports_what_it_answered()
    {
        var lab = new LabPlugin(cancellable: true);
        await using var client = await PluginClient.ConnectInProcessAsync(lab, Options);
        Assert.True(client.Supports("cancel"));
        using var source = new CancellationTokenSource();
        var call = client.DiscoverAsync(Manifest("block-a"), source.Token);
        lab.WaitStarted(1);
        await source.CancelAsync();
        var interrupted = await Assert.ThrowsAsync<CallInterruptedError>(() => call);
        Assert.Equal(InterruptionReason.Cancelled, interrupted.Reason);
        Assert.False(interrupted.Killed);
        Assert.Equal(ErrorCodes.RequestCancelled, interrupted.Response!.Code);
        Assert.Equal("a", (await client.DiscoverAsync(Manifest("a"))).Projects[0].Name);
    }

    [Fact]
    public async Task Returns_the_result_of_a_request_that_finished_despite_the_cancel()
    {
        var lab = new LabPlugin(cancellable: true);
        await using var client = await PluginClient.ConnectInProcessAsync(lab, Options);
        using var source = new CancellationTokenSource();
        var call = client.DiscoverAsync(Manifest("stubborn"), source.Token);
        lab.WaitStarted(1);
        await source.CancelAsync();
        lab.Release.Set();
        Assert.Equal("stubborn", (await call).Projects[0].Name);
    }

    [Fact]
    public async Task Does_not_send_a_call_whose_token_was_cancelled_before_it_had_a_place()
    {
        var lab = new LabPlugin(concurrency: 1);
        await using var client = await PluginClient.ConnectInProcessAsync(lab, Options);
        var running = client.DiscoverAsync(Manifest("stubborn"));
        lab.WaitStarted(1);
        using var source = new CancellationTokenSource();
        var waiting = client.DiscoverAsync(Manifest("never"), source.Token);
        await source.CancelAsync();
        var interrupted = await Assert.ThrowsAsync<CallInterruptedError>(() => waiting);
        Assert.Contains("cancelled before it was sent", interrupted.Message, StringComparison.Ordinal);
        lab.Release.Set();
        await running;
        Assert.Equal(["stubborn"], lab.Started);
    }

    [Fact]
    public async Task Refuses_a_plugin_that_answers_initialize_with_an_error()
    {
        var error = await Assert.ThrowsAsync<InvalidOperationException>(() => PluginClient.ConnectInProcessAsync(new RefusingPlugin(), Options));
        Assert.Contains("initializing plugin", error.Message, StringComparison.Ordinal);
        Assert.Contains("cannot start", error.Message, StringComparison.Ordinal);
    }

    [Fact]
    public async Task Tells_the_host_name_and_the_workspace_root_to_the_plugin()
    {
        var recording = new RecordingPlugin();
        await using var client = await PluginClient.ConnectInProcessAsync(recording, new HostOptions { WorkspaceRoot = "/ws", Host = new Peer { Name = "t", Version = "9" } });
        Assert.Equal("t 9 /ws", recording.Seen);
    }

    [Fact]
    public async Task Fails_a_call_after_the_plugin_is_gone()
    {
        var client = await PluginClient.ConnectInProcessAsync(new LabPlugin(), Options);
        await client.CloseAsync();
        await Assert.ThrowsAsync<InvalidOperationException>(() => client.DiscoverAsync(Manifest("quick")));
    }

    private sealed class RefusingPlugin : IPlugin
    {
        public PluginDescription Describe(InitializeParams parameters) => throw new InvalidOperationException("cannot start");
    }

    private sealed class RecordingPlugin : WidgetPlugin
    {
        public string Seen { get; private set; } = string.Empty;

        public override PluginDescription Describe(InitializeParams parameters)
        {
            Seen = $"{parameters.Host.Name} {parameters.Host.Version} {parameters.WorkspaceRoot}";
            return base.Describe(parameters);
        }
    }
}
