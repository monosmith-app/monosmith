using System.Runtime.InteropServices;

namespace Monosmith.Launcher.ReleaseDownload;

/// <summary>The name of a release archive for one machine.</summary>
internal static class ArchiveNameMapper
{
    /// <summary>The archive's name for the system this runs on.</summary>
    /// <returns>darwin, linux or windows, or "" for a system smith is not built for.</returns>
    internal static string CurrentSystem() =>
        OperatingSystem.IsWindows() ? "windows" : OperatingSystem.IsMacOS() ? "darwin" : OperatingSystem.IsLinux() ? "linux" : "";

    /// <summary>The archive's name for the processor this runs on.</summary>
    /// <returns>amd64 or arm64, or "" for a processor smith is not built for.</returns>
    internal static string CurrentProcessor() => RuntimeInformation.OSArchitecture switch
    {
        Architecture.X64 => "amd64",
        Architecture.Arm64 => "arm64",
        _ => "",
    };

    /// <summary>The release archive of a version for a machine: monosmith_&lt;version&gt;_&lt;os&gt;_&lt;arch&gt;.zip.</summary>
    /// <param name="version">An exact version.</param>
    /// <param name="system">darwin, linux or windows.</param>
    /// <param name="processor">amd64 or arm64.</param>
    /// <returns>The archive's file name.</returns>
    /// <exception cref="InvalidOperationException">A system or processor smith is not built for.</exception>
    internal static string ArchiveName(string version, string system, string processor) =>
        system.Length == 0 || processor.Length == 0
            ? throw new InvalidOperationException($"smith is not built for {RuntimeInformation.OSDescription} on {RuntimeInformation.OSArchitecture}: it runs on macOS, Linux and Windows, on x64 and arm64")
            : $"monosmith_{version}_{system}_{processor}.zip";
}
