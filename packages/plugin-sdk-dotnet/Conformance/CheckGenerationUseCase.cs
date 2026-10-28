using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>The checks of a plugin that generates projects.</summary>
internal static class CheckGenerationUseCase
{
    /// <summary>The stack name generated projects are asked to record.</summary>
    public const string Stack = "conformance";

    private static readonly TimeSpan Long = TimeSpan.FromMinutes(20);

    /// <summary>
    /// Asks the plugin for one project of each kind and checks the files it reports exist. For a language plugin, also
    /// that its own <c>discover</c> returns what it made, with the kind and the stack it was asked for, tasks with
    /// commands, and that the dry runs of <c>version/write</c>, <c>remove</c>, <c>publish</c> and <c>published</c> answer,
    /// and, for a plugin that changes kinds, what <c>project/kind</c> edits.
    /// </summary>
    public static async Task CheckAsync(PluginClient client, string workspace, ConformanceReport report)
    {
        var language = PluginRoles.Effective(client.Info).Contains(Role.Language);
        var made = new List<GenerateParams>();
        var declared = client.Info.Kinds;
        foreach (var kind in new[] { ProjectKind.App, ProjectKind.Lib, ProjectKind.Package }.Where(each => declared.Count == 0 || declared.Contains(each)))
        {
            var name = $"c{kind}";
            var parameters = new GenerateParams
            {
                Kind = kind,
                Name = name,
                Directory = $"{ProjectKinds.DirectoryOf(kind)}/{name}",
                Stack = Stack,
                Workspace = new WorkspaceInfo { Name = "conformance", Scope = "@conformance", ModulePrefix = "example.com/conformance" },
            };
            IReadOnlyList<string> files;
            try
            {
                files = (await client.GenerateAsync(parameters, timeout: Long).ConfigureAwait(false)).Files;
            }
            catch (Exception error)
            {
                report.Add($"generates {Article(kind)}", false, error.Message);
                continue;
            }

            var missing = files.Where(file => !File.Exists(Path.Combine(workspace, file.Replace('/', Path.DirectorySeparatorChar)))
                && !Directory.Exists(Path.Combine(workspace, file.Replace('/', Path.DirectorySeparatorChar)))).ToList();
            report.Add(
                $"generates {Article(kind)}",
                files.Count > 0 && missing.Count == 0,
                $"files reported but not written: {GoListMapper.List(missing)} (of {GoListMapper.List(files)})");
            made.Add(parameters);
        }

        if (!language || made.Count == 0)
        {
            return;
        }

        IReadOnlyList<Project> projects;
        try
        {
            var manifests = FindManifestsAlgorithm.Find(workspace, client.Info.Manifests);
            projects = (await client.DiscoverAsync(new DiscoverParams { Manifests = manifests }, timeout: Long).ConfigureAwait(false)).Projects;
        }
        catch (Exception error)
        {
            report.Add("discovers what it generated", false, error.Message);
            return;
        }

        var found = projects.GroupBy(project => project.Root).ToDictionary(group => group.Key, group => group.Last());
        foreach (var parameters in made)
        {
            var kind = parameters.Kind;
            found.TryGetValue(parameters.Directory, out var project);
            report.Add($"discovers the generated {kind}", project is not null, $"no project at {parameters.Directory} among {projects.Count} discovered");
            if (project is null)
            {
                continue;
            }

            report.Add($"the generated {kind} keeps its kind", project.Kind == kind, $"discovered as {GoListMapper.Quote(project.Kind.Value)}");
            report.Add(
                $"the generated {kind} records its stack",
                project.Stack == Stack,
                $"discovered with stack {GoListMapper.Quote(project.Stack)}; generate must record params.stack in the manifest");
            var empty = project.Tasks.Where(task => task.Value.Command.Count == 0).Select(task => task.Key).Order(StringComparer.Ordinal).ToList();
            report.Add($"the generated {kind}'s tasks have commands", empty.Count == 0, $"tasks without a command: {GoListMapper.List(empty)}");
            var sixty = TimeSpan.FromSeconds(60);
            if (client.Supports(ProtocolMethods.VersionWrite))
            {
                await AttemptAsync(report, $"writes the {kind}'s version (dry run)", () => client.WriteVersionAsync(new VersionParams { Project = project, Version = "1.2.3", DryRun = true }, timeout: sixty)).ConfigureAwait(false);
            }

            if (client.Supports(ProtocolMethods.Remove))
            {
                await AttemptAsync(report, $"would remove the {kind} (dry run)", () => client.RemoveAsync(new RemoveParams { Project = project, DryRun = true }, timeout: sixty)).ConfigureAwait(false);
            }

            if (client.Supports(ProtocolMethods.ProjectKind))
            {
                await CheckKindChangeUseCase.CheckAsync(client, workspace, project, projects, report).ConfigureAwait(false);
            }

            if (kind == ProjectKind.Package && client.Supports(ProtocolMethods.Publish))
            {
                await AttemptAsync(report, "publishes the package (dry run)", () => client.PublishAsync(new PublishParams { Project = project, Version = "1.2.3", DryRun = true }, timeout: TimeSpan.FromSeconds(300))).ConfigureAwait(false);
            }

            if (kind == ProjectKind.Package && client.Supports(ProtocolMethods.Published))
            {
                await AttemptAsync(
                    report,
                    "says what the registry has of the package",
                    () => client.PublishedAsync(new PublishedParams { Project = project }, timeout: sixty),
                    " (an unreachable registry is a result, not an error)").ConfigureAwait(false);
            }
        }
    }

    private static string Article(ProjectKind kind) => kind == ProjectKind.App ? "an app" : $"a {kind}";

    private static async Task AttemptAsync(ConformanceReport report, string check, Func<Task> call, string hint = "")
    {
        try
        {
            await call().ConfigureAwait(false);
            report.Add(check, true, string.Empty);
        }
        catch (Exception error)
        {
            report.Add(check, false, error.Message + hint);
        }
    }
}
