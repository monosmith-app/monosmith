using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Answering one request.</summary>
internal static class DispatchRequestUseCase
{
    /// <summary>The text of an exception: its message, or its type when it has none.</summary>
    public static string MessageOf(Exception error) =>
        string.IsNullOrEmpty(error.Message) ? error.GetType().Name : error.Message;

    /// <summary>
    /// Answers one request: <c>initialize</c> and <c>shutdown</c> here, every other method by the interface of the
    /// plugin that implements it.
    /// </summary>
    /// <remarks>
    /// The codes are the Go SDK's: <c>MethodNotFound</c> for a method the plugin does not implement,
    /// <c>InvalidParams</c> for params of the wrong shape, <c>PluginFailure</c> for an exception the method
    /// threw (its message is shown to the user) and <c>InternalError</c> for a result that cannot be written as
    /// the schema describes.
    /// </remarks>
    public static async Task<RequestOutcome> DispatchAsync(IPlugin plugin, ReceivedRequest request, CancellationToken cancellation)
    {
        var method = request.Method;
        if (method == ProtocolMethods.Shutdown)
        {
            return RequestOutcome.Succeeded("{}");
        }

        if (method == ProtocolMethods.Initialize)
        {
            return Initialize(plugin, request);
        }

        if (!PluginMethods.ByName.TryGetValue(method, out var entry) || !entry.Implements(plugin))
        {
            return RequestOutcome.Failed(ErrorCodes.MethodNotFound, $"method not supported: {method}");
        }

        IProtocolWritable parameters;
        try
        {
            parameters = entry.ReadParams(request.Params);
        }
        catch (ParamsError error)
        {
            return RequestOutcome.Failed(ErrorCodes.InvalidParams, error.Message);
        }

        IProtocolWritable? result;
        try
        {
            result = await entry.Invoke(plugin, parameters, cancellation).ConfigureAwait(false);
        }
        catch (Exception error)
        {
            return RequestOutcome.Failed(ErrorCodes.PluginFailure, MessageOf(error));
        }

        return Encode(method, result);
    }

    private static RequestOutcome Initialize(IPlugin plugin, ReceivedRequest request)
    {
        InitializeParams parameters;
        try
        {
            parameters = InitializeParams.Read(request.Params, "params");
        }
        catch (ParamsError error)
        {
            return RequestOutcome.Failed(ErrorCodes.InvalidParams, error.Message);
        }

        InitializeResult result;
        try
        {
            result = InitializePluginUseCase.Initialize(plugin, parameters);
        }
        catch (Exception error)
        {
            return RequestOutcome.Failed(ErrorCodes.PluginFailure, MessageOf(error));
        }

        var encoded = Encode(ProtocolMethods.Initialize, result);
        return encoded.Ok ? encoded with { Initialized = result } : encoded;
    }

    private static RequestOutcome Encode(string method, IProtocolWritable? result)
    {
        try
        {
            if (result is null)
            {
                return RequestOutcome.Succeeded("{}");
            }

            var writer = new JsonBuilder();
            result.Write(writer);
            return RequestOutcome.Succeeded(writer.ToString());
        }
        catch (JsonEncodeError error)
        {
            return RequestOutcome.Failed(ErrorCodes.InternalError, $"{method}: {error.Message}");
        }
        catch (Exception error)
        {
            return RequestOutcome.Failed(ErrorCodes.InternalError, $"{method}: {MessageOf(error)}");
        }
    }
}
