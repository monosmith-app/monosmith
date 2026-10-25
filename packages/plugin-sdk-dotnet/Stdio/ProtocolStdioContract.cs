namespace MonoSmith.Plugin.Sdk;

/// <summary>The private copies of standard input and output that carry the protocol, and nothing else.</summary>
/// <param name="Input">The host's requests.</param>
/// <param name="Output">The plugin's responses.</param>
public sealed record ProtocolStdio(Stream Input, Stream Output);
