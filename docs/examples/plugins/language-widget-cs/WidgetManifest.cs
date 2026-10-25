using System.Text.RegularExpressions;
using MonoSmith.Plugin.Sdk;

namespace LanguageWidget;

/// <summary>The content of a widget.json. A member it does not have is empty.</summary>
/// <param name="Name">The widget's name.</param>
/// <param name="Version">An x.y.z version.</param>
/// <param name="Kind">app, lib or package; default lib.</param>
/// <param name="Description">What the widget is.</param>
/// <param name="License">The license it names.</param>
/// <param name="Uses">The names of the other widgets this one imports.</param>
/// <param name="Stack">The stack recorded when the widget was generated.</param>
public sealed record WidgetManifest(
    string Name,
    string Version,
    string Kind,
    string Description,
    string License,
    IReadOnlyList<string> Uses,
    string Stack)
{
    /// <summary>The file that makes a directory a widget project.</summary>
    public const string FileName = "widget.json";

    private static readonly Regex SemVer = new(@"^\d+\.\d+\.\d+$", RegexOptions.CultureInvariant);

    /// <summary>Reads a widget.json.</summary>
    /// <param name="path">The file.</param>
    /// <returns>The manifest.</returns>
    public static WidgetManifest Read(string path)
    {
        JsonMap map;
        try
        {
            map = JsonAny.Parse(File.ReadAllText(path)) as JsonMap ?? throw new JsonParseError("a widget.json is a JSON object");
        }
        catch (Exception error) when (error is IOException or JsonParseError)
        {
            throw new InvalidOperationException($"{path.Replace('\\', '/')}: {error.Message}", error);
        }

        string Text(string key) => map.TryGetValue(key, out var value) && value is JsonText text ? text.Value : string.Empty;
        var uses = map.TryGetValue("uses", out var list) && list is JsonList items ? items.OfType<JsonText>().Select(item => item.Value).ToList() : [];
        return new WidgetManifest(Text("name"), Text("version"), Text("kind"), Text("description"), Text("license"), uses, Text("stack"));
    }

    /// <summary>What the test task checks: a name and an x.y.z version.</summary>
    public void Validate()
    {
        if (Name.Length == 0)
        {
            throw new InvalidOperationException($"{FileName} has no name");
        }

        if (!SemVer.IsMatch(Version))
        {
            throw new InvalidOperationException($"{FileName}: version \"{Version}\" is not x.y.z");
        }
    }

    /// <summary>The widget.json with one string member set, every other key keeping its order and layout.</summary>
    /// <param name="path">The file.</param>
    /// <param name="key">The member.</param>
    /// <param name="value">Its new value.</param>
    /// <returns>The new text of the file.</returns>
    public static string WithField(string path, string key, string value)
    {
        var document = EditableJsonDocument.Parse(File.ReadAllText(path));
        document.SetString(key, value);
        return document.Render();
    }

    /// <summary>Rewrites one string member of a widget.json.</summary>
    /// <param name="path">The file.</param>
    /// <param name="key">The member.</param>
    /// <param name="value">Its new value.</param>
    public static void SetField(string path, string key, string value) =>
        File.WriteAllBytes(path, System.Text.Encoding.UTF8.GetBytes(WithField(path, key, value)));
}
