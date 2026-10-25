using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Scaffold;

public sealed class ScaffoldTests
{
    private static GenerateParams ParamsOf(string stack = "") => new()
    {
        Kind = ProjectKind.Lib,
        Name = "core",
        Directory = "libs/core",
        Stack = stack,
        Workspace = new WorkspaceInfo { Name = "acme", Scope = "@acme" },
    };

    private static readonly Dictionary<string, string> Files = new()
    {
        ["[[.Dir]]/widget.json"] = "{\n  \"name\": \"[[.Scope]]/[[.Name]]\"[[if .Stack]],\n  \"stack\": \"[[.Stack]]\"[[end]]\n}\n",
        ["[[.Dir]]/README.md"] = "# [[.Name]] of [[.Workspace]] ([[.Greeting]])\n",
    };

    [Fact]
    public void Renders_the_templates_into_the_project_and_reports_the_files_and_the_notes()
    {
        using var workspace = new TempWorkspace();
        var result = Sdk.Scaffold.Project(workspace.Root, ParamsOf("widget"), Files, new JsonMap { ["Greeting"] = "hi" }, ["run npm install"]);
        Assert.Equal(["libs/core/README.md", "libs/core/widget.json"], result.Files);
        Assert.Equal(["run npm install"], result.Notes);
        Assert.Equal("# core of acme (hi)\n", workspace.Read("libs/core/README.md"));
        Assert.Equal("{\n  \"name\": \"@acme/core\",\n  \"stack\": \"widget\"\n}\n", workspace.Read("libs/core/widget.json"));
    }

    [Fact]
    public void Lets_the_extra_data_win_over_the_fields_of_the_params()
    {
        using var workspace = new TempWorkspace();
        Sdk.Scaffold.Project(workspace.Root, ParamsOf(), Files, new JsonMap { ["Greeting"] = "hi", ["Name"] = "renamed" });
        Assert.Equal("# renamed of acme (hi)\n", workspace.Read("libs/core/README.md"));
    }

    [Fact]
    public void Never_overwrites_a_file()
    {
        using var workspace = new TempWorkspace();
        workspace.Write("libs/core/README.md", "mine");
        Assert.Throws<TemplateFileExistsError>(() => Sdk.Scaffold.Project(workspace.Root, ParamsOf(), Files, new JsonMap { ["Greeting"] = "hi" }));
        Assert.Equal("mine", workspace.Read("libs/core/README.md"));
        Assert.False(workspace.Exists("libs/core/widget.json"));
    }
}
