namespace MonoSmith.Plugin.Sdk.Protocol;

/// <summary>
/// A value of the protocol that reads itself from JSON and writes itself as Go writes it: every message and
/// every enumerated string of the generated types.
/// </summary>
/// <typeparam name="TSelf">The implementing type.</typeparam>
public interface IProtocolValue<TSelf>
    where TSelf : IProtocolValue<TSelf>
{
    /// <summary>
    /// Reads the value the way Go's decoder shapes it: an absent or <c>null</c> member is the zero value of its
    /// type, and a member of the wrong JSON type throws <see cref="ParamsError"/> naming <paramref name="path"/>.
    /// </summary>
    /// <param name="value">The parsed JSON, or <c>null</c> when it was absent.</param>
    /// <param name="path">Where the value is in the message, for an error: <c>params.projects[0]</c>.</param>
    /// <returns>The value.</returns>
    static abstract TSelf Read(JsonAny? value, string path);

    /// <summary>Writes the value as the Go SDK writes it: members in the schema's order, zero values left out.</summary>
    /// <param name="writer">Where to write.</param>
    void Write(JsonBuilder writer);
}
