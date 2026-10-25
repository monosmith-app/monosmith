using System.Collections;
using System.Diagnostics.CodeAnalysis;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// A JSON object. It remembers the order its members were added in, but <see cref="JsonAny.Encode"/>
/// sorts them, as Go sorts the keys of a map. A key that is set twice keeps its first place and takes the
/// last value, as Go's decoder does.
/// </summary>
public sealed class JsonMap : JsonAny, IEnumerable<KeyValuePair<string, JsonAny>>
{
    private readonly List<string> _keys = [];
    private readonly Dictionary<string, JsonAny> _values = new(StringComparer.Ordinal);

    /// <summary>How many members there are.</summary>
    public int Count => _keys.Count;

    /// <summary>The keys, in the order they were first set.</summary>
    public IReadOnlyList<string> Keys => _keys;

    /// <summary>The value of a member. Setting it adds the member, or replaces its value in place.</summary>
    /// <param name="key">The member's name.</param>
    /// <returns>The value; throws when there is no such member.</returns>
    public JsonAny this[string key]
    {
        get => _values[key];
        set
        {
            ArgumentNullException.ThrowIfNull(key);
            if (!_values.ContainsKey(key))
            {
                _keys.Add(key);
            }

            _values[key] = value ?? JsonNull.Instance;
        }
    }

    /// <summary>Whether the object has a member.</summary>
    /// <param name="key">The member's name.</param>
    /// <returns><c>true</c> when it does.</returns>
    public bool ContainsKey(string key) => _values.ContainsKey(key);

    /// <summary>The value of a member, when there is one.</summary>
    /// <param name="key">The member's name.</param>
    /// <param name="value">The value.</param>
    /// <returns><c>true</c> when the member exists.</returns>
    public bool TryGetValue(string key, [MaybeNullWhen(false)] out JsonAny value) => _values.TryGetValue(key, out value);

    /// <summary>Adds or replaces a member.</summary>
    /// <param name="key">The member's name.</param>
    /// <param name="value">The value; <c>null</c> is <see cref="JsonNull"/>.</param>
    public void Add(string key, JsonAny? value) => this[key] = value ?? JsonNull.Instance;

    /// <inheritdoc />
    public override string Describe() => "an object";

    /// <inheritdoc />
    public IEnumerator<KeyValuePair<string, JsonAny>> GetEnumerator() =>
        _keys.Select(key => new KeyValuePair<string, JsonAny>(key, _values[key])).GetEnumerator();

    /// <inheritdoc />
    IEnumerator IEnumerable.GetEnumerator() => GetEnumerator();
}
