using System.Text;
using System.Text.RegularExpressions;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Telling whether a file is one of a plugin's manifests.</summary>
public static class ManifestMatching
{
    /// <summary>
    /// Whether a file's base name is one of a plugin's manifests: equal to a name, or matching a pattern
    /// (<c>*.csproj</c>). A pattern is Go's <c>path.Match</c>: <c>*</c> matches any run of characters, <c>?</c> one,
    /// <c>[a-z]</c> and <c>[^a-z]</c> a class, <c>\</c> escapes. A malformed pattern matches nothing.
    /// </summary>
    /// <param name="manifests">The plugin's manifest names and patterns.</param>
    /// <param name="baseName">The file's base name.</param>
    /// <returns>Whether the file is a manifest.</returns>
    public static bool IsManifest(IEnumerable<string> manifests, string baseName)
    {
        ArgumentNullException.ThrowIfNull(manifests);
        ArgumentNullException.ThrowIfNull(baseName);
        foreach (var manifest in manifests)
        {
            if (manifest == baseName)
            {
                return true;
            }

            if (ToRegex(manifest) is { } expression && expression.IsMatch(baseName))
            {
                return true;
            }
        }

        return false;
    }

    private static Regex? ToRegex(string pattern)
    {
        var source = new StringBuilder("^(?:");
        for (var i = 0; i < pattern.Length; i++)
        {
            var c = pattern[i];
            switch (c)
            {
                case '\\':
                    if (i + 1 >= pattern.Length)
                    {
                        return null;
                    }

                    source.Append(Regex.Escape(pattern[++i].ToString()));
                    break;
                case '[':
                    var close = pattern.IndexOf(']', i + 1);
                    if (close <= i + 1)
                    {
                        return null;
                    }

                    source.Append('[').Append(pattern[(i + 1)..close].Replace("\\", "\\\\", StringComparison.Ordinal).Replace("[", "\\[", StringComparison.Ordinal)).Append(']');
                    i = close;
                    break;
                case '*':
                    source.Append("[^/]*");
                    break;
                case '?':
                    source.Append("[^/]");
                    break;
                default:
                    source.Append(Regex.Escape(c.ToString()));
                    break;
            }
        }

        source.Append(")\\z");
        try
        {
            return new Regex(source.ToString(), RegexOptions.Singleline | RegexOptions.CultureInvariant, TimeSpan.FromSeconds(1));
        }
        catch (ArgumentException)
        {
            return null;
        }
    }
}
