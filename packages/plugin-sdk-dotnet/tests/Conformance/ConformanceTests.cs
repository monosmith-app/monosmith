using MonoSmith.Plugin.Sdk.Testing;
using MonoSmith.Plugin.Sdk.Tests.Testing;

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

    /// <summary>A language whose manifest, fig.txt, holds "kind stack".</summary>
    private class Fig(string workspace) : IPlugin, IGenerator, IDiscoverer
    {
        protected string Workspace { get; } = workspace;

        public virtual PluginDescription Describe(InitializeParams parameters) => new()
        {
            Plugin = new Peer { Name = "fig", Version = "1.0.0" },
            Kind = PluginKind.Language,
            Manifests = ["fig.txt"],
        };

        public Task<GenerateResult> GenerateAsync(GenerateParams parameters, CancellationToken cancellation)
        {
            var folder = Path.Combine(Workspace, parameters.Directory.Replace('/', Path.DirectorySeparatorChar));
            Directory.CreateDirectory(folder);
            File.WriteAllText(Path.Combine(folder, "fig.txt"), $"{parameters.Kind} {parameters.Stack}");
            return Task.FromResult(new GenerateResult { Files = [$"{parameters.Directory}/fig.txt"] });
        }

        public Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation)
        {
            var projects = parameters.Manifests.Select(manifest =>
            {
                var parts = File.ReadAllText(Path.Combine(Workspace, manifest.Replace('/', Path.DirectorySeparatorChar))).Split(' ', 2);
                var root = manifest[..manifest.LastIndexOf('/')];
                return new Project
                {
                    Name = root[(root.LastIndexOf('/') + 1)..],
                    Root = root,
                    Language = "fig",
                    Kind = new ProjectKind(parts[0]),
                    Stack = parts[1],
                    Tasks = new Dictionary<string, ProjectTask> { ["build"] = new() { Command = ["fig", "build"] } },
                };
            }).ToList();
            return Task.FromResult(new DiscoverResult { Projects = projects });
        }
    }

    /// <summary>The same language, able to change kinds; <c>how</c> says how it misbehaves, if it does.</summary>
    private sealed class FigChanger(string workspace, string how = "") : Fig(workspace), IKindChanger
    {
        public Task<ChangeKindResult> ChangeKindAsync(ChangeKindParams parameters, CancellationToken cancellation)
        {
            var manifest = $"{parameters.Project.Root}/fig.txt";
            var content = $"{parameters.To} {ConformanceChecks.Stack}";
            switch (how)
            {
                case "nothing":
                    return Task.FromResult(new ChangeKindResult());
                case "missing":
                    return Task.FromResult(new ChangeKindResult { Files = [new FileEdit { Path = $"{parameters.Project.Root}/nowhere.txt", Content = "x" }] });
                case "write":
                    File.WriteAllText(Path.Combine(Workspace, manifest.Replace('/', Path.DirectorySeparatorChar)), content);
                    break;
            }

            return Task.FromResult(new ChangeKindResult { Files = [new FileEdit { Path = manifest, Content = content }] });
        }
    }

    private static async Task<ConformanceReport> CheckedAsync(Func<string, IPlugin> plugin)
    {
        using var workspace = new TempWorkspace();
        await using var client = await PluginClient.ConnectInProcessAsync(plugin(workspace.Root), new HostOptions { WorkspaceRoot = workspace.Root });
        return await ConformanceChecks.CheckPluginAsync(client, workspace.Root);
    }

    [Fact]
    public async Task A_plugin_that_changes_kinds_is_asked_what_each_generated_kind_becomes()
    {
        var report = await CheckedAsync(root => new FigChanger(root));
        Assert.True(report.Ok, report.Failures());
        var names = Names(report);
        Assert.Contains("answers what making the app a lib edits (dry run)", names);
        Assert.Contains("answers what making the lib a package edits (dry run)", names);
        Assert.Contains("answers what making the package a lib edits (dry run)", names);
    }

    private sealed class FigAppOnly(string workspace) : Fig(workspace)
    {
        public override PluginDescription Describe(InitializeParams parameters) => base.Describe(parameters) with { Kinds = [ProjectKind.App] };
    }

    [Fact]
    public async Task A_plugin_is_asked_to_generate_only_the_kinds_it_declares()
    {
        var report = await CheckedAsync(root => new FigAppOnly(root));
        Assert.True(report.Ok, report.Failures());
        var names = Names(report);
        Assert.Contains("discovers the generated app", names);
        Assert.DoesNotContain("generates a lib", names);
        Assert.DoesNotContain("generates a package", names);
    }

    [Fact]
    public async Task A_plugin_that_does_not_declare_change_kind_is_not_asked()
    {
        var report = await CheckedAsync(root => new Fig(root));
        Assert.DoesNotContain(Names(report), name => name.Contains("edits (dry run)", StringComparison.Ordinal));
    }

    [Theory]
    [InlineData("nothing", "neither a refusal nor an edit")]
    [InlineData("missing", "files that do not exist")]
    [InlineData("write", "it wrote to the workspace")]
    public async Task A_kind_change_that_answers_badly_is_reported(string how, string want)
    {
        var report = await CheckedAsync(root => new FigChanger(root, how));
        Assert.Contains(want, report.Failures(), StringComparison.Ordinal);
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
