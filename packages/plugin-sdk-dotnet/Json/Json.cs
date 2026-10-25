namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// Any JSON value, as the protocol's free-form members (<c>settings</c>, <c>params</c>, <c>result</c>) carry it.
/// </summary>
/// <remarks>
/// A tree of <see cref="JsonNull"/>, <see cref="JsonBool"/>, <see cref="JsonNumber"/>, <see cref="JsonText"/>,
/// <see cref="JsonList"/> and <see cref="JsonMap"/>. <see cref="Parse"/> reads text the way Go's
/// <c>encoding/json</c> does and <see cref="Encode"/> writes it the way Go writes a <c>map[string]any</c>.
/// </remarks>
public abstract class JsonAny
{
    private protected JsonAny()
    {
    }

    /// <summary>Reads JSON text. Throws <see cref="JsonParseError"/> when the text is not one JSON value.</summary>
    /// <param name="text">The JSON text.</param>
    /// <returns>The value.</returns>
    public static JsonAny Parse(string text) => ParseJsonAlgorithm.Parse(text);

    /// <summary>
    /// The value as Go writes it: no whitespace, members sorted by the bytes of their keys, and the
    /// characters <c>&lt;</c>, <c>&gt;</c>, <c>&amp;</c>, U+2028 and U+2029 escaped.
    /// </summary>
    /// <returns>Compact JSON text.</returns>
    public string Encode()
    {
        var builder = new JsonBuilder();
        builder.Any(this);
        return builder.ToString();
    }

    /// <summary>How an error message names this kind of value: "a string", "an object".</summary>
    /// <returns>The name, with its article.</returns>
    public abstract string Describe();

    /// <summary>A string as a JSON value.</summary>
    /// <param name="value">The text.</param>
    public static implicit operator JsonAny(string value) => new JsonText(value);

    /// <summary>A boolean as a JSON value.</summary>
    /// <param name="value">The boolean.</param>
    public static implicit operator JsonAny(bool value) => new JsonBool(value);

    /// <summary>An integer as a JSON value.</summary>
    /// <param name="value">The integer.</param>
    public static implicit operator JsonAny(long value) => new JsonNumber(value);

    /// <summary>A floating-point number as a JSON value.</summary>
    /// <param name="value">The number.</param>
    public static implicit operator JsonAny(double value) => new JsonNumber(value);

    /// <summary>Names the JSON value's kind for an error message; <c>null</c> is "null".</summary>
    /// <param name="value">The value, or <c>null</c>.</param>
    /// <returns>The name, with its article.</returns>
    public static string Describe(JsonAny? value) => value is null ? "null" : value.Describe();
}
