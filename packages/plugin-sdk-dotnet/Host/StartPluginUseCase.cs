using System.Diagnostics;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>Launching a plugin process the way smith does.</summary>
internal static class StartPluginUseCase
{
    public static async Task<PluginClient> StartAsync(StartOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        if (options.Command.Count == 0)
        {
            throw new ArgumentException("plugin command is empty", nameof(options));
        }

        var start = new ProcessStartInfo(options.Command[0])
        {
            UseShellExecute = false,
            RedirectStandardInput = true,
            RedirectStandardOutput = true,
            RedirectStandardError = options.Stderr == StderrMode.Ignore,
            CreateNoWindow = true,
        };
        foreach (var argument in options.Command.Skip(1))
        {
            start.ArgumentList.Add(argument);
        }

        if (options.WorkingDirectory is not null)
        {
            start.WorkingDirectory = options.WorkingDirectory;
        }

        if (options.Environment is not null)
        {
            start.Environment.Clear();
            foreach (var (name, value) in options.Environment)
            {
                start.Environment[name] = value;
            }
        }

        start.Environment.TryAdd("SMITH_PLUGIN", "1");

        Process process;
        try
        {
            process = Process.Start(start) ?? throw new InvalidOperationException("the process did not start");
        }
        catch (Exception error) when (error is InvalidOperationException or System.ComponentModel.Win32Exception)
        {
            throw new InvalidOperationException($"starting plugin [{string.Join(", ", options.Command)}]: {error.Message}", error);
        }

        if (options.Stderr == StderrMode.Ignore)
        {
            // Nobody reads it, but it must be drained or a chatty plugin blocks on a full pipe.
            _ = Task.Run(() => process.StandardError.BaseStream.CopyToAsync(Stream.Null));
        }

        var connection = new PluginConnection(
            process.StandardOutput.BaseStream,
            process.StandardInput.BaseStream,
            () => CloseInput(process),
            () => Kill(process),
            grace => FinishAsync(process, grace),
            process.Dispose);
        try
        {
            return await PluginClient.OpenAsync(connection, options).ConfigureAwait(false);
        }
        catch
        {
            Kill(process);
            throw;
        }
    }

    private static void CloseInput(Process process)
    {
        try
        {
            process.StandardInput.BaseStream.Close();
        }
        catch (IOException)
        {
            // The plugin is gone.
        }
    }

    private static void Kill(Process process)
    {
        try
        {
            process.Kill(entireProcessTree: true);
        }
        catch (Exception error) when (error is InvalidOperationException or System.ComponentModel.Win32Exception or NotSupportedException)
        {
            // It had exited already.
        }
    }

    private static async Task<string?> FinishAsync(Process process, TimeSpan grace)
    {
        using var timer = new CancellationTokenSource(grace);
        try
        {
            await process.WaitForExitAsync(timer.Token).ConfigureAwait(false);
        }
        catch (OperationCanceledException)
        {
            Kill(process);
            await process.WaitForExitAsync().ConfigureAwait(false);
            return "plugin did not exit after shutdown; killed";
        }

        return process.ExitCode == 0 ? null : $"plugin exited with status {process.ExitCode}";
    }
}
