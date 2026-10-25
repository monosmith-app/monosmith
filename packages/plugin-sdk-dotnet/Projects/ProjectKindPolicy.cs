using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>A project's kind, and where kinds conventionally live.</summary>
public static class ProjectKinds
{
    /// <summary>
    /// A project's kind when its manifest may declare none: the declared kind when it is valid; otherwise the directory
    /// convention (<c>apps/</c> is an app, <c>libs/</c> a lib, <c>packages/</c> a package); and finally <paramref name="fallback"/>.
    /// </summary>
    /// <param name="declared">The kind the manifest declares, if any.</param>
    /// <param name="root">The project's root, workspace-relative.</param>
    /// <param name="fallback">The kind when nothing else says.</param>
    /// <returns>The kind.</returns>
    public static ProjectKind Resolve(string? declared, string root, ProjectKind fallback)
    {
        ArgumentNullException.ThrowIfNull(root);
        var kind = (declared ?? string.Empty).Trim();
        if (kind is "app" or "lib" or "package")
        {
            return new ProjectKind(kind);
        }

        return root.Split('/', 2)[0] switch
        {
            "apps" => ProjectKind.App,
            "libs" => ProjectKind.Lib,
            "packages" => ProjectKind.Package,
            _ => fallback,
        };
    }

    /// <summary>The conventional parent directory of a kind: <c>apps</c>, <c>libs</c> or <c>packages</c>.</summary>
    /// <param name="kind">The kind.</param>
    /// <returns>The directory.</returns>
    public static string DirectoryOf(ProjectKind kind) =>
        kind == ProjectKind.App ? "apps" : kind == ProjectKind.Lib ? "libs" : "packages";
}
