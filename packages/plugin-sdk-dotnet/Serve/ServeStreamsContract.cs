namespace MonoSmith.Plugin.Sdk;

/// <summary>The streams the serve loop runs on.</summary>
/// <param name="Input">The host's requests, as bytes.</param>
/// <param name="Output">The plugin's responses.</param>
/// <param name="Log">Where a warning about the plugin goes. Default: standard error.</param>
public sealed record ServeStreams(Stream Input, Stream Output, Action<string>? Log = null);
