namespace MonoSmith.Plugin.Sdk.Tests.Protocol;

public sealed class ProtocolTests
{
    private static string Write(IProtocolWritable message)
    {
        var writer = new JsonBuilder();
        message.Write(writer);
        return writer.ToString();
    }

    private static T Read<T>(string json, string path = "params")
        where T : IProtocolValue<T> => T.Read(JsonAny.Parse(json), path);

    [Fact]
    public void Writes_required_members_always_and_optional_ones_only_when_they_are_not_zero()
    {
        Assert.Equal("{\"name\":\"n\",\"version\":\"\"}", Write(new Peer { Name = "n", Version = string.Empty }));
        Assert.Equal(
            "{\"name\":\"a\",\"root\":\"a\",\"language\":\"l\",\"kind\":\"lib\"}",
            Write(new Project { Name = "a", Root = "a", Language = "l", Kind = ProjectKind.Lib }));
        Assert.Equal(
            "{\"command\":[\"x\"],\"cache\":true}",
            Write(new ProjectTask { Command = ["x"], Cache = true }));
    }

    [Fact]
    public void Writes_a_required_list_as_an_empty_array_and_an_optional_one_not_at_all()
    {
        Assert.Equal("{\"projects\":[]}", Write(new DiscoverResult { Projects = [] }));
        Assert.Equal("{}", Write(new DependencyResult()));
        Assert.Equal("{}", Write(new Empty()));
    }

    [Fact]
    public void Writes_the_keys_of_a_map_sorted_and_the_members_in_the_order_of_the_schema()
    {
        var contribution = new Contribution
        {
            DependsOn = new Dictionary<string, IReadOnlyList<string>> { ["test"] = ["b"], ["lint"] = ["a"] },
            Tasks = new Dictionary<string, ProjectTask> { ["z"] = new() { Command = ["z"] }, ["a"] = new() { Command = ["a"] } },
        };
        Assert.Equal(
            "{\"tasks\":{\"a\":{\"command\":[\"a\"]},\"z\":{\"command\":[\"z\"]}},\"dependsOn\":{\"lint\":[\"a\"],\"test\":[\"b\"]}}",
            Write(contribution));
    }

    [Fact]
    public void Reads_an_absent_or_null_member_as_the_zero_value_of_its_type()
    {
        var project = Read<Project>("{\"name\":\"a\",\"exports\":null,\"tasks\":null,\"registries\":null}");
        Assert.Equal("a", project.Name);
        Assert.Equal(string.Empty, project.Root);
        Assert.Empty(project.Exports);
        Assert.Empty(project.Tasks);
        Assert.Null(project.Registries);
        Assert.True(project.Kind.IsEmpty);
        Assert.Equal(0, Read<Peer>("null").Name.Length);
    }

    [Fact]
    public void Ignores_members_it_does_not_know_so_the_protocol_can_grow()
    {
        var peer = Read<Peer>("{\"name\":\"n\",\"version\":\"1\",\"from\":\"the future\"}");
        Assert.Equal("1", peer.Version);
    }

    [Fact]
    public void Drops_a_null_member_of_a_map_unless_the_values_are_free_form()
    {
        var project = Read<Project>("{\"tasks\":{\"a\":null,\"b\":{\"command\":[\"x\"]}},\"settings\":{\"a\":null,\"b\":{}}}");
        Assert.Equal(["b"], project.Tasks.Keys);
        Assert.Equal(["b"], project.Settings.Keys);
    }

    [Theory]
    [InlineData("{\"name\":1}", "params.name must be a string, not a number")]
    [InlineData("{\"name\":\"a\",\"exports\":\"x\"}", "params.exports must be an array, not a string")]
    [InlineData("{\"name\":\"a\",\"exports\":[1]}", "params.exports[0] must be a string, not a number")]
    [InlineData("{\"name\":\"a\",\"tasks\":{\"t\":[]}}", "params.tasks.t must be an object, not an array")]
    [InlineData("{\"name\":\"a\",\"registries\":true}", "params.registries must be an object, not a boolean")]
    [InlineData("[]", "params must be an object, not an array")]
    public void Says_where_a_value_has_the_wrong_type(string json, string message)
    {
        var error = Assert.Throws<ParamsError>(() => Read<Project>(json));
        Assert.Equal(message, error.Message);
    }

    [Fact]
    public void Reads_an_integer()
    {
        Assert.Equal(3L, Read<InitializeResult>("{\"concurrency\":3}").Concurrency);
    }

