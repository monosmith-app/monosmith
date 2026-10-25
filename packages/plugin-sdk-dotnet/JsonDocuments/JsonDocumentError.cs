namespace MonoSmith.Plugin.Sdk;

/// <summary>The text is not a JSON object, or not JSON at all.</summary>
public sealed class JsonDocumentError : Exception
{
    /// <summary>An error with no message.</summary>
    public JsonDocumentError()
    {
    }

    /// <summary>An error that says what is wrong with the text.</summary>
    /// <param name="message">The text of the error.</param>
    public JsonDocumentError(string message)
        : base(message)
    {
    }

    /// <summary>An error that says what is wrong and what caused it.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public JsonDocumentError(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}
