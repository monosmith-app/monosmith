using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;

namespace StructureLayout;

/// <summary>The structure plugin: no junk-drawer folders in widget projects, and a README for every project.</summary>
/// <remarks>It implements <c>contribute</c> (tasks) and <c>doctor</c> (findings): the SDK advertises exactly those.</remarks>
public sealed class LayoutPlugin : IPlugin, IContributor, IDoctorContributor
{
    private static readonly HashSet<string> JunkDrawers = ["util", "utils", "helper", "helpers", "common", "shared", "misc"];
    private static readonly HashSet<string> Skipped = ["dist", "node_modules", ".git", "__pycache__", "bin", "obj"];

    private string _workspaceRoot = string.Empty;

    /// <summary>The program the layout-check task runs: this one.</summary>
    private static IReadOnlyList<string> Self
    {
        get
        {
            var location = typeof(LayoutPlugin).Assembly.Location;
            return location.Length > 0 ? ["dotnet", location] : [Environment.ProcessPath ?? "StructureLayout"];
        }
    }

    /// <summary>The folders below <paramref name="directory"/> named for nothing, as slash-separated paths relative to it.</summary>
    /// <param name="directory">Where to look.</param>
    /// <param name="prefix">The path below <paramref name="directory"/> the search is at.</param>
    /// <returns>The junk drawers.</returns>
    public static IReadOnlyList<string> FindJunkDrawers(string directory, string prefix = "")
    {
        var found = new List<string>();
        var current = prefix.Length == 0 ? directory : Path.Combine(directory, prefix.Replace('/', Path.DirectorySeparatorChar));
        foreach (var entry in new DirectoryInfo(current).GetDirectories().OrderBy(entry => entry.Name, StringComparer.Ordinal))
        {
            if (Skipped.Contains(entry.Name))
            {
                continue;
            }

            var relative = prefix.Length == 0 ? entry.Name : $"{prefix}/{entry.Name}";
            if (JunkDrawers.Contains(entry.Name.ToLowerInvariant()))
            {
                found.Add(relative);
            }

            found.AddRange(FindJunkDrawers(directory, relative));
        }

        return found;
    }

    /// <inheritdoc />
    public PluginDescription Describe(InitializeParams parameters)
    {
        _workspaceRoot = parameters.WorkspaceRoot;
        return new PluginDescription
        {
            Plugin = new Peer { Name = "widget-layout", Version = "0.1.0" },
            Kind = PluginKind.Convention,
            Roles = [Role.Structure],

            // The languages it applies to. A convention that names none applies to no project at all.
            Languages = ["widget"],
        };
    }

    /// <summary>A layout-check task per project, and the project's lint (when it has one) waits for it.</summary>
    public Task<ContributeResult> ContributeAsync(ContributeParams parameters, CancellationToken cancellation) =>
        Task.FromResult(Contribute.To(parameters.Projects, project => new Contribution
        {
            Tasks = new Dictionary<string, ProjectTask>
            {
                // No inputs listed: every file of the project is one, so a new or renamed folder is a cache miss.
                ["layout-check"] = ProjectTasks.Create(
                    [.. Self, "layout-check"],
                    cache: true,
                    description: "fail on folders named util, helpers, common, shared or misc"),
            },

            // DependsOn only touches a task the project has once every earlier plugin has contributed; for a project
            // without lint it is ignored. Do not decide from project.Tasks here: it holds only what the language
            // plugin discovered, not what a toolchain plugin adds.
            DependsOn = new Dictionary<string, IReadOnlyList<string>> { ["lint"] = ["layout-check"] },
        }));

    /// <summary>Every project needs a README.md.</summary>
    public Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation)
    {
        var findings = new List<Finding>();
        foreach (var project in parameters.Projects)
        {
            var check = $"{project.Name} has a README.md";
            var readme = Path.Combine(_workspaceRoot, project.Root.Replace('/', Path.DirectorySeparatorChar), "README.md");
            if (File.Exists(readme))
            {
                findings.Add(new Finding { Check = check, Ok = true });
                continue;
            }

            findings.Add(new Finding
            {
                Check = check,
                Ok = false,
                Detail = $"{project.Root}/README.md does not exist",
                Remedy = $"write {project.Root}/README.md: what the project is for",
                Warning = true, // Reported, but does not fail smith doctor.
            });
        }

        return Task.FromResult(new DoctorResult { Findings = findings });
    }
}
