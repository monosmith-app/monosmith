using System.ComponentModel;
using System.Diagnostics;

using Monosmith.Launcher.ReleaseDownload;
using Monosmith.Launcher.SignatureCheck;
using Monosmith.Launcher.VersionChoice;
using Monosmith.Launcher.VersionStore;

namespace Monosmith.Launcher.SmithRun;

/// <summary>The smith program a run starts, and the run itself.</summary>
internal static class RunSmithUseCase
{
    /// <summary>
    /// The smith program a run starts. A pinned version (SMITH_VERSION or the workspace's pin) is downloaded when it is
    /// missing. Without a pin, the newest installed version runs, and only a machine with none asks for the newest
    /// release: from then on <c>smith update</c> moves it.
    /// </summary>
    /// <param name="env">The environment.</param>
    /// <param name="cwd">The current folder.</param>
    /// <param name="say">Where progress goes.</param>
    /// <param name="system">The archive's system name; default this machine's.</param>
    /// <param name="processor">The archive's processor name; default this machine's.</param>
    /// <param name="publicKey">The release key; default smith's (tests sign with their own).</param>
    /// <returns>The program's path.</returns>
    /// <exception cref="InvalidOperationException">The program does not pass the checks.</exception>
    /// <exception cref="HttpRequestException">The program cannot be downloaded.</exception>
    /// <exception cref="ArgumentException">SMITH_VERSION is not a version.</exception>
    internal static string ResolveProgram(IReadOnlyDictionary<string, string> env, string cwd, Action<string> say, string? system = null, string? processor = null, string publicKey = VerifyChecksumsAlgorithm.ReleasePublicKey)
    {
        system ??= ArchiveNameMapper.CurrentSystem();
        processor ??= ArchiveNameMapper.CurrentProcessor();
        var pinned = ChooseVersionPolicy.Choose(env, cwd);
        if (pinned is not null)
        {
            var program = VersionStoreRepository.ProgramPath(env, pinned, system);
            return File.Exists(program) ? program : InstallVersionUseCase.Install(env, pinned, system, processor, say, publicKey);
        }
        var installed = VersionStoreRepository.NewestInstalled(env, system);
        return installed is not null
            ? VersionStoreRepository.ProgramPath(env, installed, system)
            : InstallVersionUseCase.Install(env, ReleasesClient.NewestRelease(env), system, processor, say, publicKey);
    }

    /// <summary>
    /// Runs smith with the launcher's arguments and gives back its exit code. SMITH_LAUNCHER tells smith it was started
    /// by this launcher, so <c>smith update</c> installs into the versions folder.
    /// </summary>
    /// <param name="program">The smith program.</param>
    /// <param name="args">The arguments.</param>
    /// <param name="launcher">The launcher's name and version (nuget@3.0.0).</param>
    /// <returns>smith's exit code (1 when it could not start).</returns>
    internal static int RunSmith(string program, IEnumerable<string> args, string launcher)
    {
        var start = new ProcessStartInfo(program) { UseShellExecute = false };
        foreach (var arg in args)
        {
            start.ArgumentList.Add(arg);
        }
        start.Environment["SMITH_LAUNCHER"] = launcher;
        try
        {
            using var process = Process.Start(start)!;
            process.WaitForExit();
            return process.ExitCode;
        }
        catch (Win32Exception error)
        {
            Console.Error.WriteLine($"smith could not start: {error.Message}");
            return 1;
        }
    }
}
