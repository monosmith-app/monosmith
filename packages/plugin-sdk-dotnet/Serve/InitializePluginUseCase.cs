using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>The handshake.</summary>
internal static class InitializePluginUseCase
{
    /// <summary>
    /// Answers <c>initialize</c>: refuses a host that does not speak this SDK's protocol version (the message is
    /// the Go SDK's), asks the plugin to describe itself, and fills in the version it chose and the capabilities
    /// derived from the interfaces the plugin implements.
    /// </summary>
    public static InitializeResult Initialize(IPlugin plugin, InitializeParams parameters)
    {
        if (!parameters.ProtocolVersions.Contains(ProtocolInfo.Version))
        {
            // The Go SDK formats the host's list with %v: [1 2].
            var listed = string.Join(' ', parameters.ProtocolVersions);
            throw new InvalidOperationException($"no common protocol version: host speaks [{listed}], plugin speaks {ProtocolInfo.Version}");
        }

        var description = plugin.Describe(parameters)
            ?? throw new InvalidOperationException("describe returned nothing");
        return description.ToInitializeResult(PluginMethods.CapabilitiesOf(plugin));
    }
}
