using System.Diagnostics;
using System.Text;

namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>A message to send as it is: the id is the position in the conversation, from 1.</summary>
/// <param name="Method">The method.</param>
/// <param name="ParamsJson">The params as JSON text; <c>null</c> to send none.</param>
/// <param name="Notification">Send it as a notification (no id). The ids of the requests after it are still their positions.</param>
internal sealed record RawRequest(string Method, string? ParamsJson = null, bool Notification = false);

/// <summary>
/// Helpers of the integration tests: they start real processes (smith, go, node, python, the example plugins).
/// </summary>
internal static class ProcessConversation
{
    private static readonly TimeSpan Wait = TimeSpan.FromSeconds(120);

    /// <summary>Whether the build runs in a CI service, where a missing tool is a failure and not a reason to skip.</summary>
    public static bool InCi =>
        new[] { "CI", "GITHUB_ACTIONS", "TF_BUILD" }.Any(name => !string.IsNullOrEmpty(Environment.GetEnvironmentVariable(name)));

    /// <summary>The root of the repository: where smith.toml is, above the test binaries.</summary>
    public static string RepositoryRoot { get; } = FindRoot();

    /// <summary>docs/examples/plugins</summary>
    public static string Examples => Path.Combine(RepositoryRoot, "docs", "examples", "plugins");

    /// <summary>The path of the dotnet program: smith looks a bare name up as a plugin before it looks for a program.</summary>
    public static string DotnetPath { get; } = FindOnPath("dotnet");

    /// <summary>Runs a program to the end and returns its exit status and everything it printed.</summary>
    public static (int Status, string Output) RunAllowingFailure(string program, IEnumerable<string> arguments, string? directory = null, IDictionary<string, string>? environment = null)
    {
        var start = new ProcessStartInfo(program)
        {
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            WorkingDirectory = directory ?? Environment.CurrentDirectory,
        };
        foreach (var argument in arguments)
        {
            start.ArgumentList.Add(argument);
        }

        start.Environment.Remove("SMITH_WORKSPACE_ROOT");
        start.Environment["NO_COLOR"] = "1";
        foreach (var (name, value) in environment ?? new Dictionary<string, string>())
        {
            start.Environment[name] = value;
        }

        using var process = Process.Start(start)!;
        var error = process.StandardError.ReadToEndAsync();
        var output = process.StandardOutput.ReadToEnd();
        Assert.True(process.WaitForExit((int)Wait.TotalMilliseconds), $"{program} did not end");
        return (process.ExitCode, output + error.Result);
    }

    /// <summary>Whether a program starts and succeeds.</summary>
    public static bool Runs(string program, params string[] arguments)
    {
        try
        {
            var start = new ProcessStartInfo(program) { RedirectStandardOutput = true, RedirectStandardError = true, UseShellExecute = false };
            foreach (var argument in arguments)
            {
                start.ArgumentList.Add(argument);
            }

            using var process = Process.Start(start)!;
            process.StandardOutput.ReadToEnd();
            process.StandardError.ReadToEnd();
            return process.WaitForExit(60_000) && process.ExitCode == 0;
        }
        catch (Exception error) when (error is System.ComponentModel.Win32Exception or InvalidOperationException)
        {
            return false;
        }
    }

    /// <summary>Runs a program to the end and returns its output; fails the test when it does not succeed.</summary>
    public static string Run(string program, IEnumerable<string> arguments, string? directory = null, IDictionary<string, string>? environment = null)
    {
        var start = new ProcessStartInfo(program)
        {
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            UseShellExecute = false,
            WorkingDirectory = directory ?? Environment.CurrentDirectory,
        };
        foreach (var argument in arguments)
        {
            start.ArgumentList.Add(argument);
        }

        foreach (var (name, value) in environment ?? new Dictionary<string, string>())
        {
            start.Environment[name] = value;
        }

        using var process = Process.Start(start)!;
        var error = process.StandardError.ReadToEndAsync();
        var output = process.StandardOutput.ReadToEnd();
        Assert.True(process.WaitForExit((int)Wait.TotalMilliseconds), $"{program} did not end");
        Assert.True(process.ExitCode == 0, $"{program} {string.Join(' ', arguments)} failed ({process.ExitCode}): {error.Result}{output}");
        return output;
    }

