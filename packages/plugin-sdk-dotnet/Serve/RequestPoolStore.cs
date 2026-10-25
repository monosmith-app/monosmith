using System.Runtime.ExceptionServices;

namespace MonoSmith.Plugin.Sdk;

/// <summary>One request from the moment it is read until it is answered.</summary>
internal sealed class PooledRequest(ReceivedRequest request) : IDisposable
{
    private readonly CancellationTokenSource _cancel = new();

    /// <summary>The request; its id is never <c>null</c>.</summary>
    public ReceivedRequest Request { get; } = request;

    /// <summary>The request id.</summary>
    public long Id => Request.Id!.Value;

    /// <summary>Cancelled when the host cancels the request (or the plugin is shutting down).</summary>
    public CancellationToken Token => _cancel.Token;

    /// <summary>Whether the host asked for the request to stop.</summary>
    public bool Cancelled { get; set; }

    /// <summary>Asks the work to stop. Call it outside the pool's lock: it runs the token's callbacks.</summary>
    public void Cancel()
    {
        try
        {
            _cancel.Cancel();
        }
        catch (ObjectDisposedException)
        {
            // The request was answered meanwhile: there is nothing left to stop.
        }
    }

    public void Dispose() => _cancel.Dispose();
}

/// <summary>What <see cref="RequestPoolStore.Submit"/> decided.</summary>
internal enum Admission
{
    /// <summary>The request runs, or waits for its turn.</summary>
    Admitted,

    /// <summary>A request with the same id is in flight.</summary>
    Duplicate,

    /// <summary>Too many requests are waiting.</summary>
    Full,
}

/// <summary>
/// The state of the serve loop: the requests in flight, the ones waiting for a worker, and a bounded number of
/// workers running them. Requests start in the order they arrive and may finish in any order, as in the Go SDK's
/// request pool. Workers are started when there is work and a free place and end when there is none, so an idle
/// plugin holds no thread.
/// </summary>
internal sealed class RequestPoolStore(Func<PooledRequest, Task> run)
{
    /// <summary>
    /// The requests waiting for a worker are limited to this many. The host never has more in flight than the
    /// plugin declared, so more means a host that breaks the protocol.
    /// </summary>
    public const int MaxQueued = 1024;

    private readonly object _gate = new();
    private readonly LinkedList<PooledRequest> _queue = new();
    private readonly Dictionary<long, PooledRequest> _inFlight = [];
    private int _limit = 1;
    private int _running;
    private TaskCompletionSource? _idle;
    private ExceptionDispatchInfo? _failure;

    /// <summary>Throws the first exception a run threw, if any.</summary>
    public void ThrowIfFailed() => _failure?.Throw();

    /// <summary>Whether a run threw.</summary>
    public bool HasFailed => _failure is not null;

    /// <summary>How many workers may run at once. Raising it starts waiting requests; lowering it lets running ones finish.</summary>
    public void SetLimit(int limit)
    {
        lock (_gate)
        {
            _limit = Math.Max(1, limit);
            while (_running < _limit && _queue.First is { } next)
            {
                _queue.RemoveFirst();
                Start(next.Value);
            }
        }
    }

    /// <summary>Takes a request: it runs at once on a free worker, or waits its turn.</summary>
    public Admission Submit(ReceivedRequest request)
    {
        var id = request.Id!.Value;
        lock (_gate)
        {
            if (_inFlight.ContainsKey(id))
            {
                return Admission.Duplicate;
            }

            if (_running >= _limit && _queue.Count >= MaxQueued)
            {
                return Admission.Full;
            }

            var pooled = new PooledRequest(request);
            _inFlight[id] = pooled;
            if (_running >= _limit)
            {
                _queue.AddLast(pooled);
            }
            else
            {
                Start(pooled);
            }
        }

        return Admission.Admitted;
    }

    /// <summary>
    /// Forgets an answered request. Call it before the answer is written, so a host that reuses the id right
    /// away is not refused as a duplicate.
    /// </summary>
    public void Release(PooledRequest pooled)
    {
        lock (_gate)
        {
            _inFlight.Remove(pooled.Id);
        }

        pooled.Dispose();
    }

    /// <summary>
    /// Stops the request with an id. One that had not started is taken out of the queue and returned, because no
    /// worker will ever answer it. An unknown id is ignored.
    /// </summary>
    public PooledRequest? Cancel(long id)
    {
        PooledRequest? pooled;
        var waiting = false;
        lock (_gate)
        {
            if (!_inFlight.TryGetValue(id, out pooled))
            {
                return null;
            }

            pooled.Cancelled = true;
            if (_queue.Remove(pooled))
            {
                _inFlight.Remove(id);
                waiting = true;
            }
        }

        pooled.Cancel();
        return waiting ? pooled : null;
    }

    /// <summary>Stops every request. The ones that had not started are returned, to be answered.</summary>
    public IReadOnlyList<PooledRequest> CancelAll()
    {
        List<PooledRequest> all;
        List<PooledRequest> waiting;
        lock (_gate)
        {
            all = [.. _inFlight.Values];
            foreach (var pooled in all)
            {
                pooled.Cancelled = true;
            }

            waiting = [.. _queue];
            _queue.Clear();
            foreach (var pooled in waiting)
            {
                _inFlight.Remove(pooled.Id);
            }
        }

        foreach (var pooled in all)
        {
            pooled.Cancel();
        }

        return waiting;
    }

    /// <summary>Completes when no worker is running.</summary>
    public Task IdleAsync()
    {
        lock (_gate)
        {
            if (_running == 0)
            {
                return Task.CompletedTask;
            }

            _idle ??= new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            return _idle.Task;
        }
    }

    // Called with the lock held.
    private void Start(PooledRequest first)
    {
        _running++;
        _ = Task.Run(() => WorkAsync(first));
    }

    private async Task WorkAsync(PooledRequest first)
    {
        PooledRequest? pooled = first;
        while (pooled is not null)
        {
            try
            {
                await run(pooled).ConfigureAwait(false);
            }
            catch (Exception error)
            {
                // Kept for the loop to throw; the pool goes on.
                _failure ??= ExceptionDispatchInfo.Capture(error);
            }

            pooled = Next();
        }
    }

    private PooledRequest? Next()
    {
        lock (_gate)
        {
            if (_queue.First is { } next && _running <= _limit)
            {
                _queue.RemoveFirst();
                return next.Value;
            }

            _running--;
            if (_running == 0)
            {
                _idle?.TrySetResult();
                _idle = null;
            }

            return null;
        }
    }
}
