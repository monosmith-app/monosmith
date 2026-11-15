using System.Globalization;
using System.Text.RegularExpressions;

namespace Monosmith.Launcher.VersionChoice;

/// <summary>The version a run uses: SMITH_VERSION, the workspace's pin, or the newest.</summary>
internal static partial class ChooseVersionPolicy
{
    /// <summary>A version smith publishes: 2.50.0, 3.0.0-beta.1.</summary>
    [GeneratedRegex(@"^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$")]
    internal static partial Regex Version();

    [GeneratedRegex(@"^\[([^\]]+)\]$")]
    private static partial Regex Header();

    [GeneratedRegex("^smith\\s*=\\s*[\"']([^\"']+)[\"']$")]
    private static partial Regex Pin();

    /// <summary>The <c>[workspace] smith</c> pin in smith.toml text.</summary>
    /// <param name="text">The text of smith.toml.</param>
    /// <returns>The version, or null for none or <c>latest</c>.</returns>
    internal static string? PinIn(string text)
    {
        var table = "";
        foreach (var raw in text.Split('\n'))
        {
            var hash = raw.IndexOf('#', StringComparison.Ordinal);
            var line = (hash < 0 ? raw : raw[..hash]).Trim();
            var header = Header().Match(line);
            if (header.Success)
            {
                table = header.Groups[1].Value.Trim();
                continue;
            }
            var pin = Pin().Match(line);
            if (table == "workspace" && pin.Success)
            {
                return Version().IsMatch(pin.Groups[1].Value) ? pin.Groups[1].Value : null;
            }
        }
        return null;
    }

    /// <summary>The pin of the workspace a folder is in: the nearest smith.toml above it.</summary>
    /// <param name="start">The folder to start from (the current folder).</param>
    /// <returns>The pinned version, or null when no smith.toml pins one.</returns>
    internal static string? WorkspacePin(string start)
    {
        for (var dir = new DirectoryInfo(start); dir is not null; dir = dir.Parent)
        {
            var file = Path.Combine(dir.FullName, "smith.toml");
            if (File.Exists(file))
            {
                return PinIn(File.ReadAllText(file));
            }
        }
        return null;
    }

    /// <summary>The version a run uses: SMITH_VERSION, else the workspace's pin, else null (the newest).</summary>
    /// <param name="env">The environment.</param>
    /// <param name="cwd">The current folder.</param>
    /// <returns>An exact version, or null for the newest.</returns>
    /// <exception cref="ArgumentException">SMITH_VERSION is set to something that is not a version.</exception>
    internal static string? Choose(IReadOnlyDictionary<string, string> env, string cwd)
    {
        var wanted = env.TryGetValue("SMITH_VERSION", out var value) ? value.Trim() : "";
        if (wanted.Length > 0 && wanted != "latest")
        {
            return Version().IsMatch(wanted) ? wanted : throw new ArgumentException($"SMITH_VERSION is \"{wanted}\": use a version such as 2.50.0, or latest");
        }
        return wanted == "latest" ? null : WorkspacePin(cwd);
    }

    /// <summary>Compares two versions by their numbers; a prerelease sorts before its release.</summary>
    /// <param name="a">A version.</param>
    /// <param name="b">Another version.</param>
    /// <returns>Negative when a is older, positive when newer, 0 when equal.</returns>
    internal static int Compare(string a, string b)
    {
        var (mainA, preA) = Split(a);
        var (mainB, preB) = Split(b);
        for (var i = 0; i < 3; i++)
        {
            var x = int.Parse(mainA[i], CultureInfo.InvariantCulture);
            var y = int.Parse(mainB[i], CultureInfo.InvariantCulture);
            if (x != y)
            {
                return x - y;
            }
        }
        if (preA == preB)
        {
            return 0;
        }
        if (preA.Length == 0)
        {
            return 1;
        }
        return preB.Length == 0 ? -1 : string.CompareOrdinal(preA, preB);
    }

    private static (string[] Main, string Pre) Split(string version)
    {
        var dash = version.IndexOf('-', StringComparison.Ordinal);
        var main = dash < 0 ? version : version[..dash];
        return (main.Split('.'), dash < 0 ? "" : version[(dash + 1)..]);
    }
}