    /// <summary>
    /// Starts a plugin the way smith does (<c>SMITH_PLUGIN=1</c>), writes every request as a frame and returns the body of
    /// every frame it answered with, exactly as the plugin wrote it. The conversation ends when the plugin answers
    /// <c>shutdown</c>, which this adds to the end.
    /// </summary>
    public static List<string> Converse(IReadOnlyList<string> command, string directory, IReadOnlyList<RawRequest> requests, IDictionary<string, string>? environment = null)
    {
        var start = new ProcessStartInfo(command[0])
        {
            UseShellExecute = false,
            RedirectStandardInput = true,
            RedirectStandardOutput = true,
            WorkingDirectory = directory,
            CreateNoWindow = true,
        };
        foreach (var argument in command.Skip(1))
        {
            start.ArgumentList.Add(argument);
        }

        start.Environment["SMITH_PLUGIN"] = "1";
        foreach (var (name, value) in environment ?? new Dictionary<string, string>())
        {
            start.Environment[name] = value;
        }

        using var child = Process.Start(start)!;
        var bodies = new List<string>();
        var reading = Task.Run(() =>
        {
            var reader = new FrameReader();
            var buffer = new byte[65536];
            var stream = child.StandardOutput.BaseStream;
            int read;
            while ((read = stream.Read(buffer, 0, buffer.Length)) > 0)
            {
                lock (bodies)
                {
                    bodies.AddRange(reader.Push(buffer.AsSpan(0, read)));
                }
            }
        });
        try
        {
            var messages = requests.Append(new RawRequest("shutdown", "{}")).ToList();
            for (var position = 1; position <= messages.Count; position++)
            {
                var message = messages[position - 1];
                var body = new StringBuilder("{\"jsonrpc\":\"2.0\"");
                if (!message.Notification)
                {
                    body.Append(",\"id\":").Append(position);
                }

                body.Append(",\"method\":\"").Append(message.Method).Append('"');
                if (message.ParamsJson is not null)
                {
                    body.Append(",\"params\":").Append(message.ParamsJson);
                }

                var frame = Sdk.Framing.Frame(body.Append('}').ToString());
                child.StandardInput.BaseStream.Write(frame, 0, frame.Length);
                child.StandardInput.BaseStream.Flush();
            }

            // A plugin that refused every request does not stop at shutdown: closing its input does.
            child.StandardInput.Close();
            Assert.True(child.WaitForExit((int)Wait.TotalMilliseconds), "the plugin did not end");
        }
        finally
        {
            if (!child.HasExited)
            {
                child.Kill(entireProcessTree: true);
            }
        }

        Assert.True(reading.Wait(Wait), "the output of the plugin did not end");
        return bodies;
    }

    /// <summary>The bodies in the order of their ids: a plugin that runs requests together may answer in any order.</summary>
    public static List<string> ById(IEnumerable<string> bodies) =>
        [.. bodies.OrderBy(body => ((JsonNumber)((JsonMap)JsonAny.Parse(body))["id"]).Integer)];

    /// <summary>The program that runs Python with the SDK importable, or <c>null</c>.</summary>
    public static string? PythonWithSdk()
    {
        var venv = Path.Combine(RepositoryRoot, ".venv", OperatingSystem.IsWindows() ? "Scripts" : "bin", OperatingSystem.IsWindows() ? "python.exe" : "python");
        foreach (var candidate in new[] { venv, "python3", "python" })
        {
            if (Runs(candidate, "-c", "import monosmith_plugin_sdk"))
            {
                return candidate;
            }
        }

        return null;
    }

    private static string FindOnPath(string program)
    {
        var names = OperatingSystem.IsWindows() ? new[] { program + ".exe", program + ".cmd" } : [program];
        foreach (var directory in (Environment.GetEnvironmentVariable("PATH") ?? string.Empty).Split(Path.PathSeparator, StringSplitOptions.RemoveEmptyEntries))
        {
            foreach (var name in names)
            {
                var candidate = Path.Combine(directory, name);
                if (File.Exists(candidate))
                {
                    return candidate;
                }
            }
        }

        return program;
    }

    private static string FindRoot()
    {
        for (var directory = new DirectoryInfo(AppContext.BaseDirectory); directory is not null; directory = directory.Parent)
        {
            if (File.Exists(Path.Combine(directory.FullName, "smith.toml")) && Directory.Exists(Path.Combine(directory.FullName, "docs", "examples")))
            {
                return directory.FullName;
            }
        }

        return Environment.CurrentDirectory;
    }
}
