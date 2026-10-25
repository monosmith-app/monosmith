namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>Finding a plugin's manifests below a workspace.</summary>
internal static class FindManifestsAlgorithm
{
    private static readonly HashSet<string> SkippedFolders = ["node_modules", "__pycache__"];

    /// <summary>
    /// The manifests of a plugin below a workspace, workspace-relative with forward slashes: files named like
    /// <paramref name="names"/> (see <see cref="ManifestMatching.IsManifest"/>), skipping dot-folders and dependency
    /// folders, and files at the root itself, as the Go conformance suite does.
    /// </summary>
    public static IReadOnlyList<string> Find(string workspace, IReadOnlyList<string> names)
    {
        var found = new List<string>();
        try
        {
            Walk(workspace, string.Empty, names, found);
        }
        catch (Exception error) when (error is IOException or UnauthorizedAccessException)
        {
            // An unreadable folder ends the walk; what was found stays.
        }

        return found;
    }

    private static void Walk(string workspace, string relative, IReadOnlyList<string> names, List<string> found)
    {
        var folder = relative.Length == 0 ? workspace : Path.Combine(workspace, relative.Replace('/', Path.DirectorySeparatorChar));
        var entries = new DirectoryInfo(folder).GetFileSystemInfos().OrderBy(entry => entry.Name, StringComparer.Ordinal);
        foreach (var entry in entries)
        {
            var path = relative.Length == 0 ? entry.Name : $"{relative}/{entry.Name}";
            if (entry is DirectoryInfo)
            {
                if (!entry.Name.StartsWith('.') && !SkippedFolders.Contains(entry.Name))
                {
                    Walk(workspace, path, names, found);
                }
            }
            else if (ManifestMatching.IsManifest(names, entry.Name) && path.Contains('/', StringComparison.Ordinal))
            {
                found.Add(path);
            }
        }
    }
}
