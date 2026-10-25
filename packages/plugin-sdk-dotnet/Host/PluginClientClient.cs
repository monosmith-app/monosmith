using System.Collections.Concurrent;
using System.Globalization;
using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>
/// The host's side of a connection to one initialized plugin, for a plugin author's tests. smith itself is the host
/// in real use.
/// </summary>
/// <remarks>
/// <para>
/// The client never has more requests in flight than the plugin declared in <c>concurrency</c>, as smith does: with
/// the default of one, calls are serialised, each waiting for the one before it to be answered. Each call has a
/// timeout and may be cancelled with a <see cref="CancellationToken"/>. Calls may be made from several tasks at once.
/// </para>
/// <para>
/// Open one with <see cref="StartAsync"/> (a process), <see cref="ConnectAsync"/> (streams) or
/// <see cref="ConnectInProcessAsync"/> (a plugin object served in this process), and dispose it when done: it asks the
/// plugin to shut down.
/// </para>
/// </remarks>
public sealed class PluginClient : IAsyncDisposable
{
    private static readonly TimeSpan ShutdownTimeout = TimeSpan.FromSeconds(5);

    private readonly PluginConnection _connection;
    private readonly ConcurrentDictionary<long, TaskCompletionSource<string>> _waiters = new();
    private readonly SemaphoreSlim _writeGate = new(1, 1);
    private readonly Task _reader;
    private SemaphoreSlim _places = new(1, 1);
    private Exception? _lost;
    private long _nextId;
    private int _inFlight;
    private int _peak;
    private int _closed;

    private PluginClient(PluginConnection connection, TimeSpan cancelGrace)
    {
        _connection = connection;
        CancelGrace = cancelGrace;
        Info = new InitializeResult { Plugin = new Peer { Name = string.Empty, Version = string.Empty }, Kind = default, Capabilities = new Capabilities() };
        _reader = Task.Run(ReadResponsesAsync);
    }

    /// <summary>What the plugin answered to <c>initialize</c>: identity, kind, roles, capabilities.</summary>
    public InitializeResult Info { get; private set; }

    /// <summary>How long the plugin has to answer after it was asked to stop.</summary>
    public TimeSpan CancelGrace { get; }

    /// <summary>The most requests this client has had in flight at once.</summary>
    public int PeakInFlight => Volatile.Read(ref _peak);

    /// <summary>Launches a plugin process and performs the handshake, as smith does.</summary>
    /// <remarks>
    /// The plugin's argv, <c>SMITH_PLUGIN=1</c> in its environment, its standard input and output as the protocol's
    /// channel, its standard error forwarded to this process's. Dispose the client when done: it asks the plugin to shut
    /// down and kills it if it does not go.
    /// </remarks>
    /// <param name="options">What to launch and what to tell it.</param>
    /// <returns>The connected client.</returns>
    public static Task<PluginClient> StartAsync(StartOptions options) => StartPluginUseCase.StartAsync(options);

    /// <summary>
    /// Performs the handshake with a plugin reachable through a pair of streams. Disposing the client shuts the plugin
    /// down and ends the streams; there is no process to kill.
    /// </summary>
    /// <param name="streams">The plugin's streams.</param>
    /// <param name="options">What to tell the plugin.</param>
    /// <returns>The connected client.</returns>
    public static Task<PluginClient> ConnectAsync(PluginStreams streams, HostOptions options) =>
        ConnectPluginUseCase.ConnectAsync(streams, options);

    /// <summary>
    /// Serves a plugin object in this process over in-memory pipes and connects to it: the real framing, handshake and
    /// dispatch, on worker threads, with no process. Disposing the client waits for the serve loop to end.
    /// </summary>
    /// <param name="plugin">The plugin to serve.</param>
    /// <param name="options">What to tell the plugin.</param>
    /// <returns>The connected client.</returns>
    public static Task<PluginClient> ConnectInProcessAsync(IPlugin plugin, HostOptions options) =>
        ConnectPluginUseCase.ConnectInProcessAsync(plugin, options);

