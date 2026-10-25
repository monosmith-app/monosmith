namespace MonoSmith.Plugin.Sdk.Protocol;

/// <summary>
/// The error of received params (or a received result) that do not have the shape the schema describes. The
/// serve loop answers it with the code <c>-32602</c>.
/// </summary>
public sealed class ParamsError : Exception
{
    /// <summary>An error with no message.</summary>
    public ParamsError()
    {
    }

    /// <summary>An error that says what is wrong.</summary>
    /// <param name="message">The text of the error.</param>
    public ParamsError(string message)
        : base(message)
    {
    }

    /// <summary>An error that says what is wrong and what caused it.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public ParamsError(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}
