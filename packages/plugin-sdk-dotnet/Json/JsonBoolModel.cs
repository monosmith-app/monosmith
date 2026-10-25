namespace MonoSmith.Plugin.Sdk;

/// <summary>A JSON boolean.</summary>
/// <param name="Value">The boolean.</param>
public sealed record JsonBool(bool Value) : JsonAny
{
    /// <inheritdoc />
    public override string Describe() => "a boolean";

    /// <inheritdoc />
    public override string ToString() => Value ? "true" : "false";
}