    /// <summary>Performs the handshake on a connection.</summary>
    internal static async Task<PluginClient> OpenAsync(PluginConnection connection, HostOptions options)
    {
        var client = new PluginClient(connection, options.CancelGrace);
        InitializeResult info;
        try
        {
            var parameters = new InitializeParams
            {
                ProtocolVersions = [ProtocolInfo.Version],
                Host = options.Host ?? new Peer { Name = "smith", Version = "dev" },
                WorkspaceRoot = options.WorkspaceRoot,
            };
            var result = await client.SendAsync(ProtocolMethods.Initialize, Encode(parameters), options.InitializeTimeout, CancellationToken.None).ConfigureAwait(false);
            info = InitializeResult.Read(result, "result");
        }
        catch (Exception error)
        {
            connection.Kill();
            await client.AbandonAsync().ConfigureAwait(false);
            throw new InvalidOperationException($"initializing plugin: {error.Message}", error);
        }

        if (info.ProtocolVersion != ProtocolInfo.Version)
        {
            connection.Kill();
            await client.AbandonAsync().ConfigureAwait(false);
            throw new InvalidOperationException(
                $"plugin {info.Plugin.Name} chose protocol version {info.ProtocolVersion}, host speaks {ProtocolInfo.Version}");
        }

        client.Info = info;
        var limit = ConcurrencyPolicy.InFlightLimit(info.Concurrency);
        client._places = new SemaphoreSlim(limit, limit);
        return client;
    }

    /// <summary>Sends one request and returns its decoded result.</summary>
    /// <remarks>
    /// <para>
    /// An error answer throws an <see cref="RpcError"/> carrying its code. The call waits for a place first, so the
    /// plugin never has more requests in flight than it declared; the timeout counts from the moment the request is sent.
    /// </para>
    /// <para>
    /// When <paramref name="cancellation"/> is cancelled or the time runs out the host stops waiting. A plugin that
    /// advertised <c>cancel</c> is sent the notification and has <see cref="CancelGrace"/> to answer: with the result it
    /// had already (the call returns it), or an error (the call throws a <see cref="CallInterruptedError"/> carrying it).
    /// A plugin that does not answer in time is killed, because its stream can no longer be trusted, and the call
    /// throws; a plugin that cannot be asked to stop is killed at once on a timeout, and after the grace period when
    /// the caller cancelled.
    /// </para>
    /// </remarks>
    /// <typeparam name="TParams">The message of the params.</typeparam>
    /// <typeparam name="TResult">The message of the result.</typeparam>
    /// <param name="method">The method's name.</param>
    /// <param name="parameters">The params.</param>
    /// <param name="timeout">How long the call may take; default two minutes.</param>
    /// <param name="cancellation">Cancels the call.</param>
    /// <returns>The result.</returns>
    /// <exception cref="RpcError">The plugin answered an error.</exception>
    /// <exception cref="CallInterruptedError">The call was cancelled or timed out.</exception>
    /// <exception cref="InvalidOperationException">The plugin stopped answering.</exception>
    public async Task<TResult> CallAsync<TParams, TResult>(
        string method,
        TParams parameters,
        TimeSpan? timeout = null,
        CancellationToken cancellation = default)
        where TParams : IProtocolWritable
        where TResult : IProtocolValue<TResult>
    {
        ArgumentNullException.ThrowIfNull(parameters);
        var result = await CallRawAsync(method, Encode(parameters), timeout, cancellation).ConfigureAwait(false);
        try
        {
            return TResult.Read(result, "result");
        }
        catch (ParamsError error)
        {
            throw new InvalidOperationException($"{method}: plugin stopped answering: {error.Message}", error);
        }
    }

