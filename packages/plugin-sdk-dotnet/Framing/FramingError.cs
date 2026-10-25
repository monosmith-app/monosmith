namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// The bytes on a stream are not a sequence of frames: a malformed header, or a missing or invalid
/// <c>Content-Length</c>. The stream cannot be trusted after it.
/// </summary>
public sealed class FramingError : Exception
{
    /// <summary>An error with no message.</summary>
    public FramingError()
    {
    }

    /// <summary>An error that says what is wrong with the stream.</summary>
    /// <param name="message">The text of the error.</param>
    public FramingError(string message)
        : base(message)
    {
    }

    /// <summary>An error that says what is wrong and what caused it.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public FramingError(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}
