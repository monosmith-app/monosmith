namespace MonoSmith.Plugin.Sdk.Tests.Projects;

public sealed class ProjectsTests
{
    [Theory]
    [InlineData("package.json", "package.json", true)]
    [InlineData("*.csproj", "Widget.csproj", true)]
    [InlineData("*.csproj", "Widget.fsproj", false)]
    [InlineData("go.?od", "go.mod", true)]
    [InlineData("go.?od", "go.mmod", false)]
    [InlineData("[a-c]x", "bx", true)]
    [InlineData("[a-c]x", "dx", false)]
    [InlineData("[^a-c]x", "dx", true)]
    [InlineData("a\\*b", "a*b", true)]
    [InlineData("a\\*b", "aXb", false)]
    [InlineData("[", "[", true)]
    [InlineData("[x", "x", false)]
    [InlineData("a\\", "a", false)]
    [InlineData("*", "dir/file", false)]
    public void Matches_a_manifest_by_name_or_by_the_pattern_Go_matches(string manifest, string file, bool expected) =>
        Assert.Equal(expected, ManifestMatching.IsManifest([manifest], file));

    [Fact]
    public void Matches_any_of_the_manifests()
    {
        Assert.True(ManifestMatching.IsManifest(["a.json", "*.csproj"], "b.csproj"));
        Assert.False(ManifestMatching.IsManifest([], "b.csproj"));
    }

    [Fact]
    public void Resolves_the_kind_from_the_manifest_then_the_directory_then_the_fallback()
    {
        Assert.Equal(ProjectKind.App, ProjectKinds.Resolve("app", "libs/x", ProjectKind.Lib));
        Assert.Equal(ProjectKind.App, ProjectKinds.Resolve(" app ", "libs/x", ProjectKind.Lib));
        Assert.Equal(ProjectKind.Lib, ProjectKinds.Resolve("service", "libs/x", ProjectKind.App));
        Assert.Equal(ProjectKind.Package, ProjectKinds.Resolve(null, "packages/x", ProjectKind.App));
        Assert.Equal(ProjectKind.App, ProjectKinds.Resolve("", "apps/x", ProjectKind.Lib));
        Assert.Equal(ProjectKind.Lib, ProjectKinds.Resolve(null, "tools/x", ProjectKind.Lib));
        Assert.Equal("apps", ProjectKinds.DirectoryOf(ProjectKind.App));
        Assert.Equal("libs", ProjectKinds.DirectoryOf(ProjectKind.Lib));
        Assert.Equal("packages", ProjectKinds.DirectoryOf(ProjectKind.Package));
    }

    [Fact]
    public void Publishes_a_package_unless_the_workspace_says_none()
    {
        var package = new Project { Name = "p", Root = "packages/p", Language = "l", Kind = ProjectKind.Package };
        Assert.True(ProjectPublishing.Publishes(package));
        Assert.False(ProjectPublishing.Publishes(package with { Registries = new Registries { NoPublish = true } }));
        Assert.False(ProjectPublishing.Publishes(package with { Kind = ProjectKind.Lib }));
    }

    [Fact]
    public void Reads_the_audit_policy_choices()
    {
        var audit = new AuditSettings { Policies = ["node"], Warn = ["go"] };
        Assert.True(AuditPolicies.IsEnabled(audit, "node"));
        Assert.False(AuditPolicies.IsEnabled(audit, "go"));
        Assert.False(AuditPolicies.IsEnabled(null, "node"));
        Assert.True(AuditPolicies.Blocks(audit, "node"));
        Assert.False(AuditPolicies.Blocks(audit, "go"));
        Assert.True(AuditPolicies.Blocks(null, "go"));
    }

    [Fact]
    public void Records_the_toolchain_on_every_project_and_asks_for_it_once_and_only_when_there_is_a_project()
    {
        var asked = 0;
        string Probe()
        {
            asked++;
            return "dotnet 8";
        }

        var empty = new DiscoverResult { Projects = [] };
        Assert.Same(empty, ToolchainRecording.Set(empty, Probe));
        Assert.Equal(0, asked);
        var two = new DiscoverResult
        {
            Projects =
            [
                new Project { Name = "a", Root = "a", Language = "l", Kind = ProjectKind.Lib },
                new Project { Name = "b", Root = "b", Language = "l", Kind = ProjectKind.Lib },
            ],
        };
        var result = ToolchainRecording.Set(two, Probe);
        Assert.Equal(1, asked);
        Assert.All(result.Projects, project => Assert.Equal("dotnet 8", project.Toolchain));
    }

    [Fact]
    public void Lists_the_plugins_of_a_stack_after_its_language_and_scaffold()
    {
        var stack = new Stack { Name = "s", Language = "node", Toolchain = ["t1", "t2"], Structure = ["s1"], Deploy = ["d1"], Scaffold = "sc" };
        Assert.Equal(["t1", "t2", "s1", "d1"], StackPlugins.Of(stack));
    }
}
