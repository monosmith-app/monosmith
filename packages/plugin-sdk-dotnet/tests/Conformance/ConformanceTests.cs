using MonoSmith.Plugin.Sdk.Tests.Testing;
using MonoSmith.Plugin.Sdk.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Conformance;

public sealed class ConformanceTests
{
    private static readonly HostOptions Options = new() { WorkspaceRoot = "/w" };

    private sealed class Layout : IPlugin, IContributor, IDoctorContributor
    {
        public PluginDescription Describe(InitializeParams parameters) => new()
        {
            Plugin = new Peer { Name = "layout", Version = "0.1.0" },
            Kind = PluginKind.Convention,
            Roles = [Role.Structure],
        };

        public Task<ContributeResult> ContributeAsync(ContributeParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new ContributeResult());

        public Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new DoctorResult { Findings = [] });
    }

    private sealed class Scan : IPlugin, IDoctorContributor
    {
        public bool Cancellable => true;

        public PluginDescription Describe(InitializeParams parameters) => new()
        {
            Plugin = new Peer { Name = "scan", Version = "1.0.0" },
            Kind = PluginKind.Convention,
            Roles = [Role.Structure],
            Concurrency = 4,
        };

        public Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation) =>
            Task.FromResult(new DoctorResult { Findings = [] });
    }

    private sealed class Nameless : IPlugin
    {
        public PluginDescription Describe(InitializeParams parameters) => new()
        {
            Plugin = new Peer { Name = string.Empty, Version = string.Empty },
            Kind = new PluginKind("weird"),
            Roles = [new Role("nonsense")],
        };
    }

    private static List<string> Names(ConformanceReport report) => [.. report.Results.Select(result => result.Check)];

    [Fact]
    public async Task A_plugin_that_keeps_its_promises_passes()
    {
        await using var client = await PluginClient.ConnectInProcessAsync(new Layout(), Options);
        var report = await ConformanceChecks.CheckPluginAsync(client);
        Assert.True(report.Ok, report.Failures());
        var names = Names(report);
        Assert.Contains("names itself", names);
        Assert.Contains("a structure plugin does something", names);
        Assert.Contains("answers 3 requests sent together, one at a time", names);
        Assert.Contains("doctor answers with no projects", names);
    }

    [Fact]
    public async Task A_cancellable_plugin_with_a_concurrency_is_checked_for_both()
    {
        await using var client = await PluginClient.ConnectInProcessAsync(new Scan(), Options);
        var report = await ConformanceChecks.CheckPluginAsync(client);
        Assert.True(report.Ok, report.Failures());
        var names = Names(report);
        Assert.Contains("accepts 4 requests at the same time", names);
        Assert.Contains("answers a cancelled request with its result or the cancelled error", names);
        Assert.Contains("ignores a cancel for a request it does not know", names);
    }

    [Fact]
    public async Task Reports_what_is_wrong_with_a_plugin_that_does_not_name_itself()
    {
        await using var client = await PluginClient.ConnectInProcessAsync(new Nameless(), Options);
        var report = await ConformanceChecks.CheckPluginAsync(client);
        Assert.False(report.Ok);
        Assert.Equal(
            ["names itself", "has a known kind", "plays a known role: nonsense"],
            report.Results.Where(result => !result.Ok).Select(result => result.Check));
        Assert.StartsWith("initialize returned name \"\"", report.Results[0].Detail, StringComparison.Ordinal);
    }

    [Fact]
    public async Task A_language_plugin_needs_to_discover_and_name_manifests()
    {
        await using var client = await PluginClient.ConnectInProcessAsync(new Nameless2(), Options);
        var report = await ConformanceChecks.CheckPluginAsync(client);
        Assert.Contains("a language plugin discovers projects", report.Results.Where(result => !result.Ok).Select(result => result.Check));
    }

    [Fact]
    public async Task Writes_the_report_as_smith_plugin_check_writes_it()
    {
        await using var client = await PluginClient.ConnectInProcessAsync(new Nameless(), Options);
        var report = await ConformanceChecks.CheckPluginAsync(client);
        var json = JsonAny.Parse(report.ToJson());
        var results = Assert.IsType<JsonList>(((JsonMap)json)["results"]);
        Assert.Equal("names itself", ((JsonText)((JsonMap)results[0])["check"]).Value);
        Assert.Equal(new JsonBool(false), ((JsonMap)results[0])["ok"]);
        Assert.StartsWith("{\"plugin\":{\"name\":\"\",\"version\":\"\"},\"results\":[{\"check\":\"names itself\",\"ok\":false,\"detail\":", report.ToJson(), StringComparison.Ordinal);
    }

    private sealed class Nameless2 : WidgetPlugin
    {
        public override PluginDescription Describe(InitializeParams parameters) => new()
        {
            Plugin = new Peer { Name = "widget", Version = "0.1.0" },
            Kind = PluginKind.Language,
        };
    }
}
