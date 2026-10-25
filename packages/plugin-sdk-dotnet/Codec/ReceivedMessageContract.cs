namespace MonoSmith.Plugin.Sdk;

/// <summary>A request or notification as it was received.</summary>
/// <param name="Id">The request id; <c>null</c> for a notification, which gets no response.</param>
/// <param name="Method">The method name; empty when the message carried none.</param>
/// <param name="Params">The params as the JSON parser returned them; <c>null</c> when absent.</param>
internal sealed record ReceivedRequest(long? Id, string Method, JsonAny? Params);

/// <summary>The error object of a response.</summary>
/// <param name="Code">The numeric error code.</param>
/// <param name="Message">The text of the error.</param>
internal sealed record ReceivedError(long Code, string Message);

/// <summary>A response as it was received.</summary>
/// <param name="Id">The id of the request answered; <c>null</c> when the request could not be read.</param>
/// <param name="Result">The result, when the response is not an error.</param>
/// <param name="Error">The error, when the response is one.</param>
internal sealed record ReceivedResponse(long? Id, JsonAny? Result, ReceivedError? Error);
