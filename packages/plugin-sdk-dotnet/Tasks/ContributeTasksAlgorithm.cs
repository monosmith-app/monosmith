using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>The answer to <c>contribute</c> when every project gets the same kind of addition.</summary>
public static class Contribute
{
    /// <summary>
    /// Calls <paramref name="contribution"/> for each project and keys the result by project name. Return <c>null</c>
    /// for a project the plugin has nothing to add to; it is left out.
    /// </summary>
    /// <example>
    /// <code>
    /// public Task&lt;ContributeResult&gt; ContributeAsync(ContributeParams parameters, CancellationToken cancellation) =>
    ///     Task.FromResult(Contribute.To(parameters.Projects, project => new Contribution
    ///     {
    ///         Tasks = new Dictionary&lt;string, ProjectTask&gt; { ["layout-check"] = ProjectTasks.Create(["dotnet", "check.dll"], cache: true) },
    ///         DependsOn = new Dictionary&lt;string, IReadOnlyList&lt;string&gt;&gt; { ["lint"] = ["layout-check"] },
    ///     }));
    /// </code>
    /// </example>
    /// <param name="projects">The projects the host asked about.</param>
    /// <param name="contribution">What to add to a project, or <c>null</c>.</param>
    /// <returns>The result to answer.</returns>
    public static ContributeResult To(IEnumerable<Project> projects, Func<Project, Contribution?> contribution)
    {
        ArgumentNullException.ThrowIfNull(projects);
        ArgumentNullException.ThrowIfNull(contribution);
        var result = new Dictionary<string, Contribution>();
        foreach (var project in projects)
        {
            if (contribution(project) is { } added)
            {
                result[project.Name] = added;
            }
        }

        return new ContributeResult { Projects = result };
    }
}
