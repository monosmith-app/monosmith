namespace MonoSmith.Plugin.Sdk;

/// <summary>What a declared <c>concurrency</c> means: one at a time by default, a count, or unlimited.</summary>
public static class ConcurrencyPolicy
{
    /// <summary>The <c>concurrency</c> value that leaves the number of simultaneous requests to the host.</summary>
    public const long Unlimited = -1;

    /// <summary>How many requests the host keeps in flight to a plugin that declared unlimited concurrency.</summary>
    public const int UnlimitedInFlight = 32;

    /// <summary>The most workers a plugin that declared unlimited concurrency runs: the pool stays bounded.</summary>
    public const int UnlimitedWorkers = 1024;

    /// <summary>Whether a declared concurrency is one the protocol defines: absent (0), a count (1 and above) or unlimited (-1).</summary>
    /// <param name="declared">The declared value.</param>
    /// <returns>Whether the protocol defines it.</returns>
    public static bool IsValid(long declared) => declared >= Unlimited;

    /// <summary>How many requests the host keeps in flight to a plugin that declared <paramref name="declared"/>.</summary>
    /// <param name="declared">The declared value; absent, zero and invalid values are 1.</param>
    /// <returns>The number of requests.</returns>
    public static int InFlightLimit(long? declared) => declared == Unlimited ? UnlimitedInFlight : Count(declared);

    /// <summary>How many requests a plugin that declared <paramref name="declared"/> works on at once (its pool size).</summary>
    /// <param name="declared">The declared value; absent, zero and invalid values are 1.</param>
    /// <returns>The number of workers.</returns>
    public static int WorkerLimit(long? declared) => declared == Unlimited ? UnlimitedWorkers : Count(declared);

    private static int Count(long? declared) => declared is null or < 1 ? 1 : (int)Math.Min(declared.Value, int.MaxValue);
}