    /// <summary>Sends one request whose params are JSON text and returns its result as JSON. For a method this SDK does not know.</summary>
    /// <param name="method">The method's name.</param>
    /// <param name="paramsJson">The params as JSON text, or <c>null</c> to send none.</param>
    /// <param name="timeout">How long the call may take; default two minutes.</param>
    /// <param name="cancellation">Cancels the call.</param>
    /// <returns>The result, or <c>null</c> when it was absent or <c>null</c>.</returns>
    public async Task<JsonAny?> CallRawAsync(
        string method,
        string? paramsJson,
        TimeSpan? timeout = null,
        CancellationToken cancellation = default)
    {
        ArgumentNullException.ThrowIfNull(method);
        var span = timeout ?? HostOptions.DefaultCallTimeout;
        if (method == ProtocolMethods.Shutdown)
        {
            // It must reach a plugin that is busy.
            return await SendAsync(method, paramsJson, span, cancellation).ConfigureAwait(false);
        }

        await AcquireAsync(method, cancellation).ConfigureAwait(false);
        try
        {
            return await SendAsync(method, paramsJson, span, cancellation).ConfigureAwait(false);
        }
        finally
        {
            Release();
        }
    }

    /// <summary>Sends a message that gets no answer, such as <c>cancel</c>.</summary>
    /// <typeparam name="TParams">The message of the params.</typeparam>
    /// <param name="method">The method's name.</param>
    /// <param name="parameters">The params.</param>
    /// <returns>A task that ends when the message was written.</returns>
    public async Task NotifyAsync<TParams>(string method, TParams parameters)
        where TParams : IProtocolWritable
    {
        ArgumentNullException.ThrowIfNull(parameters);
        if (_lost is { } lost)
        {
            throw new InvalidOperationException($"{method}: plugin stopped answering: {lost.Message}", lost);
        }

        try
        {
            await WriteAsync(WireMessageMapper.WriteNotification(method, Encode(parameters))).ConfigureAwait(false);
        }
        catch (Exception error) when (error is IOException or ObjectDisposedException)
        {
            throw new InvalidOperationException($"{method}: plugin is not accepting requests: {error.Message}", error);
        }
    }

    /// <summary>Whether the plugin advertised a method: <c>initialize</c> and <c>shutdown</c> always, the others by their capability.</summary>
    /// <param name="method">The method's name.</param>
    /// <returns>Whether it can be called.</returns>
    public bool Supports(string method)
    {
        if (method is ProtocolMethods.Initialize or ProtocolMethods.Shutdown)
        {
            return true;
        }

        return ProtocolMethods.CapabilityByMethod.TryGetValue(method, out var capability)
            && capability is not null
            && Info.Capabilities.Advertises(capability);
    }

    /// <summary>Asks the plugin to shut down and waits briefly for it to go. Throws if it did not go quietly.</summary>
    /// <returns>A task that ends when the plugin is gone.</returns>
    public async Task CloseAsync()
    {
        if (Interlocked.Exchange(ref _closed, 1) == 1)
        {
            return;
        }

        try
        {
            await SendAsync(ProtocolMethods.Shutdown, "{}", ShutdownTimeout, CancellationToken.None).ConfigureAwait(false);
        }
        catch (Exception)
        {
            // A plugin that is already gone cannot be asked; waiting for it below reports how it went.
        }

        _connection.CloseInput();
        var trouble = await _connection.Finish(ShutdownTimeout).ConfigureAwait(false);
        await Task.WhenAny(_reader, Task.Delay(TimeSpan.FromSeconds(2))).ConfigureAwait(false);
        _connection.Dispose();
        _places.Dispose();
        _writeGate.Dispose();
        if (trouble is not null)
        {
            throw new InvalidOperationException(trouble);
        }
    }

    /// <inheritdoc />
    public async ValueTask DisposeAsync() => await CloseAsync().ConfigureAwait(false);

    private static string Encode(IProtocolWritable message)
    {
        var writer = new JsonBuilder();
        message.Write(writer);
        return writer.ToString();
    }

