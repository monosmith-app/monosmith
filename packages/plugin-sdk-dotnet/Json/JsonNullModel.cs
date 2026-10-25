namespace MonoSmith.Plugin.Sdk;

/// <summary>The JSON value <c>null</c>.</summary>
public sealed class JsonNull : JsonAny
{
    private JsonNull()
    {
    }

    /// <summary>The one <c>null</c>.</summary>
    public static JsonNull Instance { get; } = new();

    /// <inheritdoc />
    public override string Describe() => "null";

    /// <inheritdoc />
    public override string ToString() => "null";
}
