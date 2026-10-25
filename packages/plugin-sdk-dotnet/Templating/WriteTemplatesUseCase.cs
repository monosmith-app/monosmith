using System.Text;

namespace MonoSmith.Plugin.Sdk;

/// <summary><c>[[ ]]</c> templates, written for new projects and never over existing files.</summary>
public static partial class Templates
{
    private static readonly UTF8Encoding Utf8 = new(encoderShouldEmitUTF8Identifier: false);

    /// <summary>
    /// Renders every template in <paramref name="files"/> (target path to template text) under <paramref name="root"/>
    /// with <paramref name="data"/>, creating directories as needed. A target path may itself contain template actions
    /// (<c>[[.Dir]]/Program.cs</c>).
    /// </summary>
    /// <remarks>
    /// Nothing is written if any template fails to render or any target exists (<see cref="TemplateFileExistsError"/>):
    /// it never overwrites a file. A target that would land outside <paramref name="root"/> is refused.
    /// </remarks>
    /// <param name="root">The directory to write under.</param>
    /// <param name="files">Target paths and the templates of their content.</param>
    /// <param name="data">The data for the templates and the paths.</param>
    /// <returns>The written paths, forward-slashed and relative to the root, in the order of the sorted templates.</returns>
    public static IReadOnlyList<string> Write(string root, IReadOnlyDictionary<string, string> files, JsonAny? data = null)
    {
        ArgumentNullException.ThrowIfNull(root);
        ArgumentNullException.ThrowIfNull(files);
        var baseDirectory = Path.GetFullPath(root);
        var rendered = new List<(string Path, string FullPath, string Content)>();
        var targets = files.Keys.ToList();
        targets.Sort(GoJsonTextAlgorithm.CompareUtf8);
        foreach (var target in targets)
        {
            var path = NormalizePosix(Render(target, data, "path"));
            var full = Path.GetFullPath(Path.Combine(baseDirectory, path.Replace('/', Path.DirectorySeparatorChar)));
            var inside = full.StartsWith(baseDirectory + Path.DirectorySeparatorChar, StringComparison.Ordinal);
            if (path == ".." || path.StartsWith("../", StringComparison.Ordinal) || path.StartsWith('/') || Path.IsPathRooted(path) || !inside)
            {
                throw new TemplateError(path, "the target is outside the project root");
            }

            if (File.Exists(full) || Directory.Exists(full))
            {
                throw new TemplateFileExistsError(path);
            }

            rendered.Add((path, full, Render(files[target], data, path)));
        }

        foreach (var (_, full, content) in rendered)
        {
            Directory.CreateDirectory(Path.GetDirectoryName(full)!);
            File.WriteAllBytes(full, Utf8.GetBytes(content));
        }

        return [.. rendered.Select(entry => entry.Path)];
    }

    /// <summary>Python's <c>posixpath.normpath</c>: collapses <c>.</c>, <c>..</c> and repeated slashes.</summary>
    private static string NormalizePosix(string path)
    {
        var rooted = path.StartsWith('/');
        var parts = new List<string>();
        foreach (var part in path.Split('/'))
        {
            if (part is "" or ".")
            {
                continue;
            }

            if (part == ".." && parts.Count > 0 && parts[^1] != "..")
            {
                parts.RemoveAt(parts.Count - 1);
            }
            else if (part != ".." || !rooted)
            {
                parts.Add(part);
            }
        }

        var joined = string.Join('/', parts);
        return rooted ? "/" + joined : joined.Length == 0 ? "." : joined;
    }
}
