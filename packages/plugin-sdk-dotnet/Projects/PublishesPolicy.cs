using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Whether a release publishes a project.</summary>
public static class ProjectPublishing
{
    /// <summary>
    /// Whether a release publishes the project to a registry: a package, unless <c>smith.toml</c> says <c>registry = "none"</c>.
    /// Plugins list the credentials a release needs only for such projects.
    /// </summary>
    /// <param name="project">The project.</param>
    /// <returns>Whether a release publishes it.</returns>
    public static bool Publishes(Project project)
    {
        ArgumentNullException.ThrowIfNull(project);
        return project.Kind == ProjectKind.Package && project.Registries?.NoPublish != true;
    }
}
