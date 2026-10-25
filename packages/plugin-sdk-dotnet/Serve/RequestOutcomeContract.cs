using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>What answering one request came to: the result already written as JSON, or an error object.</summary>
/// <param name="ResultJson">The result as JSON text; <c>null</c> for an error.</param>
/// <param name="Code">The error code.</param>
/// <param name="Message">The error message.</param>
/// <param name="Initialized">What an <c>initialize</c> request answered, so the loop can read the concurrency from it.</param>
internal sealed record RequestOutcome(string? ResultJson, long Code, string Message, InitializeResult? Initialized = null)
{
    /// <summary>Whether the request has a result.</summary>
    public bool Ok => ResultJson is not null;

    /// <summary>An outcome with a result.</summary>
    public static RequestOutcome Succeeded(string resultJson, InitializeResult? initialized = null) =>
        new(resultJson, 0, string.Empty, initialized);

    /// <summary>An outcome with an error.</summary>
    public static RequestOutcome Failed(long code, string message) => new(null, code, message);
}
