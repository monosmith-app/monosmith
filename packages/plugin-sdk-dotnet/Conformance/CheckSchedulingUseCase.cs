using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>The checks of how a plugin is called: its declared concurrency and its cancel.</summary>
internal static class CheckSchedulingUseCase
{
    /// <summary>How long a plugin that advertised cancel has to answer a request the host cancelled.</summary>
    public static readonly TimeSpan CancelBound = TimeSpan.FromSeconds(10);

    /// <summary>An id no request ever had.</summary>
    private const long UnknownRequest = 1L << 40;

    private static readonly TimeSpan CallTimeout = TimeSpan.FromMinutes(2);

    /// <summary>
    /// Checks what a plugin says about how it is called. The concurrency it declares must be one the protocol defines
    /// and be accepted, and the cancel it advertises must stop a request within a bound and leave the plugin serving.
    /// The same checks, under the same names, as the Go SDK's.
    /// </summary>
    public static async Task CheckAsync(PluginClient client, ConformanceReport report, TimeSpan? cancelBound = null)
    {
        var info = client.Info;
        var declared = info.Concurrency;
        report.Add(
            "declares a concurrency the protocol defines",
            ConcurrencyPolicy.IsValid(declared),
            $"concurrency {declared} is below -1; it must be absent, 1 or more, or -1 for unlimited");
        var probe = ProbeOf(client);
        if (probe is null)
        {
            return;
        }

        await CheckConcurrencyAsync(client, probe, report).ConfigureAwait(false);
        if (info.Capabilities.Cancel)
        {
            await CheckCancelAsync(client, probe, report, cancelBound ?? CancelBound).ConfigureAwait(false);
        }
    }

    /// <summary>The cheapest call the plugin answers without doing work: a method that takes a list of projects or manifests, asked about none.</summary>
    private static Func<CancellationToken, Task>? ProbeOf(PluginClient client)
    {
        if (client.Supports(ProtocolMethods.Discover))
        {
            return cancellation => client.DiscoverAsync(new DiscoverParams { Manifests = [] }, cancellation, CallTimeout);
        }

        if (client.Supports(ProtocolMethods.Contribute))
        {
            return cancellation => client.ContributeAsync(new ContributeParams { Projects = [] }, cancellation, CallTimeout);
        }

        if (client.Supports(ProtocolMethods.Ci))
        {
            return cancellation => client.CiAsync(new CIParams { Projects = [] }, cancellation, CallTimeout);
        }

        if (client.Supports(ProtocolMethods.Doctor))
        {
            return cancellation => client.DoctorAsync(new DoctorParams { Projects = [] }, cancellation, CallTimeout);
        }

        return null;
    }

    /// <summary>
    /// Sends requests together. A plugin that declares nothing is sent them one at a time and must answer them all;
    /// one that declares a concurrency is sent that many at once (at most four, to keep the check cheap) and must
    /// answer them all.
    /// </summary>
    private static async Task CheckConcurrencyAsync(PluginClient client, Func<CancellationToken, Task> probe, ConformanceReport report)
    {
        var limit = ConcurrencyPolicy.InFlightLimit(client.Info.Concurrency);
        var together = Math.Min(limit, 4);
        var name = $"accepts {together} requests at the same time";
        if (limit == 1)
        {
            together = 3;
            name = "answers 3 requests sent together, one at a time";
        }

        var calls = Enumerable.Range(0, together).Select(_ => Task.Run(() => probe(CancellationToken.None))).ToList();
        Exception? failure = null;
        foreach (var call in calls)
        {
            try
            {
                await call.ConfigureAwait(false);
            }
            catch (Exception error)
            {
                failure ??= error;
            }
        }

        var peak = client.PeakInFlight;
        var detail = failure is null ? "<nil>" : failure.Message;
        report.Add(name, failure is null && peak <= limit, $"{detail} (the host had {peak} in flight, the plugin declared {limit})");
    }

    /// <summary>Cancels a request, then a request that does not exist.</summary>
    private static async Task CheckCancelAsync(PluginClient client, Func<CancellationToken, Task> probe, ConformanceReport report, TimeSpan cancelBound)
    {
        using var source = new CancellationTokenSource();
        var began = DateTime.UtcNow;
        var run = Task.Run(() => probe(source.Token));
        await Task.Delay(5).ConfigureAwait(false); // Let the request leave; a cancel that wins the race is just as good an answer.
        await source.CancelAsync().ConfigureAwait(false);
        var finished = await Task.WhenAny(run, Task.Delay(cancelBound + client.CancelGrace)).ConfigureAwait(false);
        var late = finished != run;
        Exception? error = null;
        if (late)
        {
            error = new InvalidOperationException($"no answer within {cancelBound.TotalSeconds:G}s of the cancel");
        }
        else
        {
            try
            {
                await run.ConfigureAwait(false);
            }
            catch (Exception caught)
            {
                error = caught;
            }
        }

        var stopped = !late && (error is null || error is CallInterruptedError { Killed: false });
        var detail =
            $"{(error is null ? "<nil>" : error.Message)} (a plugin that advertised cancel answers the request, with " +
            $"{ErrorCodes.RequestCancelled} if it stopped, within {cancelBound.TotalSeconds:G}s)";
        report.Add("answers a cancelled request with its result or the cancelled error", stopped && DateTime.UtcNow - began < cancelBound, detail);

        Exception? notifyError = null;
        Exception? followUp = null;
        try
        {
            await client.NotifyAsync(ProtocolMethods.Cancel, new CancelParams { Id = UnknownRequest }).ConfigureAwait(false);
        }
        catch (Exception caught)
        {
            notifyError = caught;
        }

        try
        {
            await probe(CancellationToken.None).ConfigureAwait(false);
        }
        catch (Exception caught)
        {
            followUp = caught;
        }

        report.Add(
            "ignores a cancel for a request it does not know",
            notifyError is null && followUp is null,
            $"{(notifyError is null ? "<nil>" : notifyError.Message)}, then {(followUp is null ? "<nil>" : followUp.Message)}");
    }
}
