using System.Collections;

namespace MonoSmith.Plugin.Sdk;

/// <summary>A JSON array.</summary>
public sealed class JsonList : JsonAny, IReadOnlyList<JsonAny>
{
    private readonly List<JsonAny> _items;

    /// <summary>An array of the given items.</summary>
    /// <param name="items">The items, in order. A <c>null</c> item is <see cref="JsonNull"/>.</param>
    public JsonList(IEnumerable<JsonAny?> items)
    {
        ArgumentNullException.ThrowIfNull(items);
        _items = items.Select(item => item ?? JsonNull.Instance).ToList();
    }

    /// <summary>An empty array.</summary>
    public JsonList()
        : this([])
    {
    }

    /// <inheritdoc />
    public int Count => _items.Count;

    /// <inheritdoc />
    public JsonAny this[int index] => _items[index];

    /// <inheritdoc />
    public override string Describe() => "an array";

    /// <inheritdoc />
    public IEnumerator<JsonAny> GetEnumerator() => _items.GetEnumerator();

    /// <inheritdoc />
    IEnumerator IEnumerable.GetEnumerator() => GetEnumerator();

    /// <summary>Appends an item.</summary>
    /// <param name="item">The item; <c>null</c> is <see cref="JsonNull"/>.</param>
    public void Add(JsonAny? item) => _items.Add(item ?? JsonNull.Instance);
}
