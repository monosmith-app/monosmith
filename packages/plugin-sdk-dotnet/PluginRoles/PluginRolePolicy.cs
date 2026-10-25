using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>The roles a plugin can play in a project's make-up: known roles, their cardinality and the defaults of a plugin kind.</summary>
public static class PluginRoles
{
    /// <summary>Every known role, in the order a project's make-up lists them.</summary>
    public static IReadOnlyList<Role> All { get; } = [.. ProtocolRoles.All.Select(row => row.Name)];

    /// <summary>Whether the role is one of <see cref="All"/>.</summary>
    /// <param name="role">The role.</param>
    /// <returns>Whether it is known.</returns>
    public static bool IsKnown(Role role) => All.Contains(role);

    /// <summary>
    /// Whether a project has at most one plugin at this level (language, scaffold). A toolchain is exclusive per slot
    /// instead.
    /// </summary>
    /// <param name="role">The role.</param>
    /// <returns>Whether the level is exclusive.</returns>
    public static bool IsExclusive(Role role) => CardinalityOf(role) == "exclusive";

    /// <summary>Whether any number of plugins can share the level, each only adding (structure, deploy).</summary>
    /// <param name="role">The role.</param>
    /// <returns>Whether the level is additive.</returns>
    public static bool IsAdditive(Role role) => CardinalityOf(role) == "additive";

    /// <summary>
    /// The roles of a plugin that declares none. A language plugin brings its own default toolchain and scaffold; any
    /// other kind of plugin is a structure rule.
    /// </summary>
    /// <param name="kind">The plugin's kind.</param>
    /// <returns>The default roles.</returns>
    public static IReadOnlyList<Role> DefaultFor(PluginKind kind)
    {
        var wanted = kind == PluginKind.Language ? "language" : "convention";
        return [.. ProtocolRoles.All.Where(row => row.DefaultFor.Contains(wanted)).Select(row => row.Name)];
    }

    /// <summary>The plugin's roles (<c>roles</c> of its <c>initialize</c> result), or the default of its kind.</summary>
    /// <param name="info">What the plugin answered to <c>initialize</c>.</param>
    /// <returns>The roles the plugin plays.</returns>
    public static IReadOnlyList<Role> Effective(InitializeResult info)
    {
        ArgumentNullException.ThrowIfNull(info);
        return info.Roles.Count > 0 ? info.Roles : DefaultFor(info.Kind);
    }

    private static string? CardinalityOf(Role role) =>
        ProtocolRoles.All.FirstOrDefault(row => row.Name == role)?.Cardinality;
}
