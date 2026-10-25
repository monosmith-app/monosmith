namespace MonoSmith.Plugin.Sdk;

/// <summary>A frame body that is not a JSON-RPC message the SDK can read. The serve loop answers it with the parse-error code.</summary>
internal sealed class MessageParseError : Exception
{
    public MessageParseError()
    {
    }

    public MessageParseError(string message)
        : base(message)
    {
    }

    public MessageParseError(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}
