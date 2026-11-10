using System.Text;
using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Integration;

/// <summary>
/// The compatibility rule of the SDKs: for the same plugin logic, a C# plugin, a Python plugin, a TypeScript plugin and a
/// Go plugin give byte-identical responses. Each trio is the same plugin written in every language. The one thing that
/// cannot be equal is the command of the tasks (the program's path, or an interpreter and a script), which is replaced
/// by a name on every side. Answers to requests that run together are compared by id: they may legitimately come out of
/// order. A side whose tool is not installed is left out locally; CI has them all.
/// </summary>
[Collection(ExampleBuildsCollection.Name)]
public sealed class CrossLanguageTests(ExampleSides examples) : IClassFixture<ExampleSides>
{
    private static readonly UTF8Encoding Utf8 = new(encoderShouldEmitUTF8Identifier: false);

    private static string Initialize(string root, string versions = "[1]", string host = "1.62.0") =>
        $"{{\"protocolVersions\":{versions},\"host\":{{\"name\":\"smith\",\"version\":\"{host}\"}},\"workspaceRoot\":{new JsonText(root).Encode()}}}";

    private static string Project(string name, string root, string kind) =>
        $"{{\"name\":\"{name}\",\"root\":\"{root}\",\"language\":\"widget\",\"kind\":\"{kind}\"}}";

    private static Dictionary<string, string> Snapshot(string root) =>
        Directory.GetFiles(root, "*", SearchOption.AllDirectories)
            .OrderBy(path => path, StringComparer.Ordinal)
            .ToDictionary(path => Path.GetRelativePath(root, path).Replace('\\', '/'), path => File.ReadAllText(path, Utf8));

    private static (Dictionary<string, List<string>> Answers, Dictionary<string, string> Roots) RunAll(
        Dictionary<string, Side> sides,
        Dictionary<string, string> files,
        Func<string, List<RawRequest>> requests,
        Dictionary<string, string>? environment = null,
        bool ordered = true)
    {
        var answers = new Dictionary<string, List<string>>();
        var roots = new Dictionary<string, string>();
        foreach (var (name, side) in sides)
        {
            var root = Directory.CreateTempSubdirectory($"ws-{name}-").FullName;
            foreach (var (relative, content) in files)
            {
                var path = Path.Combine(root, relative.Replace('/', Path.DirectorySeparatorChar));
                Directory.CreateDirectory(Path.GetDirectoryName(path)!);
                File.WriteAllText(path, content, Utf8);
            }

            var bodies = ProcessConversation.Converse(side.Command, root, requests(root), environment).Select(side.Normalise).ToList();
            answers[name] = ordered ? bodies : ProcessConversation.ById(bodies);
            roots[name] = root;
        }

        return (answers, roots);
    }

    private static void AssertIdentical(Dictionary<string, List<string>> answers)
    {
        Assert.Contains("go", answers.Keys);
        Assert.Contains("cs", answers.Keys);
        var reference = answers["go"];
        foreach (var (name, bodies) in answers)
        {
            Assert.True(bodies.Count == reference.Count, $"{name} answered {bodies.Count} messages, go {reference.Count}");
            for (var position = 0; position < reference.Count; position++)
            {
                Assert.True(reference[position] == bodies[position], $"{name}, response {position + 1}:\n{bodies[position]}\nexpected (go):\n{reference[position]}");
            }
        }
    }

    private Dictionary<string, Side> Widget() => examples.Sides("language-widget", "language-widget-ts", "language-widget-py", "language-widget-cs");

    private Dictionary<string, Side> Scan() => examples.Sides("structure-scan-go", "structure-scan-node", "structure-scan-py", "structure-scan-cs");

    private Dictionary<string, Side> Layout() => examples.Sides("structure-layout-go", "structure-layout-node", "structure-layout-py", "structure-layout-cs");

    private static readonly Dictionary<string, string> WidgetFiles = new()
    {
        ["libs/core/widget.json"] = "{\n  \"name\": \"core\",\n  \"version\": \"1.0.0\",\n  \"description\": \"Core <widgets> & more\",\n  \"kind\": \"lib\"\n}\n",
        ["apps/shop/widget.json"] = "{\"name\":\"shop\",\"version\":\"2.1.0\",\"kind\":\"app\",\"uses\":[\"core\",\"other\"],\"license\":\"MIT\",\"stack\":\"widget-plain\",\"description\":\"café\"}",
        ["libs/skip/widget.json"] = "{}",
    };

