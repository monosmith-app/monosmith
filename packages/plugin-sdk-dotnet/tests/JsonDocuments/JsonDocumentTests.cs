namespace MonoSmith.Plugin.Sdk.Tests.JsonDocuments;

public sealed class JsonDocumentTests
{
    [Fact]
    public void Keeps_the_order_of_the_keys_the_text_of_leaves_and_edits_in_place()
    {
        var document = EditableJsonDocument.Parse("""
            {
              "name": "@acme/web",
              "version": "1.0.0",
              "nx": {
                "targets": {
                  "build": { "executor": "x" }
                }
              },
              "files": ["dist", "!**/*.map"],
              "scripts": {
                "test": "jest",
                "e2e": "a && b <c>"
              },
              "count": 1.50
            }

            """);
        document.Delete("nx");
        var scripts = document.Object("scripts");
        Assert.NotNull(scripts);
        scripts.SetString("build", "tsc -p tsconfig.lib.json");
        scripts.SetString("test", "jest --ci");
        document.Object("smith", create: true)!.SetString("kind", "package");
        Assert.Equal(
            """
            {
              "name": "@acme/web",
              "version": "1.0.0",
              "files": [
                "dist",
                "!**/*.map"
              ],
              "scripts": {
                "test": "jest --ci",
                "e2e": "a && b <c>",
                "build": "tsc -p tsconfig.lib.json"
              },
              "count": 1.50,
              "smith": {
                "kind": "package"
              }
            }

            """.ReplaceLineEndings("\n"),
            document.Render());
    }

    [Fact]
    public void Refuses_what_is_not_an_object_and_writes_an_empty_object_as_braces()
    {
        Assert.Contains("not a JSON object", Assert.Throws<JsonDocumentError>(() => EditableJsonDocument.Parse("[1]")).Message, StringComparison.Ordinal);
        Assert.Throws<JsonDocumentError>(() => EditableJsonDocument.Parse("{\"a\": "));
        Assert.Contains("invalid JSON value", Assert.Throws<JsonDocumentError>(() => EditableJsonDocument.Parse("{\"a\": nope}")).Message, StringComparison.Ordinal);
        Assert.Contains("invalid JSON value", Assert.Throws<JsonDocumentError>(() => EditableJsonDocument.Parse("{\"a\": NaN}")).Message, StringComparison.Ordinal);
        Assert.Contains("after the object", Assert.Throws<JsonDocumentError>(() => EditableJsonDocument.Parse("{\"a\": 1} x")).Message, StringComparison.Ordinal);
        Assert.Equal("{\n  \"a\": {}\n}\n", EditableJsonDocument.Parse("{\"a\": {}}").Render());
        Assert.Equal("{}\n", EditableJsonDocument.Parse("{}").Render());
    }

    [Fact]
    public void Makes_an_object_set_with_SetJson_editable_and_writes_other_values_as_they_are()
    {
        var document = EditableJsonDocument.Parse("{\"name\": \"x\"}");
        document.SetJson("jest", JsonAny.Parse("{\"preset\":\"ts-jest\",\"roots\":[\"<rootDir>/src\"]}"));
        document.Object("jest")!.SetString("testEnvironment", "node");
        document.SetJson("private", new JsonBool(true));
        Assert.Equal(
            """
            {
              "name": "x",
              "jest": {
                "preset": "ts-jest",
                "roots": [
                  "<rootDir>/src"
                ],
                "testEnvironment": "node"
              },
              "private": true
            }

            """.ReplaceLineEndings("\n"),
            document.Render());
    }

    [Fact]
    public void Writes_the_members_of_a_map_in_the_order_they_were_added()
    {
        var document = EditableJsonDocument.Parse("{}");
        document.SetJson("o", new JsonMap { ["z"] = 1L, ["a"] = 2L });
        Assert.Equal(["z", "a"], document.Object("o")!.Keys);
    }

    [Fact]
    public void Reads_strings_and_says_no_for_what_is_not_one()
    {
        var document = EditableJsonDocument.Parse("{\"a\": \"x\", \"n\": 1, \"o\": {}}");
        Assert.Equal("x", document.GetString("a"));
        Assert.All(new[] { "n", "o", "missing" }, key => Assert.Null(document.GetString(key)));
        Assert.Equal(new JsonNumber(1L), document.GetValue("n"));
        Assert.Equal(new JsonMap(), document.GetValue("o"));
        Assert.Null(document.GetValue("missing"));
        Assert.Equal(["a", "n", "o"], document.Keys);
        Assert.Equal(3, document.Count);
        Assert.True(document.Has("a"));
        Assert.False(document.Has("z"));
    }

    [Fact]
    public void Gives_a_duplicate_key_its_first_position_and_last_value_and_puts_a_new_key_last()
    {
        var document = EditableJsonDocument.Parse("{\"a\": 1, \"b\": 2, \"a\": 3}");
        document.SetString("c", "x");
        document.SetString("a", "y");
        Assert.Equal("{\n  \"a\": \"y\",\n  \"b\": 2,\n  \"c\": \"x\"\n}\n", document.Render());
    }

    [Fact]
    public void Formats_nested_arrays_and_escapes_as_Go_does()
    {
        var document = EditableJsonDocument.Parse("{\"matrix\": [[1,2],[3,[]],{\"a\":[ ],\"b\":\"x\\\"y\"}], \"s\": \"<>& \"}");
        Assert.Equal(
            """
            {
              "matrix": [
                [
                  1,
                  2
                ],
                [
                  3,
                  []
                ],
                {
                  "a": [],
                  "b": "x\"y"
                }
              ],
              "s": "<>& "
            }

            """.ReplaceLineEndings("\n"),
            document.Render());
        document.SetString("t", char.ConvertFromUtf32(0x2028) + "<");
        Assert.Contains("\"t\": \"" + "\\" + "u2028<\"", document.Render(), StringComparison.Ordinal);
    }
}
