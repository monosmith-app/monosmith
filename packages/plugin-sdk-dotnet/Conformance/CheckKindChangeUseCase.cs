using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>The check of a plugin that changes kinds.</summary>
internal static class CheckKindChangeUseCase
{
    /// <summary>
    /// Asks what making a generated project another kind edits. The plugin must answer with a refusal or with edits to
    /// files that exist, and write nothing itself: the host applies the edits.
    /// </summary>
    public static async Task CheckAsync(PluginClient client, string workspace, Project project, IReadOnlyList<Project> projects, ConformanceReport report)
    {
        var to = TargetOf(project.Kind);
        var newRoot = $"{ProjectKinds.DirectoryOf(to)}/{project.Root.Split('/')[^1]}";
        var check = $"answers what making the {project.Kind} {Article(to)} edits (dry run)";
        var before = Snapshot(workspace, client.Info.Manifests);
        ChangeKindResult answer;
        try
        {
            answer = await client.ChangeKindAsync(new ChangeKindParams { Project = project, To = to, NewRoot = newRoot, Projects = projects }, timeout: TimeSpan.FromMinutes(5)).ConfigureAwait(false);
        }
        catch (Exception error)
        {
            report.Add(check, false, error.Message);
            return;
        }

        var missing = answer.Files.Select(file => file.Path)
            .Where(path => !File.Exists(Path.Combine(workspace, path.Replace('/', Path.DirectorySeparatorChar)))).ToList();
        if (answer.Refused.Length == 0 && answer.Files.Count == 0)
        {
            report.Add(check, false, "it answered with neither a refusal nor an edit: a changed kind is recorded in a file");
        }
        else if (missing.Count > 0)
        {
            report.Add(check, false, $"edits files that do not exist (paths are those before the move): {GoListMapper.List(missing)}");
        }
        else
        {
            var after = Snapshot(workspace, client.Info.Manifests);
            var untouched = before.Count == after.Count && before.All(entry => after.TryGetValue(entry.Key, out var text) && text == entry.Value);
            report.Add(check, untouched, "it wrote to the workspace: it must only answer with the new content, and let the host write it");
        }
    }

    /// <summary>The kind a generated project is asked to become: a lib is published as a package, an app becomes a lib, a package goes back to a lib.</summary>
    private static ProjectKind TargetOf(ProjectKind kind) => kind == ProjectKind.Lib ? ProjectKind.Package : ProjectKind.Lib;

    private static string Article(ProjectKind kind) => kind == ProjectKind.App ? "an app" : $"a {kind}";

    /// <summary>The content of every manifest of the workspace, by path.</summary>
    private static Dictionary<string, string> Snapshot(string workspace, IReadOnlyList<string> names) =>
        FindManifestsAlgorithm.Find(workspace, names)
            .ToDictionary(path => path, path => File.ReadAllText(Path.Combine(workspace, path.Replace('/', Path.DirectorySeparatorChar))));
}