    [IntegrationFact("go", "dotnet")]
    public void The_language_widget_answers_the_handshake_discover_doctor_version_write_and_generate_with_the_same_bytes()
    {
        var core = Project("core", "libs/core", "lib");
        var shop = Project("shop", "apps/shop", "app");
        var (answers, roots) = RunAll(
            Widget(),
            WidgetFiles,
            root =>
            [
                new("initialize", Initialize(root)),
                new("discover", "{\"manifests\":[\"libs/core/widget.json\",\"apps/shop/widget.json\",\"libs/skip/widget.json\"]}"),
                new("discover", "{\"manifests\":[]}"),
                new("doctor", $"{{\"projects\":[{core},{shop}]}}"),
                new("doctor", "{\"projects\":[]}"),
                new("version/write", $"{{\"project\":{core},\"version\":\"1.1.0\",\"dryRun\":true}}"),
                new("version/write", $"{{\"project\":{core},\"version\":\"1.1.0\"}}"),
                new("version/write", $"{{\"project\":{shop},\"version\":\"2.2.0\"}}"),
                new("generate", "{\"kind\":\"package\",\"name\":\"gen\",\"directory\":\"packages/gen\",\"workspace\":{\"name\":\"acme\",\"scope\":\"@acme\"},\"stack\":\"widget-plain\"}"),
                new("generate", "{\"kind\":\"app\",\"name\":\"plain\",\"directory\":\"apps/plain\",\"workspace\":{\"name\":\"acme\"}}"),
                new("contribute", $"{{\"projects\":[{core},{shop}]}}"),
                new("nope", "{}"),
            ]);
        AssertIdentical(answers);
        var reference = answers["go"];
        Assert.True(reference.Count > 12);

        // The calls that write leave the same files, byte for byte.
        Assert.Equal(Snapshot(roots["go"]), Snapshot(roots["cs"]));
        Assert.Contains("packages/gen/widget.json", Snapshot(roots["cs"]).Keys);

        // Spot checks, so that equal does not mean equally empty.
        Assert.Equal(
            "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"protocolVersion\":1,\"plugin\":{\"name\":\"widget\",\"version\":\"0.1.0\"},"
            + "\"kind\":\"language\",\"capabilities\":{\"discover\":true,\"generate\":true,\"version\":true,\"doctor\":true},"
            + "\"languages\":[\"widget\"],\"manifests\":[\"widget.json\"],\"roles\":[\"language\",\"scaffold\"]}}",
            reference[0]);
        Assert.Contains("Core \\u003cwidgets\\u003e \\u0026 more", reference[1], StringComparison.Ordinal);
        Assert.Contains("\"imports\":[\"core\",\"other\"]", reference[1], StringComparison.Ordinal);
    }

    [IntegrationFact("go", "dotnet")]
    public void The_language_widget_refuses_a_request_before_initialize_a_version_it_does_not_speak_and_a_method_it_lacks()
    {
        var early = RunAll(Widget(), WidgetFiles, _ => [new("discover", "{\"manifests\":[]}")]).Answers;
        AssertIdentical(early);
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":1,\"error\":{\"code\":-32600,\"message\":\"initialize must come first\"}}", early["go"][0]);