    [Theory]
    [InlineData("3.0")]
    [InlineData("2.5")]
    [InlineData("1e2")]
    public void An_integer_member_written_as_anything_else_is_refused_as_Go_refuses_it(string number)
    {
        var error = Assert.Throws<ParamsError>(() => Read<InitializeResult>("{\"concurrency\":" + number + "}"));
        Assert.Equal("params.concurrency must be an integer, not a number", error.Message);
    }

    [Fact]
    public void Keeps_the_text_of_an_enumeration_it_does_not_know()
    {
        var project = Read<Project>("{\"kind\":\"service\"}");
        Assert.Equal("service", project.Kind.Value);
        Assert.NotEqual(ProjectKind.App, project.Kind);
        Assert.Equal("{\"name\":\"\",\"root\":\"\",\"language\":\"\",\"kind\":\"service\"}", Write(project));
        Assert.Equal(ProjectKind.Lib, (ProjectKind)"lib");
        Assert.Equal("lib", (string)ProjectKind.Lib);
    }

    [Fact]
    public void The_response_id_may_be_null()
    {
        Assert.Null(Read<Response>("{\"jsonrpc\":\"2.0\",\"id\":null}").Id);
        Assert.Equal(7L, Read<Response>("{\"jsonrpc\":\"2.0\",\"id\":7}").Id);
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":null}", Write(new Response { Jsonrpc = "2.0", Id = null }));
    }

    [Fact]
    public void Free_form_members_are_kept_as_they_came()
    {
        var request = Read<Request>("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"m\",\"params\":{\"z\":1,\"a\":[1,2]}}");
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"m\",\"params\":{\"a\":[1,2],\"z\":1}}", Write(request));
    }

    [Fact]
    public void The_description_of_a_plugin_becomes_an_initialize_result_with_the_version_and_the_capabilities()
    {
        var description = new PluginDescription
        {
            Plugin = new Peer { Name = "widget", Version = "0.1.0" },
            Kind = PluginKind.Language,
            Manifests = ["widget.json"],
        };
        var result = description.ToInitializeResult(new Capabilities { Discover = true });
        Assert.Equal(
            "{\"protocolVersion\":1,\"plugin\":{\"name\":\"widget\",\"version\":\"0.1.0\"},\"kind\":\"language\",\"capabilities\":{\"discover\":true},\"manifests\":[\"widget.json\"]}",
            Write(result));
    }

    [Fact]
    public void Knows_which_capabilities_are_advertised_by_their_names_in_the_protocol()
    {
        var capabilities = new Capabilities { Discover = true, Cancel = true, Dependencies = true };
        Assert.True(capabilities.Advertises("discover"));
        Assert.True(capabilities.Advertises("cancel"));
        Assert.True(capabilities.Advertises("dependencies"));
        Assert.False(capabilities.Advertises("publish"));
        Assert.False(capabilities.Advertises("nothing"));
    }

    [Fact]
    public void Has_the_constants_of_the_protocol()
    {
        Assert.Equal(1, ProtocolInfo.Version);
        Assert.Equal(64L * 1024 * 1024, ProtocolInfo.MaxMessageBytes);
        Assert.Equal(-32000, ErrorCodes.PluginFailure);
        Assert.Equal(-32800, ErrorCodes.RequestCancelled);
        Assert.Equal("version/write", ProtocolMethods.VersionWrite);
        Assert.Equal("dependencies/install", ProtocolMethods.DependenciesInstall);
        Assert.Equal("version", ProtocolMethods.CapabilityByMethod[ProtocolMethods.VersionWrite]);
        Assert.Null(ProtocolMethods.CapabilityByMethod[ProtocolMethods.Initialize]);
        Assert.Equal(["language", "toolchain", "scaffold", "structure", "deploy"], ProtocolRoles.All.Select(role => role.Name.Value));
    }

    [Fact]
    public void Writes_a_fix_with_its_nested_choices_and_the_member_the_schema_names_with_an_underscore()
    {
        var fix = new Fix
        {
            Description = "d",
            Safe = false,
            AppliesTo = ["a"],
            Choices = [new FixChoice { Question = "q", Options = [new FixChoiceOption { Label = "l", Fix = new Fix { Description = "inner", Safe = true } }] }],
        };
        var json = Write(fix);
        Assert.Equal(
            "{\"description\":\"d\",\"safe\":false,\"choices\":[{\"question\":\"q\",\"options\":[{\"label\":\"l\",\"fix\":{\"description\":\"inner\",\"safe\":true}}]}],\"applies_to\":[\"a\"]}",
            json);
        Assert.Equal("inner", Read<Fix>(json).Choices[0].Options[0].Fix.Description);
    }
}
