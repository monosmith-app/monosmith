using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>The checks of <c>smith plugin check</c>, to run in a plugin's own tests.</summary>
public static class ConformanceChecks
{
    /// <summary>The stack name generated projects are asked to record.</summary>
    public const string Stack = CheckGenerationUseCase.Stack;

    private static readonly TimeSpan CallTimeout = TimeSpan.FromMinutes(20);

    /// <summary>Runs the checks of <c>smith plugin check</c> against a connected plugin.</summary>
    /// <remarks>
    /// <para>
    /// It names itself, has a known kind and known roles, its roles match what it implements, it names the slots it
    /// fills, its references and stacks are well-formed, every advertised method answers a request with nothing in it,
    /// requests sent together are answered as the declared <c>concurrency</c> allows, a plugin that advertises
    /// <c>cancel</c> stops a cancelled request, and, with <paramref name="workspace"/>, the projects it generates are
    /// found again by its own <c>discover</c>. A port of the Go SDK's <c>conformance.Check</c>, with the same checks
    /// under the same names.
    /// </para>
    /// <para>
    /// <paramref name="workspace"/> is an empty directory the plugin was started in (its <c>workspaceRoot</c>). When set,
    /// a plugin that generates projects is asked to generate an app, a lib and a package there, which a language plugin
    /// must then discover as generated. Generation may install dependencies, so it is opt-in.
    /// </para>
    /// <para>
    /// <c>smith plugin check</c> adds one more, "shuts down when asked", after closing the plugin: call
    /// <see cref="PluginClient.CloseAsync"/> and add it to the report if you want it.
    /// </para>
    /// </remarks>
    /// <param name="client">The connected plugin.</param>
    /// <param name="workspace">An empty directory the plugin generates into, or <c>null</c> to skip generation.</param>
    /// <returns>The report.</returns>
    public static async Task<ConformanceReport> CheckPluginAsync(PluginClient client, string? workspace = null)
    {
        ArgumentNullException.ThrowIfNull(client);
        var info = client.Info;
        var report = new ConformanceReport(info.Plugin);
        report.Add(
            "names itself",
            info.Plugin.Name.Length > 0 && info.Plugin.Version.Length > 0,
            $"initialize returned name {GoListMapper.Quote(info.Plugin.Name)}, version {GoListMapper.Quote(info.Plugin.Version)}; both are required");
        report.Add(
            "has a known kind",
            info.Kind == PluginKind.Language || info.Kind == PluginKind.Convention,
            $"kind {GoListMapper.Quote(info.Kind.Value)} is neither language nor convention");
        var roles = PluginRoles.Effective(info);
        foreach (var role in roles)
        {
            report.Add($"plays a known role: {role}", PluginRoles.IsKnown(role), $"roles are {GoListMapper.List(PluginRoles.All.Select(known => (object)known))}");
        }

        var capabilities = info.Capabilities;
        if (roles.Contains(Role.Language))
        {
            report.Add("a language plugin discovers projects", capabilities.Discover && info.Manifests.Count > 0, "it must implement discover and name its manifest files");
        }

        if (roles.Contains(Role.Scaffold))
        {
            report.Add("a scaffold generates projects", capabilities.Generate, "it must implement generate");
        }

        foreach (var role in new[] { Role.Toolchain, Role.Structure, Role.Deploy })
        {
            if (roles.Contains(role) && !roles.Contains(Role.Language))
            {
                var doesSomething = capabilities.Contribute || capabilities.Generate || capabilities.Ci || capabilities.Doctor;
                report.Add($"a {role} plugin does something", doesSomething, "it implements none of contribute, generate, ci and doctor");
            }
        }

        if (info.Provides.Count > 0)
        {
            report.Add("only a toolchain provides tasks", roles.Contains(Role.Toolchain), $"it provides {GoListMapper.List(info.Provides)} without the toolchain role");
        }

        if (info.Kind == PluginKind.Convention && roles.Contains(Role.Toolchain))
        {
            report.Add("a toolchain names the slots it fills", info.Provides.Count > 0, "a toolchain plugin must list the tasks it provides");
        }

        foreach (var reference in info.Requires)
        {
            report.Add($"requires {PluginRefs.Describe(reference)}: a known reference", IsValidRef(reference), "a reference sets a known level, a name or a slot");
        }

        foreach (var reference in info.Conflicts)
        {
            report.Add(
                $"conflicts with {PluginRefs.Describe(reference)}: a known reference",
                IsValidRef(reference) && reference.Name != info.Plugin.Name,
                "a reference sets a known level, a name or a slot, and is not the plugin itself");
        }

        var names = new HashSet<string>(StringComparer.Ordinal);
        foreach (var stack in info.Stacks)
        {
            report.Add(
                $"stack {stack.Name} is complete",
                stack.Name.Length > 0 && stack.Language.Length > 0 && !names.Contains(stack.Name),
                "a stack needs a unique name and a language plugin");
            names.Add(stack.Name);
        }

        await CheckSchedulingUseCase.CheckAsync(client, report).ConfigureAwait(false);

        await AnswersAsync(client, report, "discover answers with no manifests", ProtocolMethods.Discover, () => client.DiscoverAsync(new DiscoverParams { Manifests = [] }, timeout: CallTimeout)).ConfigureAwait(false);
        await AnswersAsync(client, report, "contribute answers with no projects", ProtocolMethods.Contribute, () => client.ContributeAsync(new ContributeParams { Projects = [] }, timeout: CallTimeout)).ConfigureAwait(false);
        await AnswersAsync(client, report, "ci answers with no projects", ProtocolMethods.Ci, () => client.CiAsync(new CIParams { Projects = [] }, timeout: CallTimeout)).ConfigureAwait(false);
        await AnswersAsync(client, report, "doctor answers with no projects", ProtocolMethods.Doctor, () => client.DoctorAsync(new DoctorParams { Projects = [] }, timeout: CallTimeout)).ConfigureAwait(false);

        if (workspace is not null && capabilities.Generate)
        {
            await CheckGenerationUseCase.CheckAsync(client, workspace, report).ConfigureAwait(false);
        }

        return report;
    }

    private static async Task AnswersAsync(PluginClient client, ConformanceReport report, string check, string method, Func<Task> call)
    {
        if (!client.Supports(method))
        {
            return;
        }

        try
        {
            await call().ConfigureAwait(false);
            report.Add(check, true, string.Empty);
        }
        catch (Exception error)
        {
            report.Add(check, false, error.Message);
        }
    }

    /// <summary>Whether a reference names something: a known level, a name or a slot.</summary>
    private static bool IsValidRef(PluginRef reference) =>
        !PluginRefs.IsEmpty(reference) && (reference.Level.IsEmpty || PluginRoles.IsKnown(reference.Level));
}
