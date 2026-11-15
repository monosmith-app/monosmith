using System.IO.Compression;
using System.Text;

using Monosmith.Launcher.SignatureCheck;
using Monosmith.Launcher.VersionStore;

namespace Monosmith.Launcher.ReleaseDownload;

/// <summary>Downloads one smith version, checks it, and puts its program in the versions folder.</summary>
internal static class InstallVersionUseCase
{
    /// <summary>
    /// Downloads a smith version, checks it against the signed checksums.txt, and installs its program. Nothing is
    /// written unless every check passes, and the program appears in one rename, so a run that stops halfway leaves no
    /// half-written smith behind.
    /// </summary>
    /// <param name="env">The environment.</param>
    /// <param name="version">An exact version.</param>
    /// <param name="system">darwin, linux or windows.</param>
    /// <param name="processor">amd64 or arm64.</param>
    /// <param name="say">Where progress goes (stderr).</param>
    /// <param name="publicKey">The key the checksums must be signed with; default smith's release key.</param>
    /// <returns>The installed program's path.</returns>
    /// <exception cref="InvalidOperationException">A file does not pass the checks.</exception>
    /// <exception cref="HttpRequestException">A file cannot be downloaded.</exception>
    internal static string Install(IReadOnlyDictionary<string, string> env, string version, string system, string processor, Action<string> say, string publicKey = VerifyChecksumsAlgorithm.ReleasePublicKey)
    {
        var name = ArchiveNameMapper.ArchiveName(version, system, processor);
        say($"downloading smith {version} ({name})");
        var archive = ReleasesClient.ReleaseFile(env, version, name);
        var checksums = ReleasesClient.ReleaseFile(env, version, "checksums.txt");
        var signature = Encoding.UTF8.GetString(ReleasesClient.ReleaseFile(env, version, "checksums.txt.minisig"));
        var problem = VerifyChecksumsAlgorithm.CheckArchive(archive, name, checksums, signature, publicKey);
        if (problem is not null)
        {
            throw new InvalidOperationException($"smith {version} was not installed: {problem}");
        }
        var program = VersionStoreRepository.ProgramName(system);
        var target = VersionStoreRepository.ProgramPath(env, version, system);
        var staging = Path.Combine(VersionStoreRepository.VersionsDir(env), $".{version}-{Environment.ProcessId}");
        if (Directory.Exists(staging))
        {
            Directory.Delete(staging, recursive: true);
        }
        Directory.CreateDirectory(staging);
        using (var zip = new ZipArchive(new MemoryStream(archive), ZipArchiveMode.Read))
        {
            var entry = zip.Entries.FirstOrDefault(e => e.FullName == program || e.FullName.EndsWith('/' + program, StringComparison.Ordinal))
                ?? throw new InvalidOperationException($"{name} holds no {program}");
            entry.ExtractToFile(Path.Combine(staging, program));
        }
        if (!OperatingSystem.IsWindows())
        {
            File.SetUnixFileMode(Path.Combine(staging, program), (UnixFileMode)0x1ED); // 0755
        }
        var folder = Path.GetDirectoryName(target)!;
        if (Directory.Exists(folder))
        {
            Directory.Delete(folder, recursive: true);
        }
        Directory.Move(staging, folder);
        return target;
    }
}
