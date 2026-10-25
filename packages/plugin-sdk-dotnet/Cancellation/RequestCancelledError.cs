namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// Thrown by a method that noticed its request was cancelled (<see cref="CancellationExtensions.ThrowIfCancelled"/>).
/// </summary>
/// <remarks>
/// The serve loop answers a cancelled request that fails with any error with the code
/// <c>RequestCancelled</c>, so throwing this is a convenience, not a requirement.
/// </remarks>
public sealed class RequestCancelledError : Exception
{
    /// <summary>An error with the default message.</summary>
    public RequestCancelledError()
        : base("the request was cancelled")
    {
    }

    /// <summary>An error that says what stopped.</summary>
    /// <param name="message">The text of the error.</param>
    public RequestCancelledError(string message)
        : base(message)
    {
    }

    /// <summary>An error that says what stopped and what caused it.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public RequestCancelledError(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}
