using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// The serve loop: a plugin on the process's standard input and output, or on a pair of streams.
/// </summary>
/// <example>
/// <code>
/// public static Task&lt;int&gt; Main() =&gt; PluginServer.RunAsync(new WidgetLint());
/// </code>
/// </example>
public static class PluginServer
{
    private const int ReadSize = 64 * 1024;

    /// <summary>
    /// Runs a plugin on the process's standard input and output until the host sends <c>shutdown</c> or closes the
    /// stream. Use it as the whole body of the plugin's entry point and return its result.
    /// </summary>
    /// <remarks>
    /// <para>
    /// Standard output carries only the protocol's frames: whatever else the plugin or a library writes there
    /// (a <c>Console.WriteLine</c>, a child process) goes to standard error instead.
    /// </para>
    /// <para>The result is the exit code: 0 after <c>shutdown</c>, and 1, after the reason is printed, when the stream fails.</para>
    /// </remarks>
    /// <param name="plugin">The plugin.</param>
    /// <returns>The exit code of the process.</returns>
    public static async Task<int> RunAsync(IPlugin plugin)
    {
        ArgumentNullException.ThrowIfNull(plugin);
        try
        {
            var stdio = ClaimStdioUseCase.Claim();
            await ServeAsync(plugin, new ServeStreams(stdio.Input, stdio.Output)).ConfigureAwait(false);
            return 0;
        }
        catch (Exception error)
        {
            await Console.Error.WriteLineAsync(DispatchRequestUseCase.MessageOf(error)).ConfigureAwait(false);
            return 1;
        }
    }

    /// <summary>
    /// Answers requests from <see cref="ServeStreams.Input"/> on <see cref="ServeStreams.Output"/> until the host
    /// sends <c>shutdown</c> or the input ends, then returns. This is the loop <see cref="RunAsync(IPlugin)"/> runs, on
    /// streams you choose: a test serves a plugin in-process over pipes with it.
    /// </summary>
    /// <remarks>
    /// <para>
    /// It has the semantics of the Go SDK's <c>serve.Serve</c>. Requests are answered by a bounded pool of workers,
    /// as many at a time as the plugin declared in <c>concurrency</c> (one, in order, when it said nothing), and the
    /// answers may be written in any order. A request before <c>initialize</c> is refused (<c>InvalidRequest</c>); a
    /// notification (no <c>id</c>) gets no answer; an unreadable frame body is answered with <c>ParseError</c> and
    /// the loop goes on; an unreadable header ends it with a <see cref="FramingError"/>. A
    /// <see cref="IPlugin.Cancellable"/> plugin has the token of a request cancelled when the host sends
    /// <c>cancel</c> for it, and at the end of the input. <c>shutdown</c> is answered after everything read before it.
    /// </para>
    /// </remarks>
    /// <param name="plugin">The plugin.</param>
    /// <param name="streams">The streams to serve on.</param>
    /// <param name="cancellation">Stops waiting for requests.</param>
    /// <returns>A task that ends with the loop. It faults with the first error writing a response (the host is gone) or a <see cref="FramingError"/>.</returns>
    public static async Task ServeAsync(IPlugin plugin, ServeStreams streams, CancellationToken cancellation = default)
    {
        ArgumentNullException.ThrowIfNull(plugin);
        ArgumentNullException.ThrowIfNull(streams);
        using var writer = new FrameWriter(streams.Output);
        Session? session = null;
        var pool = new RequestPoolStore(pooled => RunRequestAsync(session!, pooled));
        session = new Session(plugin, writer, pool, plugin.Cancellable);
        var reader = new FrameReader();
        var buffer = new byte[ReadSize];
        try
        {
            while (true)
            {
                var read = await streams.Input.ReadAsync(buffer, cancellation).ConfigureAwait(false);
                if (read == 0)
                {
                    break;
                }

                foreach (var body in reader.Push(buffer.AsSpan(0, read)))
                {
                    pool.ThrowIfFailed();
                    if (await ReceiveAsync(session, body).ConfigureAwait(false))
                    {
                        return;
                    }
                }
            }
        }
        finally
        {
            await DrainAsync(session).ConfigureAwait(false);
        }

        pool.ThrowIfFailed();
    }

