namespace MonoSmith.Plugin.Sdk;

/// <summary>
/// Cooperative cancellation: the <see cref="CancellationToken"/> every method of a plugin receives is cancelled when
/// the host cancels the request (a plugin that is <c>IPlugin.Cancellable</c>) or the plugin shuts down.
/// </summary>
/// <remarks>
/// The usual loop checks the token between steps and waits on it instead of sleeping:
/// <code>
/// foreach (var path in files)
/// {
///     cancellation.ThrowIfCancelled("the scan was stopped");
///     Scan(path);
/// }
///
/// await Task.Delay(100, cancellation);
/// </code>
/// </remarks>
public static class CancellationExtensions
{
    /// <summary>Throws <see cref="RequestCancelledError"/> when the work was asked to stop.</summary>
    /// <param name="cancellation">The token of the request.</param>
    /// <param name="message">What the error says; the cancelled answer carries it.</param>
    /// <exception cref="RequestCancelledError">The token is cancelled.</exception>
    public static void ThrowIfCancelled(this CancellationToken cancellation, string message = "the request was cancelled")
    {
        if (cancellation.IsCancellationRequested)
        {
            throw new RequestCancelledError(message);
        }
    }
}
