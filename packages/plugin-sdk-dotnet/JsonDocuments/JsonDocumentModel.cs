using System.Globalization;
using System.Text;

namespace MonoSmith.Plugin.Sdk;

/// <summary>A JSON object that keeps its key order and the text of everything you do not touch.</summary>
/// <remarks>
/// <para>
/// So editing a <c>package.json</c> changes only what was edited: set a member and it keeps its place, add one and it
/// goes last, delete one and the others stay where they were. It is a port of the Go SDK's <c>jsondoc.Document</c>.
/// </para>
/// <code>
/// var document = EditableJsonDocument.Parse(File.ReadAllText(path));
/// document.SetString("version", "1.2.3");
/// File.WriteAllText(path, document.Render());
/// </code>
/// </remarks>
public sealed class EditableJsonDocument
{
    private readonly List<string> _keys = [];
    private readonly Dictionary<string, object> _entries = new(StringComparer.Ordinal);

    /// <summary>The keys, in order.</summary>
    public IReadOnlyList<string> Keys => _keys;

    /// <summary>How many members the object has.</summary>
    public int Count => _keys.Count;

    /// <summary>Reads a JSON object.</summary>
    /// <param name="text">The text of the object.</param>
    /// <returns>The document.</returns>
    /// <exception cref="JsonDocumentError">The text is anything else.</exception>
    public static EditableJsonDocument Parse(string text)
    {
        ArgumentNullException.ThrowIfNull(text);
        return FromNode(ScanJsonObjectAlgorithm.Scan(text));
    }

    /// <summary>Whether the key exists.</summary>
    /// <param name="key">The key.</param>
    /// <returns>Whether it exists.</returns>
    public bool Has(string key) => _entries.ContainsKey(key);

    /// <summary>The object under a key, editable in place.</summary>
    /// <param name="key">The key.</param>
    /// <param name="create">Adds an empty one (last) when there is none.</param>
    /// <returns>The object; <c>null</c> for a key that holds something else, or none when <paramref name="create"/> is false.</returns>
    public EditableJsonDocument? Object(string key, bool create = false)
    {
        if (_entries.TryGetValue(key, out var entry))
        {
            return entry as EditableJsonDocument;
        }

        if (!create)
        {
            return null;
        }

        var child = new EditableJsonDocument();
        Put(key, child);
        return child;
    }

    /// <summary>A string value.</summary>
    /// <param name="key">The key.</param>
    /// <returns>The string; <c>null</c> when the key is missing or does not hold a string.</returns>
    public string? GetString(string key) =>
        _entries.TryGetValue(key, out var entry) && entry is string raw && ScanJsonObjectAlgorithm.ParseText(raw) is JsonText text ? text.Value : null;

    /// <summary>Any value under a key.</summary>
    /// <param name="key">The key.</param>
    /// <returns>The value; <c>null</c> for a missing key.</returns>
    public JsonAny? GetValue(string key) => !_entries.TryGetValue(key, out var entry)
        ? null
        : ScanJsonObjectAlgorithm.ParseText(entry is EditableJsonDocument child ? child.Render() : (string)entry);

    /// <summary>Sets a string, keeping the key's position.</summary>
    /// <param name="key">The key.</param>
    /// <param name="value">The string.</param>
    public void SetString(string key, string value)
    {
        ArgumentNullException.ThrowIfNull(value);
        Put(key, Quote(value));
    }

    /// <summary>
    /// Sets any JSON value, keeping the key's position. A <see cref="JsonMap"/> becomes an editable document, with its
    /// members in the order you wrote them.
    /// </summary>
    /// <param name="key">The key.</param>
    /// <param name="value">The value.</param>
    /// <exception cref="JsonDocumentError">The value has a number that is not finite.</exception>
    public void SetJson(string key, JsonAny value)
    {
        ArgumentNullException.ThrowIfNull(value);
        string text;
        try
        {
            text = WriteCompact(value);
        }
        catch (JsonEncodeError error)
        {
            throw new JsonDocumentError($"{key}: {error.Message}", error);
        }

        Put(key, value is JsonMap ? Parse(text) : text);
    }

    /// <summary>Removes a key.</summary>
    /// <param name="key">The key.</param>
    public void Delete(string key)
    {
        if (_entries.Remove(key))
        {
            _keys.Remove(key);
        }
    }

    /// <summary>The document as npm writes a <c>package.json</c>: two-space indentation, a final newline.</summary>
    /// <returns>The text.</returns>
    public string Render() => $"{Write(string.Empty)}\n";

    /// <inheritdoc />
    public override string ToString() => Render();

    private static EditableJsonDocument FromNode(ObjectNode node)
    {
        var document = new EditableJsonDocument();
        foreach (var (key, value) in node.Members)
        {
            document.Put(key, value is ObjectNode child ? FromNode(child) : ((RawNode)value).Text);
        }

        return document;
    }

    /// <summary>A string as Go's JSON encoder writes it without HTML escaping.</summary>
    private static string Quote(string text)
    {
        var output = new StringBuilder();
        GoJsonTextAlgorithm.AppendString(output, text, escapeHtml: false);
        return output.ToString();
    }

    /// <summary>A value as compact JSON, its members in the order they were added.</summary>
    private static string WriteCompact(JsonAny value) => value switch
    {
        JsonNull => "null",
        JsonBool boolean => boolean.Value ? "true" : "false",
        JsonText text => Quote(text.Value),
        JsonNumber number when number.Integer is null && !double.IsFinite(number.Real) =>
            throw new JsonEncodeError($"{number.Real.ToString(CultureInfo.InvariantCulture)} cannot be written as JSON"),
        JsonNumber number => number.ToString(),
        JsonList list => "[" + string.Join(',', list.Select(WriteCompact)) + "]",
        JsonMap map => "{" + string.Join(',', map.Select(pair => $"{Quote(pair.Key)}:{WriteCompact(pair.Value)}")) + "}",
        _ => throw new JsonEncodeError($"a {value.GetType().Name} cannot be written as JSON"),
    };

    private void Put(string key, object entry)
    {
        if (!_entries.ContainsKey(key))
        {
            _keys.Add(key);
        }

        _entries[key] = entry;
    }

    private string Write(string indent)
    {
        if (_keys.Count == 0)
        {
            return "{}";
        }

        var inner = $"{indent}  ";
        var lines = _keys.Select(key =>
            $"{inner}{Quote(key)}: " + (_entries[key] is EditableJsonDocument child ? child.Write(inner) : IndentRawJsonAlgorithm.Indent((string)_entries[key], inner)));
        return "{\n" + string.Join(",\n", lines) + $"\n{indent}}}";
    }
}
