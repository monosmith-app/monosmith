using Monosmith.Launcher.VersionChoice;

namespace Monosmith.Launcher.VersionStore;

/// <summary>The folder that holds every smith the launchers and <c>smith update</c> installed.</summary>
internal static class VersionStoreRepository
{
    /// <summary>The versions folder: <c>&lt;home&gt;/versions</c>. SMITH_HOME moves it; default ~/.smith.</summary>
    /// <param name="env">The environment.</param>
    /// <returns>The versions folder.</returns>
    internal static string VersionsDir(IReadOnlyDictionary<string, string> env)
    {
        var home = env.TryGetValue("SMITH_HOME", out var value) && value.Length > 0
            ? value
            : Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".smith");
        return Path.Combine(home, "versions");
    }

    /// <summary>The file name of the smith program on a system.</summary>
    /// <param name="system">The archive's system name (windows, linux, darwin).</param>
    /// <returns>smith.exe on Windows, smith elsewhere.</returns>
    internal static string ProgramName(string system) => system == "windows" ? "smith.exe" : "smith";

    /// <summary>Where one version's program lives, whether it exists or not.</summary>
    /// <param name="env">The environment.</param>
    /// <param name="version">An exact version.</param>
    /// <param name="system">The archive's system name.</param>
    /// <returns>The program's path.</returns>
    internal static string ProgramPath(IReadOnlyDictionary<string, string> env, string version, string system) =>
        Path.Combine(VersionsDir(env), version, ProgramName(system));

    /// <summary>The newest installed version that has its program.</summary>
    /// <param name="env">The environment.</param>
    /// <param name="system">The archive's system name.</param>
    /// <returns>The version, or null when none is installed.</returns>
    internal static string? NewestInstalled(IReadOnlyDictionary<string, string> env, string system)
    {
        var dir = VersionsDir(env);
        if (!Directory.Exists(dir))
        {
            return null;
        }
        return Directory.EnumerateDirectories(dir)
            .Select(Path.GetFileName)
            .OfType<string>()
            .Where(v => ChooseVersionPolicy.Version().IsMatch(v) && File.Exists(ProgramPath(env, v, system)))
            .Order(Comparer<string>.Create(ChooseVersionPolicy.Compare))
            .LastOrDefault();
    }
}
