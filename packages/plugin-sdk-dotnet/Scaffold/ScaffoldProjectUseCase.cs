using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>The answer to <c>generate</c> for a scaffold.</summary>
public static class Scaffold
{
    /// <summary>
    /// The answer to <c>generate</c> for a scaffold: renders the templates into the new project, never overwriting a
    /// file, and reports what was written.
    /// </summary>
    /// <remarks>
    /// A language plugin that scaffolds must record <see cref="GenerateParams.Stack"/> in the manifest it writes
    /// (<c>[[if .Stack]]</c> in a template), so discovery reports it back.
    /// </remarks>
    /// <param name="workspaceRoot">The workspace root from <c>initialize</c>: the project's directory is relative to it.</param>
    /// <param name="parameters">The params of <c>generate</c>.</param>
    /// <param name="files">Target paths and template text, with <c>[[ ]]</c> actions in both (<c>"[[.Dir]]/widget.json"</c>).</param>
    /// <param name="data">More fields for the templates, next to the ones <see cref="ScaffoldData.From"/> lists; they win over them.</param>
    /// <param name="notes">Lines printed to the user after generation.</param>
    /// <returns>What was written.</returns>
    public static GenerateResult Project(
        string workspaceRoot,
        GenerateParams parameters,
        IReadOnlyDictionary<string, string> files,
        JsonMap? data = null,
        IEnumerable<string>? notes = null)
    {
        var templateData = ScaffoldData.From(parameters);
        if (data is not null)
        {
            foreach (var (name, value) in data)
            {
                templateData[name] = value;
            }
        }

        var written = Templates.Write(workspaceRoot, files, templateData);
        return new GenerateResult { Files = written, Notes = [.. notes ?? []] };
    }
}
