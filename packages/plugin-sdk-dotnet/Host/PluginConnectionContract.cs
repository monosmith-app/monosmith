namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>
/// The two ends of the pipe to a plugin, and what to do with the plugin when the stream cannot be trusted.
/// </summary>
/// <param name="Input">The plugin's responses.</param>
/// <param name="Output">The plugin's requests: written to, raises when the plugin is gone.</param>
/// <param name="CloseInput">Ends the plugin's input, which is how a plugin that does not shut down is told the host is gone.</param>
/// <param name="Kill">Called when the plugin did not answer in time: ends the plugin, if it is a process.</param>
/// <param name="Finish">Waits up to the given time for the plugin to go; returns the reason it did not end well, or <c>null</c>.</param>
/// <param name="Dispose">Releases what the connection holds, once nothing reads from it any more.</param>
internal sealed record PluginConnection(
    Stream Input,
    Stream Output,
    Action CloseInput,
    Action Kill,
    Func<TimeSpan, Task<string?>> Finish,
    Action Dispose);