        var version = RunAll(Widget(), WidgetFiles, root => [new("initialize", Initialize(root, "[2,3]", "9"))]).Answers;
        AssertIdentical(version);
        Assert.Contains("no common protocol version: host speaks [2 3], plugin speaks 1", version["go"][0], StringComparison.Ordinal);
    }

    [IntegrationFact("go", "dotnet")]
    public void The_language_widget_ignores_a_cancel_notification_it_did_not_advertise()
    {
        var answers = RunAll(
            Widget(),
            WidgetFiles,
            root => [new("initialize", Initialize(root)), new("cancel", "{\"id\":1}", Notification: true), new("discover", "{\"manifests\":[]}")]).Answers;
        AssertIdentical(answers);

        // initialize, discover and shutdown are answered; the notification is not, and no cancel or concurrency is declared.
        Assert.Equal(3, answers["go"].Count);
        Assert.DoesNotContain("\"cancel\"", answers["go"][0], StringComparison.Ordinal);
        Assert.DoesNotContain("\"concurrency\"", answers["go"][0], StringComparison.Ordinal);
    }

    [IntegrationFact("go", "dotnet")]
    public void The_language_widget_answers_params_of_the_wrong_shape_with_the_same_code()
    {
        var answers = RunAll(
            Widget(),
            WidgetFiles,
            root => [new("initialize", Initialize(root)), new("discover", "{\"manifests\":\"libs/core/widget.json\"}"), new("doctor", "[]")]).Answers;
        var codes = answers.ToDictionary(
            pair => pair.Key,
            pair => pair.Value.Select(body => ((JsonMap)JsonAny.Parse(body)).TryGetValue("error", out var error) ? ((JsonMap)error)["code"].ToString() : string.Empty).ToList());
        foreach (var (name, list) in codes)
        {
            Assert.True(list.SequenceEqual(codes["go"]), $"{name}: {string.Join(',', list)} against go: {string.Join(',', codes["go"])}");
        }

        Assert.Equal(["-32602", "-32602"], codes["go"].Skip(1).Take(2));
    }

    [IntegrationFact("go", "dotnet")]
    public void The_structure_scan_answers_the_handshake_and_doctor_with_the_same_bytes_concurrency_and_cancel_included()
    {
        var files = "abcdefgh".ToDictionary(letter => $"libs/core/{letter}.txt", letter => letter.ToString());
        files["libs/core/data/big.bin"] = new string('x', (1 << 20) + 1);
        var core = Project("core", "libs/core", "lib");
        var answers = RunAll(
            Scan(),
            files,
            root =>
            [
                new("initialize", Initialize(root)),
                new("doctor", $"{{\"projects\":[{core}]}}"),
                new("doctor", "{\"projects\":[]}"),
                new("cancel", "{\"id\":424242}", Notification: true),
                new("doctor", $"{{\"projects\":[{core}]}}"),
            ],
            ordered: false).Answers; // The scans run together, so they are compared by id.
        AssertIdentical(answers);
        var reference = answers["go"];
        Assert.Equal(5, reference.Count); // The notification gets no answer.
        Assert.Equal(
            "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"protocolVersion\":1,\"plugin\":{\"name\":\"widget-scan\",\"version\":\"0.1.0\"},"
            + "\"kind\":\"convention\",\"capabilities\":{\"doctor\":true,\"cancel\":true},\"concurrency\":4,"
            + "\"languages\":[\"widget\"],\"roles\":[\"structure\"]}}",
            reference[0]);
        Assert.Contains("\"check\":\"core has no file over 1 MiB\",\"ok\":false,\"detail\":\"data/big.bin\"", reference[1], StringComparison.Ordinal);
    }

    [IntegrationFact("go", "dotnet")]
    public void The_structure_scan_stops_a_scan_that_is_cancelled_and_answers_it_with_the_cancelled_error_and_the_same_bytes()
    {
        var files = "abcdefgh".ToDictionary(letter => $"libs/core/{letter}.txt", letter => letter.ToString());
        var core = Project("core", "libs/core", "lib");

        // A scan that would take 8 files times a minute: only the cancel ends it, so the answer is no matter of timing.
        var answers = RunAll(
            Scan(),
            files,
            root =>
            [
                new("initialize", Initialize(root)),
                new("doctor", $"{{\"projects\":[{core}]}}"),
                new("cancel", "{\"id\":2}", Notification: true),
                new("doctor", "{\"projects\":[]}"),
            ],
            new Dictionary<string, string> { ["WIDGET_SCAN_DELAY_MS"] = "60000" },
            ordered: false).Answers;
        AssertIdentical(answers);
        Assert.Equal(4, answers["go"].Count);
        Assert.Equal("{\"jsonrpc\":\"2.0\",\"id\":2,\"error\":{\"code\":-32800,\"message\":\"request cancelled: the scan was stopped\"}}", answers["go"][1]);
    }

    [IntegrationFact("go", "dotnet")]
    public void The_structure_layout_answers_the_handshake_contribute_and_doctor_with_the_same_bytes()
    {
        var core = Project("core", "libs/core", "lib");
        var web = Project("web", "apps/web", "app");
        var answers = RunAll(
            Layout(),
            new Dictionary<string, string> { ["libs/core/README.md"] = "# core\n" },
            root =>
            [
                new("initialize", Initialize(root)),
                new("contribute", $"{{\"projects\":[{core},{web}]}}"),
                new("contribute", "{\"projects\":[]}"),
                new("doctor", $"{{\"projects\":[{core},{web}]}}"),
                new("doctor", "{\"projects\":[]}"),
                new("discover", "{\"manifests\":[]}"),
            ]).Answers;
        AssertIdentical(answers);
        var reference = answers["go"];
        Assert.Equal(
            "{\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"protocolVersion\":1,\"plugin\":{\"name\":\"widget-layout\",\"version\":\"0.1.0\"},"
            + "\"kind\":\"convention\",\"capabilities\":{\"doctor\":true,\"contribute\":true},\"languages\":[\"widget\"],"
            + "\"roles\":[\"structure\"]}}",
            reference[0]);
        Assert.Contains("\"dependsOn\":{\"lint\":[\"layout-check\"]}", reference[1], StringComparison.Ordinal);
        Assert.Contains("\"warning\":true", reference[3], StringComparison.Ordinal);
    }
}
