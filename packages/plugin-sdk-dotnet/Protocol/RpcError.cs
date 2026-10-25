namespace MonoSmith.Plugin.Sdk.Protocol;

/// <summary>
/// A JSON-RPC 2.0 error object as an exception: what a host's call throws when the plugin answers an error.
/// </summary>
/// <remarks>
/// Its <see cref="Code"/> is one of <see cref="ErrorCodes"/>. The serve loop answers any exception a plugin
/// method throws with <see cref="ErrorCodes.PluginFailure"/>, whatever it is.
/// </remarks>
public class RpcError : Exception
{
    /// <summary>An error with no code and no message.</summary>
    public RpcError()
    {
    }

    /// <summary>An error with a message and no code.</summary>
    /// <param name="message">The text of the error.</param>
    public RpcError(string message)
        : base(message)
    {
    }

    /// <summary>An error with a message and a cause.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public RpcError(string message, Exception innerException)
        : base(message, innerException)
    {
    }

    /// <summary>The error object an answer carried.</summary>
    /// <param name="code">The numeric error code.</param>
    /// <param name="message">The text of the error, as the plugin wrote it.</param>
    public RpcError(long code, string message)
        : base(message)
    {
        Code = code;
    }

    /// <summary>The numeric error code.</summary>
    public long Code { get; }
}
