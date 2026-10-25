using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Recording the toolchain on discovered projects.</summary>
public static class ToolchainRecording
{
    /// <summary>
    /// Records the toolchain on every discovered project. The probe (<c>dotnet --version</c> and the like) runs only
    /// when there is a project, and once.
    /// </summary>
    /// <param name="result">What <c>discover</c> found.</param>
    /// <param name="probe">Asks the toolchain for its version.</param>
    /// <returns>The result with the toolchain set on its projects; the one given when there is none.</returns>
    public static DiscoverResult Set(DiscoverResult result, Func<string> probe)
    {
        ArgumentNullException.ThrowIfNull(result);
        ArgumentNullException.ThrowIfNull(probe);
        if (result.Projects.Count == 0)
        {
            return result;
        }

        var toolchain = probe();
        return result with { Projects = [.. result.Projects.Select(project => project with { Toolchain = toolchain })] };
    }
}