    /// <summary>Handles one frame. Returns whether it was <c>shutdown</c>, after which the loop ends.</summary>
    private static async Task<bool> ReceiveAsync(Session session, string body)
    {
        ReceivedRequest request;
        try
        {
            request = WireMessageMapper.ParseRequest(body);
        }
        catch (MessageParseError error)
        {
            await session.SendAsync(WireMessageMapper.WriteFailure(null, ErrorCodes.ParseError, error.Message)).ConfigureAwait(false);
            return false;
        }

        if (request.Id is not { } id)
        {
            await NotifiedAsync(session, request).ConfigureAwait(false);
            return false;
        }

        if (!session.Initialized && request.Method != ProtocolMethods.Initialize)
        {
            await session.SendAsync(WireMessageMapper.WriteFailure(id, ErrorCodes.InvalidRequest, "initialize must come first")).ConfigureAwait(false);
            return false;
        }

        if (request.Method == ProtocolMethods.Initialize)
        {
            var outcome = await DispatchRequestUseCase.DispatchAsync(session.Plugin, request, CancellationToken.None).ConfigureAwait(false);
            if (outcome.Ok)
            {
                session.Initialized = true;
                session.Pool.SetLimit(ConcurrencyPolicy.WorkerLimit(outcome.Initialized?.Concurrency));
            }

            await session.AnswerAsync(id, outcome).ConfigureAwait(false);
            return false;
        }

        if (request.Method == ProtocolMethods.Shutdown)
        {
            // Everything read before shutdown, started or still waiting, is answered with its result first, as in the
            // Go SDK; only a cancel message stops a request.
            await session.Pool.IdleAsync().ConfigureAwait(false);
            await session.SendAsync(WireMessageMapper.WriteResult(id, "{}")).ConfigureAwait(false);
            return true;
        }

        switch (session.Pool.Submit(request))
        {
            case Admission.Duplicate:
                await session.SendAsync(WireMessageMapper.WriteFailure(id, ErrorCodes.InvalidRequest, $"request id {id} is already in flight")).ConfigureAwait(false);
                break;
            case Admission.Full:
                await session.SendAsync(WireMessageMapper.WriteFailure(
                    id,
                    ErrorCodes.InvalidRequest,
                    "too many requests in flight: the host must not send more than the plugin's declared concurrency")).ConfigureAwait(false);
                break;
            default:
                break;
        }

        return false;
    }

    /// <summary>
    /// Handles a message without an <c>id</c>. <c>cancel</c> is the only one the protocol defines for a plugin;
    /// every other is ignored, and so is a <c>cancel</c> to a plugin that is not cancellable, or one that names no request.
    /// </summary>
    private static async Task NotifiedAsync(Session session, ReceivedRequest request)
    {
        if (request.Method != ProtocolMethods.Cancel || !session.Cancellable)
        {
            return;
        }

        CancelParams parameters;
        try
        {
            parameters = CancelParams.Read(request.Params, "params");
        }
        catch (ParamsError)
        {
            return;
        }

        if (session.Pool.Cancel(parameters.Id) is { } waiting)
        {
            await AnswerCancelledAsync(session, waiting).ConfigureAwait(false);
        }
    }

    /// <summary>A worker's job: handles one request and answers it.</summary>
    private static async Task RunRequestAsync(Session session, PooledRequest pooled)
    {
        RequestOutcome outcome;
        try
        {
            outcome = await DispatchRequestUseCase.DispatchAsync(session.Plugin, pooled.Request, pooled.Token).ConfigureAwait(false);
        }
        catch (Exception error)
        {
            outcome = RequestOutcome.Failed(ErrorCodes.PluginFailure, DispatchRequestUseCase.MessageOf(error));
        }

        if (!outcome.Ok && outcome.Code == ErrorCodes.PluginFailure && pooled.Cancelled)
        {
            outcome = RequestOutcome.Failed(ErrorCodes.RequestCancelled, $"request cancelled: {outcome.Message}");
        }

        session.Pool.Release(pooled);
        await session.AnswerAsync(pooled.Id, outcome).ConfigureAwait(false);
    }

    private static Task AnswerCancelledAsync(Session session, PooledRequest pooled) =>
        session.SendAsync(WireMessageMapper.WriteFailure(pooled.Id, ErrorCodes.RequestCancelled, "request cancelled before it started"));

    /// <summary>
    /// Ends the work in flight when the loop returns (the input ended, the host is gone). A cancellable plugin is
    /// asked to stop and what had not started is refused; any other finishes what it is doing.
    /// </summary>
    private static async Task DrainAsync(Session session)
    {
        if (session.Cancellable)
        {
            foreach (var waiting in session.Pool.CancelAll())
            {
                try
                {
                    await AnswerCancelledAsync(session, waiting).ConfigureAwait(false);
                }
                catch (IOException)
                {
                    break;
                }
            }
        }

        await session.Pool.IdleAsync().ConfigureAwait(false);
    }

    /// <summary>What the loop remembers between requests.</summary>
    private sealed class Session(IPlugin plugin, FrameWriter writer, RequestPoolStore pool, bool cancellable)
    {
        public IPlugin Plugin { get; } = plugin;

        public RequestPoolStore Pool { get; } = pool;

        public bool Cancellable { get; } = cancellable;

        public bool Initialized { get; set; }

        public Task SendAsync(string body) => writer.SendAsync(body);

        public Task AnswerAsync(long id, RequestOutcome outcome) =>
            SendAsync(outcome.Ok
                ? WireMessageMapper.WriteResult(id, outcome.ResultJson!)
                : WireMessageMapper.WriteFailure(id, outcome.Code, outcome.Message));
    }
}
