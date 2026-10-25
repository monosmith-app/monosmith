namespace MonoSmith.Plugin.Sdk.Protocol;

/// <summary>
/// Shaping received JSON the way the Go SDK's decoder does, for the readers the generator writes: an absent or
/// <c>null</c> member is the zero value of its type (Go writes a nil slice as <c>null</c>), an optional
/// <c>null</c> is absent, and a value of the wrong JSON type is a <see cref="ParamsError"/>.
/// </summary>
internal static class MessageReadingAlgorithm
{
    /// <summary>The object a message is read from. Absent or <c>null</c> is an empty one.</summary>
    public static JsonMap ObjectOf(JsonAny? value, string path) => value switch
    {
        null or JsonNull => new JsonMap(),
        JsonMap map => map,
        _ => throw WrongType(path, "an object", value),
    };

    /// <summary>The value of a member, or <c>null</c> when it is absent or <c>null</c>.</summary>
    public static JsonAny? Member(JsonMap map, string name) =>
        map.TryGetValue(name, out var value) && value is not JsonNull ? value : null;

    /// <summary>Where a member is, for an error.</summary>
    public static string Join(string path, string name) => $"{path}.{name}";

    public static string StringValue(JsonAny? value, string path) => value switch
    {
        null or JsonNull => string.Empty,
        JsonText text => text.Value,
        _ => throw WrongType(path, "a string", value),
    };

    public static long IntegerValue(JsonAny? value, string path) => value switch
    {
        null or JsonNull => 0,
        JsonNumber { Integer: { } integer } => integer,
        _ => throw WrongType(path, "an integer", value),
    };

    /// <summary>An integer that may be <c>null</c>: <c>Response.id</c> is the only one.</summary>
    public static long? NullableIntegerValue(JsonAny? value, string path) => value switch
    {
        null or JsonNull => null,
        JsonNumber { Integer: { } integer } => integer,
        _ => throw WrongType(path, "an integer", value),
    };

    public static double NumberValue(JsonAny? value, string path) => value switch
    {
        null or JsonNull => 0,
        JsonNumber number => number.Real,
        _ => throw WrongType(path, "a number", value),
    };

    public static bool BooleanValue(JsonAny? value, string path) => value switch
    {
        null or JsonNull => false,
        JsonBool boolean => boolean.Value,
        _ => throw WrongType(path, "a boolean", value),
    };

    /// <summary>A free-form member, kept as it came.</summary>
    public static JsonAny? AnyValue(JsonAny? value, string path) => value;

    /// <summary>An optional message: absent is <c>null</c>.</summary>
    public static T? OptionalStruct<T>(JsonAny? value, string path)
        where T : class, IProtocolValue<T> =>
        value is null or JsonNull ? null : T.Read(value, path);

    public static IReadOnlyList<T> ListOf<T>(JsonAny? value, string path, Func<JsonAny?, string, T> item)
    {
        switch (value)
        {
            case null or JsonNull:
                return [];
            case JsonList list:
                var items = new List<T>(list.Count);
                for (var i = 0; i < list.Count; i++)
                {
                    items.Add(item(list[i], $"{path}[{i}]"));
                }

                return items;
            default:
                throw WrongType(path, "an array", value);
        }
    }

    /// <summary>A map. A member that is <c>null</c> is left out unless <paramref name="keepNull"/> says it is a value.</summary>
    public static IReadOnlyDictionary<string, T> MapOf<T>(JsonAny? value, string path, Func<JsonAny?, string, T> item, bool keepNull = false)
    {
        switch (value)
        {
            case null or JsonNull:
                return new Dictionary<string, T>();
            case JsonMap map:
                var members = new Dictionary<string, T>(map.Count, StringComparer.Ordinal);
                foreach (var (key, member) in map)
                {
                    if (member is JsonNull && !keepNull)
                    {
                        continue;
                    }

                    members[key] = item(member, Join(path, key));
                }

                return members;
            default:
                throw WrongType(path, "an object", value);
        }
    }

    public static ParamsError WrongType(string path, string expected, JsonAny? value) =>
        new($"{path} must be {expected}, not {JsonAny.Describe(value)}");
}
