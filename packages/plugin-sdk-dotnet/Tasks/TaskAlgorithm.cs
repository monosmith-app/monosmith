using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Building the tasks a plugin contributes or discovers.</summary>
public static class ProjectTasks
{
    /// <summary>
    /// A task the host can run, hash and cache. The command is an argv list and is never passed through a shell. The
    /// other arguments are the task's other members, left out when <c>null</c>.
    /// </summary>
    /// <example>
    /// <code>
    /// ProjectTasks.Create(["dotnet", "check.dll"], inputs: ["widget.json"], cache: true, description: "check the manifest")
    /// </code>
    /// </example>
    /// <param name="command">The argv.</param>
    /// <param name="cwd">The workspace-relative directory to run in; the project root when <c>null</c>.</param>
    /// <param name="env">Environment variables to add.</param>
    /// <param name="inputs">Globs relative to the project root; every file of the project when <c>null</c>.</param>
    /// <param name="outputs">Directories relative to the project root, restored on a cache hit.</param>
    /// <param name="dependsOn">Task names that run first (<c>build</c>, <c>^build</c>).</param>
    /// <param name="cache">Whether the result is cached.</param>
    /// <param name="description">What <c>smith explain</c> and the TUI show.</param>
    /// <param name="verify">Whether CI's verify phase runs the task too.</param>
    /// <param name="serve">A task smith starts before this one and stops after it.</param>
    /// <param name="serveUrl">What smith waits for (an HTTP 2xx) before it runs the task.</param>
    /// <param name="serveTimeout">How long to wait for <paramref name="serveUrl"/>, as a Go duration (<c>60s</c>).</param>
    /// <returns>The task.</returns>
    public static ProjectTask Create(
        IEnumerable<string> command,
        string? cwd = null,
        IReadOnlyDictionary<string, string>? env = null,
        IEnumerable<string>? inputs = null,
        IEnumerable<string>? outputs = null,
        IEnumerable<string>? dependsOn = null,
        bool? cache = null,
        string? description = null,
        bool? verify = null,
        string? serve = null,
        string? serveUrl = null,
        string? serveTimeout = null)
    {
        ArgumentNullException.ThrowIfNull(command);
        return new ProjectTask
        {
            Command = [.. command],
            Cwd = cwd ?? string.Empty,
            Env = env ?? new Dictionary<string, string>(),
            Inputs = [.. inputs ?? []],
            Outputs = [.. outputs ?? []],
            DependsOn = [.. dependsOn ?? []],
            Cache = cache ?? false,
            Description = description ?? string.Empty,
            Verify = verify ?? false,
            Serve = serve ?? string.Empty,
            ServeUrl = serveUrl ?? string.Empty,
            ServeTimeout = serveTimeout ?? string.Empty,
        };
    }

    /// <summary>
    /// The <c>build</c> task of a project: cached, built into the project's directory under <c>dist/</c>. That directory
    /// is also its <c>outputs</c>, restored on a cache hit, and the task runs after the build of every project this one
    /// imports (<c>^build</c>).
    /// </summary>
    /// <example>
    /// <code>
    /// ProjectTasks.Build(project.Root, output => ["dotnet", "build.dll", output], inputs: ["widget.json"])
    /// </code>
    /// </example>
    /// <param name="root">The project's root, workspace-relative.</param>
    /// <param name="command">Gets the directory the command is told to write into, <c>../../dist/libs/core</c> for <c>libs/core</c>, and returns the argv.</param>
    /// <param name="inputs">Globs relative to the project root.</param>
    /// <param name="description">What <c>smith explain</c> and the TUI show.</param>
    /// <returns>The task.</returns>
    public static ProjectTask Build(
        string root,
        Func<string, IEnumerable<string>> command,
        IEnumerable<string>? inputs = null,
        string? description = null)
    {
        ArgumentNullException.ThrowIfNull(command);
        var output = BuildDirectory.FromProject(root);
        return Create(command(output), inputs: inputs, outputs: [output], dependsOn: ["^build"], cache: true, description: description);
    }
}
