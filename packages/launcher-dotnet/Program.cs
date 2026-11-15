// The smith command `dotnet tool install` puts on PATH: a launcher, not smith itself (ADR 0101). It runs the smith
// the workspace pins in smith.toml ([workspace] smith), SMITH_VERSION, or the newest installed, and downloads and
// checks that version the first time.
using System.Reflection;

using Monosmith.Launcher.SmithRun;

var own = typeof(RunSmithUseCase).Assembly.GetCustomAttribute<AssemblyInformationalVersionAttribute>()?.InformationalVersion.Split('+')[0] ?? "0.0.0";
var env = Environment.GetEnvironmentVariables().Cast<System.Collections.DictionaryEntry>()
    .ToDictionary(e => (string)e.Key, e => (string?)e.Value ?? "", StringComparer.Ordinal);

try
{
    var program = RunSmithUseCase.ResolveProgram(env, Environment.CurrentDirectory, line => Console.Error.WriteLine($"smith launcher: {line}"));
    return RunSmithUseCase.RunSmith(program, args, $"nuget@{own}");
}
catch (Exception error) when (error is IOException or HttpRequestException or InvalidOperationException or ArgumentException or UnauthorizedAccessException)
{
    Console.Error.WriteLine($"smith launcher: {error.Message}");
    return 1;
}
