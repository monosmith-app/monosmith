using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;

namespace LanguageWidget;

/// <summary>
/// The plugin. Every interface it implements beyond <see cref="IPlugin"/> is a capability the SDK advertises:
/// <c>discover</c>, <c>generate</c>, <c>version</c> and <c>doctor</c>.
/// </summary>
public sealed class WidgetPlugin : IPlugin, IDiscoverer, IGenerator, IVersioner, IDoctorContributor
{
    /// <summary>
    /// A new project's starter files. The templates use [[ ]], so a body can contain {{ }} untouched.
    /// </summary>
    public static readonly IReadOnlyDictionary<string, string> GeneratedFiles = new Dictionary<string, string>
    {
        ["[[.Dir]]/widget.json"] = """
            {
              "name": "[[.Name]]",
              "version": "0.0.0",
              "kind": "[[.Kind]]",
              "description": "A widget of the [[.Workspace]] workspace.",
              "license": "MIT"[[if .Stack]],
              "stack": "[[.Stack]]"[[end]]
            }

            """.ReplaceLineEndings("\n"),
        ["[[.Dir]]/README.md"] = "# [[.Name]]\n\nA widget. Edit widget.json.\n",
    };

    private string _workspaceRoot = string.Empty;

    /// <summary>
    /// The command the tasks run: this program. A plugin that wraps a real tool would name the tool here instead.
    /// </summary>
    private static IReadOnlyList<string> Self
    {
        get
        {
            // Run as an assembly (`dotnet LanguageWidget.dll`) the program is dotnet and the assembly; published as a single file
            // there is no assembly beside it and the program is the executable.
            var assembly = Path.Combine(AppContext.BaseDirectory, "LanguageWidget.dll");
            return File.Exists(assembly) ? ["dotnet", assembly] : [Environment.ProcessPath ?? "LanguageWidget"];
        }
    }

    /// <inheritdoc />
    public PluginDescription Describe(InitializeParams parameters)
    {
        _workspaceRoot = parameters.WorkspaceRoot;
        return new PluginDescription
        {
            Plugin = new Peer { Name = "widget", Version = "0.1.0" },
            Kind = PluginKind.Language,

            // A language plugin that also writes new projects plays the scaffold level too; it does not need to be
            // chosen as a toolchain.
            Roles = [Role.Language, Role.Scaffold],
            Languages = ["widget"],
            Manifests = [WidgetManifest.FileName],
        };
    }

    /// <summary>One project per widget.json that is ours.</summary>
    public Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation)
    {
        var projects = new List<Project>();
        foreach (var relative in parameters.Manifests) // Workspace-relative, forward slashes.
        {
            var manifest = WidgetManifest.Read(ManifestPath(relative));
            if (manifest.Name.Length == 0)
            {
                continue; // Not a widget manifest: leave it out.
            }

            var root = relative.Contains('/', StringComparison.Ordinal) ? relative[..relative.LastIndexOf('/')] : string.Empty;
            projects.Add(new Project
            {
                Name = manifest.Name,
                Root = root,
                Language = "widget",
                Kind = manifest.Kind.Length > 0 ? new ProjectKind(manifest.Kind) : ProjectKind.Lib,
                Version = manifest.Version,
                Description = manifest.Description,
                Stack = manifest.Stack,
                Exports = [manifest.Name],
                Imports = manifest.Uses, // The host keeps the ones another project exports.
                Tasks = new Dictionary<string, ProjectTask>
                {
                    ["build"] = ProjectTasks.Build(
                        root,
                        output => [.. Self, "build", output],
                        inputs: [WidgetManifest.FileName],
                        description: "write widget.txt into the build directory"),
                    ["test"] = ProjectTasks.Create(
                        [.. Self, "check"],
                        inputs: [WidgetManifest.FileName],
                        cache: true,
                        description: "check the manifest: a name and an x.y.z version"),
                },
            });
        }

        return Task.FromResult(new DiscoverResult { Projects = projects });
    }

    /// <summary>A new project's starter files.</summary>
    public Task<GenerateResult> GenerateAsync(GenerateParams parameters, CancellationToken cancellation) =>
        Task.FromResult(Scaffold.Project(_workspaceRoot, parameters, GeneratedFiles));

    /// <summary>Records the release version.</summary>
    public Task<VersionResult> WriteVersionAsync(VersionParams parameters, CancellationToken cancellation)
    {
        var file = Join(parameters.Project.Root, WidgetManifest.FileName);
        if (!parameters.DryRun)
        {
            WidgetManifest.SetField(ManifestPath(file), "version", parameters.Version);
        }

        return Task.FromResult(new VersionResult { Files = [file] });
    }

    /// <summary>Every failed finding names its remedy, and a fix when the plugin can write one.</summary>
    public Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation)
    {
        var findings = new List<Finding>();
        foreach (var project in parameters.Projects)
        {
            var file = Join(project.Root, WidgetManifest.FileName);
            var manifest = WidgetManifest.Read(ManifestPath(file));
            var check = $"{project.Name} names a license";
            if (manifest.License.Length > 0)
            {
                findings.Add(new Finding { Check = check, Ok = true });
                continue;
            }

            var document = EditableJsonDocument.Parse(File.ReadAllText(ManifestPath(file)));
            document.SetString("license", "MIT");
            findings.Add(new Finding
            {
                Check = check,
                Ok = false,
                Detail = $"{file} has no license field",
                Remedy = $"add \"license\": \"MIT\" (or your license) to {file}",

                // Not safe: it picks a license for the team, so smith asks first.
                Fix = new Fix
                {
                    Description = "add an MIT license to widget.json",
                    Safe = false,
                    AppliesTo = [project.Name],
                    Files = [new FileEdit { Path = file, Content = document.Render() }],
                },
            });
        }

        return Task.FromResult(new DoctorResult { Findings = findings });
    }

    private static string Join(string directory, string name) => directory.Length == 0 ? name : $"{directory.TrimEnd('/')}/{name}";

    private string ManifestPath(string relative) => Path.Combine(_workspaceRoot, relative.Replace('/', Path.DirectorySeparatorChar));
}
