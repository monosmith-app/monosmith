using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;
using MonoSmith.Plugin.Sdk.Testing;
using StructureScan;
using Xunit;

namespace StructureScan.Tests;

// Nothing here depends on how long a scan takes: the scans wait on a latch the test controls, or are cancelled.
public sealed class ScanPluginTests : IDisposable
{
    private readonly string _workspace = Directory.CreateTempSubdirectory("scan-").FullName;

    public void Dispose() => Directory.Delete(_workspace, recursive: true);

    private Project Core(Dictionary<string, int> files)
    {
        foreach (var (name, size) in files)
        {
            var path = Path.Combine(_workspace, "libs", "core", name.Replace('/', Path.DirectorySeparatorChar));
            Directory.CreateDirectory(Path.GetDirectoryName(path)!);
            File.WriteAllBytes(path, new byte[size]);
        }

        return new Project { Name = "core", Root = "libs/core", Language = "widget", Kind = ProjectKind.Lib };
    }

    private Task<PluginClient> Connect(IPlugin plugin) =>
        PluginClient.ConnectInProcessAsync(plugin, new HostOptions { WorkspaceRoot = _workspace });

    // A scan that tells the test when it has started its walk.
    private sealed class SlowScan : IPlugin, IDoctorContributor
    {
        private readonly ScanPlugin _inner = new(TimeSpan.FromSeconds(60)); // A minute a file: only a cancel ends it.

        public SemaphoreSlim Started { get; } = new(0);

        public bool Cancellable => _inner.Cancellable;

        public PluginDescription Describe(InitializeParams parameters) => _inner.Describe(parameters);

        public Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation)
        {
            Started.Release();
            return _inner.DoctorAsync(parameters, cancellation);
        }
    }

    [Fact]
    public async Task The_plugin_conforms()
    {
        await using var client = await Connect(new ScanPlugin(TimeSpan.Zero));
        var report = await ConformanceChecks.CheckPluginAsync(client);
        Assert.True(report.Ok, report.Failures());
    }

    [Fact]
    public async Task It_advertises_cancel_and_a_concurrency_of_four()
    {
        await using var client = await Connect(new ScanPlugin(TimeSpan.Zero));
        Assert.True(client.Info.Capabilities.Cancel);
        Assert.Equal(4, client.Info.Concurrency);
    }

    [Fact]
    public async Task Doctor_warns_about_a_big_file()
    {
        var project = Core(new Dictionary<string, int> { ["src/small.txt"] = 10, ["data/huge.bin"] = (1 << 20) + 1 });
        await using var client = await Connect(new ScanPlugin(TimeSpan.Zero));
        var finding = (await client.DoctorAsync(new DoctorParams { Projects = [project] })).Findings[0];
        Assert.False(finding.Ok);
        Assert.True(finding.Warning);
        Assert.Equal("data/huge.bin", finding.Detail);
    }

    [Fact]
    public async Task A_cancelled_scan_stops_midway_and_the_answer_is_the_cancelled_error()
    {
        var project = Core("abcdefgh".ToDictionary(letter => $"{letter}.txt", _ => 1));
        var plugin = new SlowScan();
        await using var client = await Connect(plugin);
        using var source = new CancellationTokenSource();
        var scan = client.DoctorAsync(new DoctorParams { Projects = [project] }, source.Token);
        Assert.True(await plugin.Started.WaitAsync(TimeSpan.FromSeconds(30)));
        await source.CancelAsync();
        var error = await Assert.ThrowsAsync<CallInterruptedError>(() => scan);
        Assert.Equal((InterruptionReason.Cancelled, false), (error.Reason, error.Killed));
        Assert.Equal(ErrorCodes.RequestCancelled, error.Response!.Code);
    }

    [Fact]
    public async Task Four_scans_run_at_the_same_time()
    {
        var project = Core(new Dictionary<string, int> { ["a.txt"] = 1, ["b.txt"] = 1 });
        var plugin = new SlowScan();
        await using var client = await Connect(plugin);
        using var source = new CancellationTokenSource();
        var scans = Enumerable.Range(0, 4).Select(_ => IgnoreInterruption(client, project, source.Token)).ToList();
        for (var started = 0; started < 4; started++)
        {
            // All four reached doctor before any of them could finish.
            Assert.True(await plugin.Started.WaitAsync(TimeSpan.FromSeconds(30)));
        }

        await source.CancelAsync();
        await Task.WhenAll(scans);
        Assert.Equal(4, client.PeakInFlight);
    }

    private static async Task IgnoreInterruption(PluginClient client, Project project, CancellationToken cancellation)
    {
        try
        {
            await client.DoctorAsync(new DoctorParams { Projects = [project] }, cancellation);
        }
        catch (CallInterruptedError)
        {
            // The test cancels the scans once it has seen them all running.
        }
    }
}
