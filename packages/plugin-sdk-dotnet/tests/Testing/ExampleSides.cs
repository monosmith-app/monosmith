namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>One implementation of a plugin: how to start it, and how it writes the command of its own tasks.</summary>
/// <param name="Command">The argv that starts it.</param>
/// <param name="SelfText">The text its tasks' command is written as in the answers, replaced by one name.</param>
public sealed record Side(IReadOnlyList<string> Command, string SelfText)
{
    /// <summary>The program and arguments to give <c>smith plugin check</c>: the program by its path, as a bare name is looked up as a plugin.</summary>
    public IReadOnlyList<string> CheckCommand { get; init; } = Command;

    /// <summary>The body with the command of the tasks replaced by <c>"&lt;self&gt;"</c>.</summary>
    public string Normalise(string body) => body.Replace(SelfText, "\"<self>\"", StringComparison.Ordinal);
}

/// <summary>
/// The examples that are the same plugin written in several languages, built once for the tests that compare them.
/// </summary>
public sealed class ExampleSides : IDisposable
{
    private readonly Dictionary<string, Side> _built = [];
    private readonly string _scratch = Directory.CreateTempSubdirectory("sdk-cross-").FullName;

    /// <summary>Which sides of a trio can run here: the Go one and the C# one always, the others when their tools are installed.</summary>
    public Dictionary<string, Side> Sides(string go, string typeScript, string python, string csharp)
    {
        var sides = new Dictionary<string, Side>
        {
            ["go"] = Go(go),
            ["cs"] = CSharp(csharp),
        };
        if (ProcessConversation.Runs("node", "--version") && Directory.Exists(Path.Combine(ProcessConversation.RepositoryRoot, "node_modules", "@monosmith", "plugin-sdk")))
        {
            sides["ts"] = TypeScript(typeScript);
        }

        if (ProcessConversation.PythonWithSdk() is { } interpreter)
        {
            sides["py"] = Python(interpreter, python);
        }

        return sides;
    }

    public Side Go(string example)
    {
        if (_built.TryGetValue(example, out var cached))
        {
            return cached;
        }

        var binary = Path.Combine(_scratch, example + (OperatingSystem.IsWindows() ? ".exe" : string.Empty));
        ProcessConversation.Run("go", ["build", "-o", binary, "."], Path.Combine(ProcessConversation.Examples, example), new Dictionary<string, string> { ["GOWORK"] = "off" });
        return _built[example] = new Side([binary], new JsonText(binary).Encode());
    }

    public Side CSharp(string example)
    {
        if (_built.TryGetValue(example, out var cached))
        {
            return cached;
        }

        var output = Path.Combine(_scratch, example);
        ProcessConversation.Run("dotnet", ["build", Path.Combine(ProcessConversation.Examples, example), "-o", output, "-nologo", "-v", "q"]);
        var assembly = Directory.GetFiles(output, "*.dll").Single(file => File.Exists(Path.ChangeExtension(file, ".runtimeconfig.json")));
        return _built[example] = new Side(["dotnet", assembly], "\"dotnet\"," + new JsonText(assembly).Encode()) { CheckCommand = [ProcessConversation.DotnetPath, assembly] };
    }

    public static Side TypeScript(string example)
    {
        var plugin = Path.Combine(ProcessConversation.Examples, example, "plugin.ts");
        return new Side(["node", plugin], "\"node\"," + new JsonText(plugin).Encode());
    }

    public static Side Python(string interpreter, string example)
    {
        var executable = ProcessConversation.Run(interpreter, ["-c", "import sys; print(sys.executable)"]).Trim();
        var plugin = Path.Combine(ProcessConversation.Examples, example, "plugin.py");
        return new Side([executable, plugin], new JsonText(executable).Encode() + "," + new JsonText(plugin).Encode());
    }

    public void Dispose()
    {
        try
        {
            Directory.Delete(_scratch, recursive: true);
        }
        catch (IOException)
        {
            // Left to the system's temporary folder cleanup.
        }
    }
}
