namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>
/// A plugin whose <c>discover</c> does what the manifest name says. <c>block*</c> waits for <see cref="Release"/> (or for
/// the cancel), <c>stubborn*</c> waits for <see cref="Release"/> and ignores the cancel, <c>fail*</c> throws, anything
/// else answers at once. Every wait is on a latch, never a sleep.
/// </summary>
internal sealed class LabPlugin(long? concurrency = null, bool cancellable = false) : IPlugin, IDiscoverer
{
    private readonly object _gate = new();
    private int _active;

    public bool Cancellable { get; } = cancellable;

    public ManualResetEventSlim Release { get; } = new(false);

    public int Peak { get; private set; }

    public int Aborted { get; private set; }

    public List<string> Started { get; } = [];

    public PluginDescription Describe(InitializeParams parameters) => new()
    {
        Plugin = new Peer { Name = "lab", Version = "1.0.0" },
        Kind = PluginKind.Language,
        Manifests = ["lab.json"],
        Concurrency = concurrency ?? 0,
    };

    public async Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation)
    {
        var name = parameters.Manifests.Count > 0 ? parameters.Manifests[0] : string.Empty;
        lock (_gate)
        {
            _active++;
            Peak = Math.Max(Peak, _active);
            Started.Add(name);
            Monitor.PulseAll(_gate);
        }

        try
        {
            if (name.StartsWith("fail", StringComparison.Ordinal))
            {
                throw new InvalidOperationException("the manifest is broken");
            }

            if (name.StartsWith("block", StringComparison.Ordinal))
            {
                while (!Release.IsSet)
                {
                    if (cancellation.IsCancellationRequested)
                    {
                        lock (_gate)
                        {
                            Aborted++;
                        }

                        throw new InvalidOperationException("aborted");
                    }

                    await Task.Delay(1, CancellationToken.None);
                }
            }

            if (name.StartsWith("stubborn", StringComparison.Ordinal))
            {
                Assert.True(Release.Wait(Conversation.Wait));
            }
        }
        finally
        {
            lock (_gate)
            {
                _active--;
            }
        }

        return new DiscoverResult { Projects = [new Project { Name = name, Root = name, Language = "lab", Kind = ProjectKind.Lib }] };
    }

    public void WaitStarted(int count)
    {
        lock (_gate)
        {
            var deadline = DateTime.UtcNow + Conversation.Wait;
            while (Started.Count < count)
            {
                var left = deadline - DateTime.UtcNow;
                Assert.True(left > TimeSpan.Zero && Monitor.Wait(_gate, left), $"only {Started.Count} requests started");
            }
        }
    }
}
