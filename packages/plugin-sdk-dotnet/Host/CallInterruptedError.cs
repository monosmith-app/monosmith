using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>Why the host stopped waiting for a call.</summary>
public enum InterruptionReason
{
    /// <summary>The cancellation token of the call was cancelled.</summary>
    Cancelled,

    /// <summary>The time of the call ran out.</summary>
    Timeout,
}

/// <summary>
/// A call the host gave up on, because its token was cancelled (<see cref="InterruptionReason.Cancelled"/>) or its
/// time ran out (<see cref="InterruptionReason.Timeout"/>).
/// </summary>
/// <remarks>
/// <see cref="Killed"/> says the plugin did not stop in time and was killed, which also fails every other call in
/// flight to it. When the plugin answered while being stopped, <see cref="Response"/> is its error (the
/// request-cancelled code, usually).
/// </remarks>
public sealed class CallInterruptedError : Exception
{
    /// <summary>An error with no message.</summary>
    public CallInterruptedError()
    {
    }

    /// <summary>An error with a message.</summary>
    /// <param name="message">The text of the error.</param>
    public CallInterruptedError(string message)
        : base(message)
    {
    }

    /// <summary>An error with a message and a cause.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="innerException">The cause.</param>
    public CallInterruptedError(string message, Exception innerException)
        : base(message, innerException)
    {
    }

    /// <summary>A call the host gave up on.</summary>
    /// <param name="message">The text of the error.</param>
    /// <param name="reason">Why the host stopped waiting.</param>
    /// <param name="killed">Whether the plugin was killed.</param>
    /// <param name="response">What the plugin answered while it was being stopped, if it did.</param>
    public CallInterruptedError(string message, InterruptionReason reason, bool killed, RpcError? response = null)
        : base(message, response)
    {
        Reason = reason;
        Killed = killed;
        Response = response;
    }

    /// <summary>Why the host stopped waiting.</summary>
    public InterruptionReason Reason { get; }

    /// <summary>Whether the plugin did not stop in time and was killed.</summary>
    public bool Killed { get; }

    /// <summary>What the plugin answered while it was being stopped, if it did.</summary>
    public RpcError? Response { get; }
}
