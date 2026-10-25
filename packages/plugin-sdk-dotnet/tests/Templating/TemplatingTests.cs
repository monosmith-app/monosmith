using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Templating;

public sealed class TemplatingTests
{
    private const string Data =
        "{\"Dir\":\"apps/web\",\"Name\":\"web\",\"Kind\":\"app\",\"Stack\":\"\",\"Port\":8080,\"Tags\":[\"a\",\"b\"],"
        + "\"Workspace\":{\"Name\":\"acme\",\"Scope\":\"@acme\"},\"Empty\":[]}";

    private static string Render(string template, string data = Data, string name = "template") =>
        Templates.Render(template, JsonAny.Parse(data), name);

    [Fact]
    public void Prints_fields_dotted_fields_and_literals() =>
        Assert.Equal(
            "web in apps/web on 8080 of @acme [a b] [[ 3 true",
            Render("[[.Name]] in [[.Dir]] on [[.Port]] of [[.Workspace.Scope]] [[.Tags]] [[ \"[[\" ]] [[ 3 ]] [[ true ]]"));

    [Fact]
    public void Chooses_with_if_else_if_and_else()
    {
        const string template = "[[if eq .Kind \"app\"]]app[[else if eq .Kind \"lib\"]]lib[[else]]other[[end]]";
        Assert.Equal("app", Render(template));
        Assert.Equal("lib", Render(template, Data.Replace("\"app\"", "\"lib\"", StringComparison.Ordinal)));
        Assert.Equal("other", Render(template, Data.Replace("\"app\"", "\"package\"", StringComparison.Ordinal)));
        Assert.Equal("|none", Render("[[if .Stack]]stack: [[.Stack]][[end]]|[[if not .Stack]]none[[end]]"));
        Assert.Equal("xy", Render("[[if ne .Kind \"package\"]]x[[end]][[if and .Name .Port]]y[[end]][[if or .Stack .Empty]]z[[end]]"));
    }

    [Fact]
    public void Writes_the_optional_member_of_a_manifest_the_way_the_SDKs_own_templates_do()
    {
        const string template = "{\n  \"license\": \"MIT\"[[if .Stack]],\n  \"stack\": \"[[.Stack]]\"[[end]]\n}\n";
        Assert.Equal("{\n  \"license\": \"MIT\"\n}\n", Render(template));
        Assert.Equal(
            "{\n  \"license\": \"MIT\",\n  \"stack\": \"widget\"\n}\n",
            Render(template, Data.Replace("\"Stack\":\"\"", "\"Stack\":\"widget\"", StringComparison.Ordinal)));
    }

    [Fact]
    public void Narrows_with_with_loops_with_range_and_reaches_the_root_with_a_dollar()
    {
        Assert.Equal("acme/web", Render("[[with .Workspace]][[.Name]]/[[$.Name]][[end]]"));
        Assert.Equal("<a><b>", Render("[[range .Tags]]<[[.]]>[[end]]"));
        Assert.Equal("none", Render("[[range .Empty]]x[[else]]none[[end]]"));
        Assert.Equal("acme@acme", Render("[[range .Workspace]][[.]][[end]]"));
    }

    [Fact]
    public void Calls_the_functions_of_text_template() =>
        Assert.Equal(
            "2 3 b true true web:8080 a1 2b \"web\"",
            Render(
                "[[len .Tags]] [[len .Name]] [[index .Tags 1]] [[lt .Port 9000]] [[ge .Name \"web\"]] "
                + "[[printf \"%s:%d\" .Name .Port]] [[print \"a\" 1 2 \"b\"]] [[.Name | printf \"%q\"]]"));

    [Fact]
    public void Trims_whitespace_with_the_dash_markers_and_skips_comments()
    {
        Assert.Equal("awebb", Render("a  \n[[- .Name -]]\n  b[[/* a comment */]]"));
        Assert.Equal("x-1", Render("x[[ -1 ]]"));
    }

