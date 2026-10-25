namespace MonoSmith.Plugin.Sdk;

/// <summary>Where every build writes its output.</summary>
/// <remarks>
/// The directory at the workspace root every build writes into is <see cref="Name"/>. It is smith's opinion, so all
/// output is in one place, outside every source tree, and cleaning is deleting one folder. A plugin builds a project
/// into <see cref="ForProject"/>; <c>smith doctor</c> fails a task whose outputs fall outside it.
/// </remarks>
public static class BuildDirectory
{
    /// <summary>The directory at the workspace root every build writes into.</summary>
    public const string Name = "dist";

    /// <summary>Where a project's build goes, workspace-relative: <c>dist/&lt;project dir&gt;</c> (<c>dist/apps/web</c>).</summary>
    /// <param name="root">The project's root, workspace-relative.</param>
    /// <returns>The directory.</returns>
    public static string ForProject(string root)
    {
        ArgumentNullException.ThrowIfNull(root);
        return $"{Name}/{root.Trim('/')}";
    }

    /// <summary>
    /// The same directory relative to the project (<c>../../dist/apps/web</c>). It is what a tool configured from inside
    /// the project, or a task's <c>outputs</c>, name.
    /// </summary>
    /// <param name="root">The project's root, workspace-relative.</param>
    /// <returns>The directory.</returns>
    public static string FromProject(string root)
    {
        ArgumentNullException.ThrowIfNull(root);
        var trimmed = root.Trim('/');
        var depth = trimmed.Split('/').Length;
        return string.Concat(Enumerable.Repeat("../", depth)) + ForProject(trimmed);
    }

    /// <summary>Whether a workspace-relative path is inside the build directory.</summary>
    /// <param name="path">The path.</param>
    /// <returns>Whether it is inside.</returns>
    public static bool Contains(string path)
    {
        ArgumentNullException.ThrowIfNull(path);
        return path == Name || path.StartsWith($"{Name}/", StringComparison.Ordinal);
    }
}
