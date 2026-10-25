using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;

namespace StructureScan;

/// <summary>
/// A structure plugin whose doctor walks every file of a project.
/// </summary>
/// <remarks>
/// <para>A walk can take a while, so it shows the two things a protocol client expects of such a method:</para>
/// <list type="bullet">
/// <item><description>
/// it is <see cref="Cancellable"/>: the SDK hands each method a <see cref="CancellationToken"/> that is cancelled when
/// smith cancels the request, the walk watches it, and the plugin advertises the cancel capability;
/// </description></item>
/// <item><description>
/// it declares a concurrency of 4: smith may send up to four requests at once, and the SDK runs them together on
/// worker threads. Scanning shares no state, so this costs nothing.
/// </description></item>
/// </list>
/// </remarks>
public sealed class ScanPlugin(TimeSpan delay) : IPlugin, IDoctorContributor
{
    /// <summary>The size from which a file is worth a warning.</summary>
    public const long BigFile = 1 << 20;

    /// <summary>
    /// What a scan throws when it is cancelled. The SDK answers a cancelled request that fails with the request-cancelled
    /// code, whatever the message; a fixed one keeps this plugin and its twins byte-equal.
    /// </summary>
    public const string Stopped = "the scan was stopped";

    private string _workspaceRoot = string.Empty;

    /// <inheritdoc />
    public bool Cancellable => true;

    /// <inheritdoc />
    public PluginDescription Describe(InitializeParams parameters)
    {
        _workspaceRoot = parameters.WorkspaceRoot;
        return new PluginDescription
        {
            Plugin = new Peer { Name = "widget-scan", Version = "0.1.0" },
            Kind = PluginKind.Convention,
            Roles = [Role.Structure],
            Languages = ["widget"],
            Concurrency = 4,
        };
    }

    /// <summary>No file of a project is bigger than a mebibyte.</summary>
    public async Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation)
    {
        var findings = new List<Finding>();
        foreach (var project in parameters.Projects)
        {
            var directory = Path.Combine(_workspaceRoot, project.Root.Replace('/', Path.DirectorySeparatorChar));
            var check = $"{project.Name} has no file over 1 MiB";
            List<string> big;
            try
            {
                big = await BigFilesAsync(directory, cancellation);
            }
            catch (IOException error)
            {
                throw new IOException($"scanning {Path.GetFileName(directory)}: {error.Message}", error);
            }

            if (big.Count == 0)
            {
                findings.Add(new Finding { Check = check, Ok = true });
                continue;
            }

            findings.Add(new Finding
            {
                Check = check,
                Ok = false,
                Detail = string.Join(", ", big),
                Remedy = $"keep large files out of the repository: {big[0]}",
                Warning = true, // Reported, but does not fail smith doctor.
            });
        }

        return new DoctorResult { Findings = findings };
    }

    /// <summary>
    /// The files below <paramref name="directory"/> over a mebibyte, relative to it with slashes. It waits the delay
    /// before each file (which stands in for a slow disk) on the cancellation token, so a cancel ends the walk within
    /// one file's time instead of after the sleep.
    /// </summary>
    private async Task<List<string>> BigFilesAsync(string directory, CancellationToken cancellation)
    {
        var found = new List<string>();
        await WalkAsync(string.Empty);
        return found;

        async Task WalkAsync(string relative)
        {
            var current = relative.Length == 0 ? directory : Path.Combine(directory, relative.Replace('/', Path.DirectorySeparatorChar));
            foreach (var entry in new DirectoryInfo(current).GetFileSystemInfos().OrderBy(entry => entry.Name, StringComparer.Ordinal))
            {
                var path = relative.Length == 0 ? entry.Name : $"{relative}/{entry.Name}";
                if (entry is DirectoryInfo)
                {
                    await WalkAsync(path);
                    continue;
                }

                if (delay > TimeSpan.Zero)
                {
                    try
                    {
                        await Task.Delay(delay, cancellation);
                    }
                    catch (OperationCanceledException)
                    {
                        // Cancelled during the pause: handled just below.
                    }
                }

                cancellation.ThrowIfCancelled(Stopped);
                if (((FileInfo)entry).Length > BigFile)
                {
                    found.Add(path);
                }
            }
        }
    }
}
