using LanguageWidget;
using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;
using MonoSmith.Plugin.Sdk.Testing;
using Xunit;

namespace LanguageWidget.Tests;

// The plugin is served in this process over in-memory pipes, with the real handshake and framing: the way smith talks to it.
public sealed class WidgetPluginTests : IDisposable
{
    private static readonly Project Core = new() { Name = "core", Root = "libs/core", Language = "widget", Kind = ProjectKind.Lib };

    private readonly string _workspace = Directory.CreateTempSubdirectory("widget-").FullName;

    public void Dispose() => Directory.Delete(_workspace, recursive: true);

    private void Write(string relative, string content)
    {
        var path = Path.Combine(_workspace, relative.Replace('/', Path.DirectorySeparatorChar));
        Directory.CreateDirectory(Path.GetDirectoryName(path)!);
        File.WriteAllText(path, content);
    }

    private string Read(string relative) => File.ReadAllText(Path.Combine(_workspace, relative.Replace('/', Path.DirectorySeparatorChar)));

    private Task<PluginClient> Connect() => PluginClient.ConnectInProcessAsync(new WidgetPlugin(), new HostOptions { WorkspaceRoot = _workspace });

    [Fact]
    public async Task The_plugin_conforms_generating_an_app_a_lib_and_a_package_that_it_discovers_again()
    {
        await using var client = await Connect();
        var report = await ConformanceChecks.CheckPluginAsync(client, _workspace);
        Assert.True(report.Ok, report.Failures());
    }

    [Fact]
    public async Task Discover_turns_uses_into_imports()
    {
        Write("libs/core/widget.json", "{\"name\":\"core\",\"version\":\"1.0.0\"}");
        Write("apps/shop/widget.json", "{\"name\":\"shop\",\"version\":\"1.0.0\",\"kind\":\"app\",\"uses\":[\"core\"]}");
        await using var client = await Connect();
        var result = await client.DiscoverAsync(new DiscoverParams { Manifests = ["libs/core/widget.json", "apps/shop/widget.json"] });
        var (core, shop) = (result.Projects[0], result.Projects[1]);
        Assert.Equal((ProjectKind.Lib, ProjectKind.App), (core.Kind, shop.Kind));
        Assert.Equal(["core"], shop.Imports);
        Assert.Equal(["core"], core.Exports);
        Assert.Equal(["../../dist/libs/core"], core.Tasks["build"].Outputs);
    }

    [Fact]
    public async Task A_dry_run_of_version_write_changes_nothing()
    {
        Write("libs/core/widget.json", "{\n  \"name\": \"core\",\n  \"version\": \"1.0.0\"\n}\n");
        await using var client = await Connect();
        var dry = await client.WriteVersionAsync(new VersionParams { Project = Core, Version = "1.1.0", DryRun = true });
        Assert.Equal(["libs/core/widget.json"], dry.Files);
        Assert.Contains("\"1.0.0\"", Read("libs/core/widget.json"), StringComparison.Ordinal);
        await client.WriteVersionAsync(new VersionParams { Project = Core, Version = "1.1.0" });
        Assert.Contains("\"version\": \"1.1.0\"", Read("libs/core/widget.json"), StringComparison.Ordinal);
    }

    [Fact]
    public async Task Doctor_offers_a_fix_for_a_missing_license()
    {
        Write("libs/core/widget.json", "{\"name\":\"core\",\"version\":\"1.0.0\"}");
        await using var client = await Connect();
        var finding = (await client.DoctorAsync(new DoctorParams { Projects = [Core] })).Findings[0];
        Assert.False(finding.Ok);
        Assert.NotEmpty(finding.Remedy);
        Assert.False(finding.Fix!.Safe);
        Assert.Contains("\"license\": \"MIT\"", finding.Fix.Files[0].Content, StringComparison.Ordinal);
        Assert.Equal(["core"], finding.Fix.AppliesTo);
    }

    [Fact]
    public void The_manifest_must_name_the_widget_and_give_an_x_y_z_version()
    {
        Assert.Throws<InvalidOperationException>(() => new WidgetManifest(string.Empty, "1.0.0", string.Empty, string.Empty, string.Empty, [], string.Empty).Validate());
        Assert.Throws<InvalidOperationException>(() => new WidgetManifest("core", "1.0", string.Empty, string.Empty, string.Empty, [], string.Empty).Validate());
        new WidgetManifest("core", "1.0.0", string.Empty, string.Empty, string.Empty, [], string.Empty).Validate();
    }
}