    [Theory]
    [InlineData("[[.Missing]]", "map has no entry for key \"Missing\"")]
    [InlineData("[[.Name.Deeper]]", "can't evaluate field Deeper in string")]
    [InlineData("[[if .Name]]x", "unclosed if: missing [[end]]")]
    [InlineData("[[end]]", "unexpected end")]
    [InlineData("[[else]]", "unexpected else")]
    [InlineData("[[.Name", "unclosed action")]
    [InlineData("[[nope .Name]]", "function \"nope\" not defined")]
    [InlineData("[[eq .Name 3]]", "incompatible types for comparison")]
    [InlineData("[[$x := 1]]", "variables are not supported")]
    [InlineData("[[template \"x\"]]", "is not supported")]
    [InlineData("[[(.Name]]", "unclosed (")]
    [InlineData("[[.Name .Port]]", "can't give an argument")]
    [InlineData("[[index .Tags 5]]", "index out of range: 5")]
    [InlineData("[[len 3]]", "len of 3")]
    [InlineData("[[range 3]]x[[end]]", "range can't iterate over 3")]
    public void Fails_with_the_name_of_the_template_and_what_is_wrong(string template, string message)
    {
        var error = Assert.Throws<TemplateError>(() => Render(template, name: "main.go"));
        Assert.StartsWith("template main.go: ", error.Message, StringComparison.Ordinal);
        Assert.Contains(message, error.Message, StringComparison.Ordinal);
    }

    [Fact]
    public void Treats_nothing_as_an_empty_data_object_only_when_it_has_no_fields()
    {
        Assert.Equal("plain text", Templates.Render("plain text"));
        var error = Assert.Throws<TemplateError>(() => Templates.Render("[[.X]]"));
        Assert.Contains("can't evaluate field X in nil", error.Message, StringComparison.Ordinal);
    }

    [Fact]
    public void Renders_paths_and_bodies_and_leaves_GitHub_syntax_alone()
    {
        using var workspace = new TempWorkspace();
        var written = Templates.Write(
            workspace.Root,
            new Dictionary<string, string>
            {
                ["[[.Dir]]/main.go"] = "package main // [[.Name]]",
                ["[[.Dir]]/ci.yml"] = "run: ${{ secrets.TOKEN }}",
                ["[[.Dir]]/sub/doc.go"] = "package sub",
            },
            JsonAny.Parse("{\"Dir\":\"apps/web\",\"Name\":\"web\"}"));
        Assert.Equal(["apps/web/ci.yml", "apps/web/main.go", "apps/web/sub/doc.go"], written);
        Assert.Equal("package main // web", workspace.Read("apps/web/main.go"));
        Assert.Equal("run: ${{ secrets.TOKEN }}", workspace.Read("apps/web/ci.yml"));
    }

    [Fact]
    public void Refuses_to_overwrite_a_file_and_writes_nothing()
    {
        using var workspace = new TempWorkspace();
        workspace.Write("b.txt", "mine");
        var error = Assert.Throws<TemplateFileExistsError>(
            () => Templates.Write(workspace.Root, new Dictionary<string, string> { ["a.txt"] = "x", ["b.txt"] = "y" }));
        Assert.Equal("b.txt: file already exists", error.Message);
        Assert.Equal("b.txt", error.TargetPath);
        Assert.False(workspace.Exists("a.txt"));
        Assert.Equal("mine", workspace.Read("b.txt"));
    }

    [Fact]
    public void Fails_on_a_missing_key_writing_nothing()
    {
        using var workspace = new TempWorkspace();
        var error = Assert.Throws<TemplateError>(
            () => Templates.Write(workspace.Root, new Dictionary<string, string> { ["ok"] = "fine", ["a"] = "[[.Missing]]" }, new JsonMap()));
        Assert.Contains("map has no entry for key \"Missing\"", error.Message, StringComparison.Ordinal);
        Assert.False(workspace.Exists("ok"));
    }

    [Fact]
    public void Refuses_a_target_outside_the_root()
    {
        using var workspace = new TempWorkspace();
        Directory.CreateDirectory(workspace.Path("inside"));
        var up = Assert.Throws<TemplateError>(
            () => Templates.Write(workspace.Path("inside"), new Dictionary<string, string> { ["[[.Up]]/x"] = "y" }, new JsonMap { ["Up"] = ".." }));
        Assert.Contains("outside the project root", up.Message, StringComparison.Ordinal);
        var escape = Assert.Throws<TemplateError>(() => Templates.Write(workspace.Root, new Dictionary<string, string> { ["../escape"] = "y" }));
        Assert.Contains("outside the project root", escape.Message, StringComparison.Ordinal);
        var absolute = Assert.Throws<TemplateError>(() => Templates.Write(workspace.Root, new Dictionary<string, string> { ["/etc/x"] = "y" }));
        Assert.Contains("outside the project root", absolute.Message, StringComparison.Ordinal);
    }

    [Fact]
    public void Writes_UTF8_without_a_byte_order_mark()
    {
        using var workspace = new TempWorkspace();
        Templates.Write(workspace.Root, new Dictionary<string, string> { ["a.txt"] = "é" });
        Assert.Equal([0xC3, 0xA9], File.ReadAllBytes(workspace.Path("a.txt")));
    }
}