    private async Task AcquireAsync(string method, CancellationToken cancellation)
    {
        try
        {
            await _places.WaitAsync(cancellation).ConfigureAwait(false);
        }
        catch (OperationCanceledException)
        {
            throw new CallInterruptedError($"{method}: cancelled before it was sent", InterruptionReason.Cancelled, false);
        }

        var now = Interlocked.Increment(ref _inFlight);
        int peak;
        do
        {
            peak = Volatile.Read(ref _peak);
        }
        while (now > peak && Interlocked.CompareExchange(ref _peak, now, peak) != peak);
    }

    private void Release()
    {
        Interlocked.Decrement(ref _inFlight);
        _places.Release();
    }

    private async Task WriteAsync(string body)
    {
        var frame = Framing.Frame(body);
        await _writeGate.WaitAsync().ConfigureAwait(false);
        try
        {
            await _connection.Output.WriteAsync(frame).ConfigureAwait(false);
            await _connection.Output.FlushAsync().ConfigureAwait(false);
        }
        finally
        {
            _writeGate.Release();
        }
    }

    private async Task<JsonAny?> SendAsync(string method, string? paramsJson, TimeSpan timeout, CancellationToken cancellation)
    {
        if (_lost is { } lost)
        {
            throw new InvalidOperationException($"{method}: plugin stopped answering: {lost.Message}", lost);
        }

        var waiter = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        var id = Interlocked.Increment(ref _nextId);
        _waiters[id] = waiter;
        try
        {
            await WriteAsync(WireMessageMapper.WriteRequest(id, method, paramsJson)).ConfigureAwait(false);
        }
        catch (Exception error) when (error is IOException or ObjectDisposedException)
        {
            _waiters.TryRemove(id, out _);
            throw new InvalidOperationException($"{method}: plugin is not accepting requests: {error.Message}", error);
        }

        var outcome = await WaitAsync(waiter.Task, timeout, cancellation).ConfigureAwait(false);
        if (outcome == WaitOutcome.Answered)
        {
            _waiters.TryRemove(id, out _);
            return await DecodeAsync(method, waiter).ConfigureAwait(false);
        }

        return await GiveUpAsync(method, id, waiter, outcome == WaitOutcome.Cancelled ? InterruptionReason.Cancelled : InterruptionReason.Timeout, timeout)
            .ConfigureAwait(false);
    }

    private static async Task<JsonAny?> DecodeAsync(string method, TaskCompletionSource<string> waiter)
    {
        string body;
        try
        {
            body = await waiter.Task.ConfigureAwait(false);
        }
        catch (Exception error)
        {
            throw new InvalidOperationException($"{method}: plugin stopped answering: {error.Message}", error);
        }

        ReceivedResponse response;
        try
        {
            response = WireMessageMapper.ParseResponse(body);
        }
        catch (MessageParseError error)
        {
            throw new InvalidOperationException($"{method}: plugin stopped answering: {error.Message}", error);
        }

        if (response.Error is { } failure)
        {
            throw new RpcError(failure.Code, failure.Message);
        }

        return response.Result;
    }

    /// <summary>
    /// The end of a call whose token was cancelled or whose time ran out. A plugin that advertised cancel is sent the
    /// notification and has the grace period to answer; one that cannot be asked is killed at once when the time ran out,
    /// and after the grace period when the caller cancelled. A plugin that does not answer in time is killed, and every
    /// other call in flight to it fails.
    /// </summary>
    private async Task<JsonAny?> GiveUpAsync(string method, long id, TaskCompletionSource<string> waiter, InterruptionReason reason, TimeSpan timeout)
    {
        var asked = Supports(ProtocolMethods.Cancel);
        if (asked)
        {
            try
            {
                await NotifyAsync(ProtocolMethods.Cancel, new CancelParams { Id = id }).ConfigureAwait(false);
            }
            catch (InvalidOperationException)
            {
                // The plugin is gone; waiting for the answer below reports how it went.
            }
        }

        if ((asked || reason == InterruptionReason.Cancelled)
            && await WaitAsync(waiter.Task, CancelGrace, CancellationToken.None).ConfigureAwait(false) == WaitOutcome.Answered)
        {
            _waiters.TryRemove(id, out _);
            return await AnsweredAfterStoppingAsync(method, waiter, reason, timeout).ConfigureAwait(false);
        }

        _waiters.TryRemove(id, out _);
        _connection.Kill();
        throw reason == InterruptionReason.Timeout
            ? new CallInterruptedError($"{method}: plugin did not answer within {Seconds(timeout)}s", reason, true)
            : new CallInterruptedError($"{method}: cancelled; the plugin did not stop within {Seconds(CancelGrace)}s and was killed", reason, true);
    }

