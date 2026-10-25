namespace MonoSmith.Plugin.Sdk;

/// <summary>The one writer of frames, so answers never interleave.</summary>
internal sealed class FrameWriter(Stream output) : IDisposable
{
    private readonly SemaphoreSlim _gate = new(1, 1);

    /// <summary>Writes one message as a frame and flushes it. Throws what the stream throws (a broken pipe).</summary>
    public async Task SendAsync(string body)
    {
        var frame = Framing.Frame(body);
        await _gate.WaitAsync().ConfigureAwait(false);
        try
        {
            await output.WriteAsync(frame).ConfigureAwait(false);
            await output.FlushAsync().ConfigureAwait(false);
        }
        finally
        {
            _gate.Release();
        }
    }

    public void Dispose() => _gate.Dispose();
}
