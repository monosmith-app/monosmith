namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>Connecting to a plugin that is reachable through streams, or served in this process.</summary>
internal static class ConnectPluginUseCase
{
    public static Task<PluginClient> ConnectAsync(PluginStreams streams, HostOptions options)
    {
        ArgumentNullException.ThrowIfNull(streams);
        ArgumentNullException.ThrowIfNull(options);
        var connection = new PluginConnection(
            streams.Input,
            streams.Output,
            () => CloseQuietly(streams.Output),
            () => CloseQuietly(streams.Output),
            _ => Task.FromResult<string?>(null),
            () => { });
        return PluginClient.OpenAsync(connection, options);
    }

    public static Task<PluginClient> ConnectInProcessAsync(IPlugin plugin, HostOptions options)
    {
        ArgumentNullException.ThrowIfNull(plugin);
        ArgumentNullException.ThrowIfNull(options);
        var toPlugin = new InProcessPipeStore();
        var fromPlugin = new InProcessPipeStore();
        var serving = Task.Run(async () =>
        {
            try
            {
                await PluginServer.ServeAsync(plugin, new ServeStreams(toPlugin.Reader, fromPlugin.Writer, static _ => { })).ConfigureAwait(false);
                return null;
            }
            catch (Exception error)
            {
                return string.IsNullOrEmpty(error.Message) ? error.GetType().Name : error.Message;
            }
            finally
            {
                await fromPlugin.Writer.DisposeAsync().ConfigureAwait(false);
                await toPlugin.Reader.DisposeAsync().ConfigureAwait(false);
            }
        });

        var connection = new PluginConnection(
            fromPlugin.Reader,
            toPlugin.Writer,
            () => CloseQuietly(toPlugin.Writer),
            () => CloseQuietly(toPlugin.Writer),
            async grace =>
            {
                var finished = await Task.WhenAny(serving, Task.Delay(grace)).ConfigureAwait(false);
                return finished == serving ? await serving.ConfigureAwait(false) : "the plugin did not stop after shutdown";
            },
            () => fromPlugin.Reader.Dispose());
        return PluginClient.OpenAsync(connection, options);
    }

    private static void CloseQuietly(Stream stream)
    {
        try
        {
            stream.Close();
        }
        catch (IOException)
        {
            // Already closed.
        }
    }
}
