using System.Text;
using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Integration;

/// <summary>A real workspace that uses the C# examples as external plugins.</summary>
public sealed class ThrowawayWorkspaceTests : IClassFixture<ExampleSides>, IDisposable
{
    private readonly string _root = Directory.CreateTempSubdirectory("sdk-workspace-").FullName;
    private string _directory;
    private bool _prepared;

    public ThrowawayWorkspaceTests(ExampleSides examples)
    {
        _directory = _root;
        Examples = examples;
    }

    private ExampleSides Examples { get; }

    public void Dispose()
    {
        TempWorkspace.DeleteQuietly(_root);
    }

    /// <summary>A path as a TOML string in smith.toml: forward slashes, so Windows paths need no escaping.</summary>
    private static string Toml(string path) => "\"" + path.Replace('\\', '/') + "\"";

    private (int Status, string Output) Smith(params string[] arguments) =>
        ProcessConversation.RunAllowingFailure("smith", arguments, _directory, new Dictionary<string, string> { ["SMITH_HOME"] = Path.Combine(_root, "smith-home") });

    private void Prepare()
    {
        if (_prepared)
        {
            return;
        }

        _prepared = true;
        Assert.Equal(0, Smith("new", "ws").Status);
        _directory = Path.Combine(_root, "ws");
        var widget = Examples.CSharp("language-widget-cs").CheckCommand;
        var layout = Examples.CSharp("structure-layout-cs").CheckCommand;
        File.AppendAllText(
            Path.Combine(_directory, "smith.toml"),
            "\n[plugins.external.widget]\n"
            + $"command = [{Toml(widget[0])}, {Toml(widget[1])}]\n\n"
            + "[plugins.external.widget-layout]\n"
            + $"command = [{Toml(layout[0])}, {Toml(layout[1])}]\n\n"
            + "[conventions.widget-layout]\n",
            new UTF8Encoding(false));

        // External plugins run only once trusted; this is the user saying yes to both. --pin writes the digest to
        // smith.toml, which is what CI requires.
        foreach (var name in new[] { "widget", "widget-layout" })
        {
            var (status, output) = Smith("plugin", "trust", name, "--yes", "--pin");
            Assert.True(status == 0, output);
        }
    }

    [IntegrationFact("smith", "dotnet")]
    public void Works_as_a_whole_in_a_real_workspace()
    {
        Prepare();

        // Creates widget projects with the plugin and discovers them with their graph edge.
        Assert.Equal(0, Smith("add", "lib", "core", "--lang", "widget").Status);
        Assert.Equal(0, Smith("add", "app", "shop", "--lang", "widget").Status);
        var manifest = Path.Combine(_directory, "apps", "shop", "widget.json");
        var document = EditableJsonDocument.Parse(File.ReadAllText(manifest));
        document.SetJson("uses", new JsonList(["core"]));
        File.WriteAllText(manifest, document.Render(), new UTF8Encoding(false));
        var projects = Smith("projects").Output.Split('\n');
        Assert.Contains(projects, line => line.StartsWith("shop ", StringComparison.Ordinal) && line.Contains(" app ", StringComparison.Ordinal) && line.Contains("widget", StringComparison.Ordinal));
        Assert.Contains(projects, line => line.StartsWith("core ", StringComparison.Ordinal) && line.Contains(" lib ", StringComparison.Ordinal) && line.Contains("widget", StringComparison.Ordinal));
        Assert.Contains("core", Smith("graph").Output, StringComparison.Ordinal);

        // Runs the build and test tasks and serves the second run from the cache.
        var (status, output) = Smith("run", "build", "test");
        Assert.True(status == 0 && output.Contains("tasks succeeded", StringComparison.Ordinal), output);
        Assert.Equal("core 0.0.0\n", File.ReadAllText(Path.Combine(_directory, "dist", "libs", "core", "widget.txt")).ReplaceLineEndings("\n"));
        Assert.Contains("from cache", Smith("run", "build", "test").Output, StringComparison.Ordinal);

        // Adds the layout-check task to every widget project and fails on a junk drawer.
        var layout = Smith("run", "layout-check");
        Assert.True(layout.Status == 0 && layout.Output.Contains("layout-check", StringComparison.Ordinal), layout.Output);

        // A file in it: the task is cached on the files of the project, and an empty folder is none.
        Directory.CreateDirectory(Path.Combine(_directory, "libs", "core", "utils"));
        File.WriteAllText(Path.Combine(_directory, "libs", "core", "utils", "index.txt"), "everything");
        var broken = Smith("run", "layout-check");
        Assert.NotEqual(0, broken.Status);
        Assert.Contains("folders named for nothing: utils", broken.Output, StringComparison.Ordinal);

        // Makes the doctor finding of the structure plugin a warning about the README it asks for.
        Assert.Contains("has a README.md", Smith("doctor").Output, StringComparison.Ordinal);
    }
}
