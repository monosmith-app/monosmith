namespace MonoSmith.Plugin.Sdk;

/// <summary>A JSON string.</summary>
/// <param name="Value">The text.</param>
public sealed record JsonText(string Value) : JsonAny
{
    /// <inheritdoc />
    public override string Describe() => "a string";

    /// <inheritdoc />
    public override string ToString() => Value;
}
