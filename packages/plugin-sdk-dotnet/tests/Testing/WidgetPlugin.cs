namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>The smallest language plugin; the plugins of the tests add the interfaces under test.</summary>
internal class WidgetPlugin : IPlugin
{
    public virtual PluginDescription Describe(InitializeParams parameters) => new()
    {
        Plugin = new Peer { Name = "widget", Version = "0.1.0" },
        Kind = PluginKind.Language,
        Manifests = ["widget.json"],
    };
}
