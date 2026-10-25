namespace MonoSmith.Plugin.Sdk.Tests.Tasks;

public sealed class TasksTests
{
    private static string Write(IProtocolWritable message)
    {
        var writer = new JsonBuilder();
        message.Write(writer);
        return writer.ToString();
    }

    [Fact]
    public void A_task_leaves_out_what_is_not_given()
    {
        Assert.Equal("{\"command\":[\"dotnet\",\"check.dll\"]}", Write(ProjectTasks.Create(["dotnet", "check.dll"])));
        Assert.Equal(
            "{\"command\":[\"x\"],\"env\":{\"A\":\"b\"},\"dependsOn\":[\"build\"],\"cache\":true,\"serveUrl\":\"http://h\"}",
            Write(ProjectTasks.Create(["x"], cache: true, dependsOn: ["build"], serveUrl: "http://h", env: new Dictionary<string, string> { ["A"] = "b" })));
    }

    [Fact]
    public void A_build_task_builds_into_the_directory_of_the_project_under_dist()
    {
        var built = ProjectTasks.Build("libs/core", output => ["tool", output], inputs: ["widget.json"]);
        Assert.Equal(
            "{\"command\":[\"tool\",\"../../dist/libs/core\"],\"inputs\":[\"widget.json\"],\"outputs\":[\"../../dist/libs/core\"],\"dependsOn\":[\"^build\"],\"cache\":true}",
            Write(built));
    }

    [Fact]
    public void Contributes_by_project_name_and_skips_a_project_it_has_nothing_for()
    {
        var projects = new[]
        {
            new Project { Name = "a", Root = "a", Language = "w", Kind = ProjectKind.Lib },
            new Project { Name = "b", Root = "b", Language = "w", Kind = ProjectKind.Lib },
        };
        var result = Contribute.To(projects, project => project.Name == "a"
            ? new Contribution { Tasks = new Dictionary<string, ProjectTask> { ["x"] = ProjectTasks.Create(["x"]) } }
            : null);
        Assert.Equal(["a"], result.Projects.Keys);
        Assert.Equal("{\"projects\":{\"a\":{\"tasks\":{\"x\":{\"command\":[\"x\"]}}}}}", Write(result));
    }

    [Fact]
    public void Builds_the_answers_of_ci()
    {
        Assert.Equal(
            "{\"uses\":\"actions/setup-node@v6\",\"inputs\":[{\"name\":\"node-version\",\"value\":\"24\"}]}",
            Write(CiSteps.Action("actions/setup-node@v6", [KeyValuePair.Create("node-version", "24")])));
        Assert.Equal("{\"uses\":\"NodeTool@0\"}", Write(CiSteps.Action("NodeTool@0")));
        Assert.Equal(
            "{\"name\":\"audit\",\"command\":[\"pip-audit\"],\"blocking\":false,\"cacheFiles\":[\"a\"]}",
            Write(CiSteps.Command("audit", ["pip-audit"], blocking: false, cacheFiles: ["a"])));
        Assert.Equal(
            "{\"env\":\"TOKEN\",\"description\":\"a token\",\"required\":true,\"setup\":true}",
            Write(CiSteps.Credential("TOKEN", "a token", setup: true)));
    }

    [Fact]
    public void Knows_where_builds_go()
    {
        Assert.Equal("dist/apps/web", BuildDirectory.ForProject("/apps/web/"));
        Assert.Equal("../../dist/apps/web", BuildDirectory.FromProject("apps/web"));
        Assert.Equal("../dist/web", BuildDirectory.FromProject("web"));
        Assert.True(BuildDirectory.Contains("dist"));
        Assert.True(BuildDirectory.Contains("dist/x"));
        Assert.False(BuildDirectory.Contains("distant"));
    }

    [Fact]
    public void Describes_references_to_other_plugins()
    {
        Assert.Equal("{\"level\":\"language\",\"name\":\"node\"}", Write(PluginRefs.RequiresLanguage("node")));
        Assert.Equal("{\"name\":\"x\"}", Write(PluginRefs.ConflictsWith("x")));
        Assert.Equal("language plugin node", PluginRefs.Describe(PluginRefs.RequiresLanguage("node")));
        Assert.Equal("plugin providing test", PluginRefs.Describe(new PluginRef { Slot = "test" }));
        Assert.Equal("plugin", PluginRefs.Describe(new PluginRef()));
        Assert.True(PluginRefs.IsEmpty(new PluginRef()));
        Assert.False(PluginRefs.IsEmpty(new PluginRef { Slot = "test" }));
    }

    [Fact]
    public void Tells_what_a_scaffold_template_can_say()
    {
        var data = ScaffoldData.From(new GenerateParams
        {
            Kind = ProjectKind.Lib,
            Name = "core",
            Directory = "libs/core",
            Workspace = new WorkspaceInfo { Name = "acme" },
        });
        Assert.Equal(
            "{\"Dir\":\"libs/core\",\"Kind\":\"lib\",\"ModulePrefix\":\"\",\"Name\":\"core\",\"Scope\":\"\",\"Stack\":\"\",\"Workspace\":\"acme\"}",
            data.Encode());
    }
}
