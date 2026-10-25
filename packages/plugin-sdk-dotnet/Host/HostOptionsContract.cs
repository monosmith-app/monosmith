using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>What the host tells the plugin, and how long it waits.</summary>
public record HostOptions
{
    /// <summary>How long a plugin has to answer after it was asked to stop, before the host kills it.</summary>
    public static readonly TimeSpan DefaultCancelGrace = TimeSpan.FromSeconds(5);

    /// <summary>The longest a call may take unless the caller says otherwise (generate and publish run toolchains).</summary>
    public static readonly TimeSpan DefaultCallTimeout = TimeSpan.FromMinutes(2);

    /// <summary>How long <c>initialize</c> may take.</summary>
    public static readonly TimeSpan DefaultInitializeTimeout = TimeSpan.FromSeconds(30);

    /// <summary>The absolute workspace root sent in <c>initialize</c>. A plugin that reads manifests resolves them against it.</summary>
    public required string WorkspaceRoot { get; init; }

    /// <summary>The host's name and version. Default: <c>smith</c>, <c>dev</c>.</summary>
    public Peer? Host { get; init; }

    /// <summary>How long <c>initialize</c> may take.</summary>
    public TimeSpan InitializeTimeout { get; init; } = DefaultInitializeTimeout;

    /// <summary>
    /// How long a plugin has to answer after it was asked to stop (or, for one that cannot be asked, after the
    /// caller cancelled) before it is killed.
    /// </summary>
    public TimeSpan CancelGrace { get; init; } = DefaultCancelGrace;
}

/// <summary>What happens to the standard error of a plugin process, where it logs.</summary>
public enum StderrMode
{
    /// <summary>It goes to this process's standard error.</summary>
    Inherit,

    /// <summary>It is discarded.</summary>
    Ignore,
}

/// <summary>Options of <see cref="PluginClient.StartAsync"/>: a plugin process to launch.</summary>
public record StartOptions : HostOptions
{
    /// <summary>The plugin's argv: the program and its arguments.</summary>
    public required IReadOnlyList<string> Command { get; init; }

    /// <summary>
    /// The environment. Default: the current one. <c>SMITH_PLUGIN=1</c> is set in it unless it has a value, as smith does.
    /// </summary>
    public IReadOnlyDictionary<string, string>? Environment { get; init; }

    /// <summary>The plugin's working directory. Default: the current one.</summary>
    public string? WorkingDirectory { get; init; }

    /// <summary>What happens to the plugin's standard error. Default: it goes to this process's.</summary>
    public StderrMode Stderr { get; init; } = StderrMode.Inherit;
}

/// <summary>The streams of a plugin that is somewhere else: its responses to read, its requests to write.</summary>
/// <param name="Input">What the plugin writes (its standard output).</param>
/// <param name="Output">What the plugin reads (its standard input).</param>
public sealed record PluginStreams(Stream Input, Stream Output);
