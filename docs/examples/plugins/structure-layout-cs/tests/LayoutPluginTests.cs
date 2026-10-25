using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;
using MonoSmith.Plugin.Sdk.Testing;
using StructureLayout;
using Xunit;

namespace StructureLayout.Tests;

public sealed class LayoutPluginTests : IDisposable
{
    private static readonly Project Core = new() { Name = "core", Root = "libs/core", Language = "widget", Kind = ProjectKind.Lib };

    private readonly string _workspace = Directory.CreateTempSubdirectory("layout-").FullName;

    public void Dispose() => Directory.Delete(_workspace, recursive: true);

    private Task<PluginClient> Connect() => PluginClient.ConnectInProcessAsync(new LayoutPlugin(), new HostOptions { WorkspaceRoot = _workspace });

    [Fact]
    public async Task The_plugin_conforms()
    {
        await using var client = await Connect();
        var report = await ConformanceChecks.CheckPluginAsync(client);
        Assert.True(report.Ok, report.Failures());
    }

    [Fact]
    public async Task The_handshake_advertises_exactly_what_is_implemented()
    {
        await using var client = await Connect();
        Assert.Equal([Role.Structure], client.Info.Roles);
        Assert.True(client.Info.Capabilities.Contribute);
        Assert.True(client.Info.Capabilities.Doctor);
        Assert.False(client.Info.Capabilities.Discover);
        Assert.False(client.Info.Capabilities.Cancel);
    }

    [Fact]
    public async Task Lint_waits_for_the_layout_check()
    {
        await using var client = await Connect();
        var core = (await client.ContributeAsync(new ContributeParams { Projects = [Core] })).Projects["core"];
        Assert.Contains("layout-check", core.Tasks.Keys);
        Assert.Equal(["layout-check"], core.DependsOn["lint"]);
    }

    [Fact]
    public async Task A_method_the_plugin_lacks_is_32601()
    {
        await using var client = await Connect();
        var error = await Assert.ThrowsAsync<RpcError>(() => client.PublishAsync(new PublishParams { Project = Core, Version = "1.0.0" }));
        Assert.Equal(-32601, error.Code);
    }

    [Fact]
    public async Task Doctor_warns_about_a_missing_readme()
    {
        Directory.CreateDirectory(Path.Combine(_workspace, "libs", "core"));
        await using var client = await Connect();
        var finding = (await client.DoctorAsync(new DoctorParams { Projects = [Core] })).Findings[0];
        Assert.False(finding.Ok);
        Assert.True(finding.Warning);
        Assert.NotEmpty(finding.Remedy);
    }

    [Fact]
    public void Find_junk_drawers_skips_build_output()
    {
        foreach (var folder in new[] { "src/orders", "src/Utils", "dist/helpers", "common" })
        {
            Directory.CreateDirectory(Path.Combine(_workspace, folder));
        }

        Assert.Equal(["common", "src/Utils"], LayoutPlugin.FindJunkDrawers(_workspace).Order(StringComparer.Ordinal));
    }
}