    /// <summary>The plugin answered while the host was giving up: a result is the work finished in time.</summary>
    private static async Task<JsonAny?> AnsweredAfterStoppingAsync(string method, TaskCompletionSource<string> waiter, InterruptionReason reason, TimeSpan timeout)
    {
        try
        {
            return await DecodeAsync(method, waiter).ConfigureAwait(false);
        }
        catch (Exception error)
        {
            var what = reason == InterruptionReason.Timeout ? $"did not answer within {Seconds(timeout)}s" : "cancelled";
            throw new CallInterruptedError($"{method}: {what}: {error.Message}", reason, false, error as RpcError);
        }
    }

    private static string Seconds(TimeSpan span) => span.TotalSeconds.ToString("G", CultureInfo.InvariantCulture);

    private static async Task<WaitOutcome> WaitAsync(Task task, TimeSpan timeout, CancellationToken cancellation)
    {
        using var timer = new CancellationTokenSource();
        var elapsed = Task.Delay(timeout, timer.Token);
        var cancelled = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        using var registration = cancellation.Register(() => cancelled.TrySetResult());
        var first = await Task.WhenAny(task, elapsed, cancelled.Task).ConfigureAwait(false);
        await timer.CancelAsync().ConfigureAwait(false);
        if (first == task)
        {
            return WaitOutcome.Answered;
        }

        return first == cancelled.Task ? WaitOutcome.Cancelled : WaitOutcome.TimedOut;
    }

    private async Task ReadResponsesAsync()
    {
        var reader = new FrameReader();
        var buffer = new byte[64 * 1024];
        try
        {
            while (true)
            {
                var read = await _connection.Input.ReadAsync(buffer).ConfigureAwait(false);
                if (read == 0)
                {
                    break;
                }

                foreach (var body in reader.Push(buffer.AsSpan(0, read)))
                {
                    Deliver(body);
                }
            }

            Lose(new EndOfStreamException("EOF"));
        }
        catch (Exception error)
        {
            Lose(error);
        }
    }

    private void Deliver(string body)
    {
        ReceivedResponse response;
        try
        {
            response = WireMessageMapper.ParseResponse(body);
        }
        catch (MessageParseError error)
        {
            // A malformed response cannot be attributed to a call: the stream is lost.
            Lose(new InvalidOperationException($"malformed response: {error.Message}", error));
            return;
        }

        if (response.Id is { } id && _waiters.TryRemove(id, out var waiter))
        {
            waiter.TrySetResult(body);
        }
    }

    private void Lose(Exception error)
    {
        _lost = error;
        foreach (var id in _waiters.Keys)
        {
            if (_waiters.TryRemove(id, out var waiter))
            {
                waiter.TrySetException(error);
            }
        }
    }

    /// <summary>Ends a connection that never completed its handshake.</summary>
    private async Task AbandonAsync()
    {
        _connection.CloseInput();
        await Task.WhenAny(_reader, Task.Delay(TimeSpan.FromSeconds(2))).ConfigureAwait(false);
        _connection.Dispose();
        _places.Dispose();
        _writeGate.Dispose();
        Interlocked.Exchange(ref _closed, 1);
    }

    private enum WaitOutcome
    {
        Answered,
        Cancelled,
        TimedOut,
    }
}
