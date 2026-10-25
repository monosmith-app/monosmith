using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>One method a plugin answers: how to see that it is implemented, read its params and call it.</summary>
/// <param name="Name">The method's name.</param>
/// <param name="Implements">Whether the plugin implements the interface that answers the method.</param>
/// <param name="ReadParams">Reads the params of a request into the message the method takes.</param>
/// <param name="Invoke">Calls the plugin's method.</param>
internal sealed record PluginMethod(
    string Name,
    Func<IPlugin, bool> Implements,
    Func<JsonAny?, IProtocolWritable> ReadParams,
    Func<IPlugin, IProtocolWritable, CancellationToken, Task<IProtocolWritable?>> Invoke);
