namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// The error of a value that cannot be written as the message the schema describes, such as a number that
/// is not finite. The serve loop answers it with the code <c>-32603</c>.
/// </summary>
public sealed class JsonEncodeError : Exception
{
    /// <summary>An error with no message.</summary>
    public JsonEncodeError()
    {
    }

    /// <summary>An error that says what cannot be written.</summary>
    /// <param name="message">The text of the error.</param>
    public JsonEncodeError(string message)
        : base(message)
    {
    }

    /// <summary>An error that says what cannot be written and what caused it.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public JsonEncodeError(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}
