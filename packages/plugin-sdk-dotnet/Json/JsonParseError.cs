namespace MonoSmith.Plugin.Sdk;

/// <summary>The error of text that is not JSON.</summary>
public sealed class JsonParseError : Exception
{
    /// <summary>An error with no message.</summary>
    public JsonParseError()
    {
    }

    /// <summary>An error that says what is wrong.</summary>
    /// <param name="message">The text of the error.</param>
    public JsonParseError(string message)
        : base(message)
    {
    }

    /// <summary>An error that says what is wrong and what caused it.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public JsonParseError(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}
