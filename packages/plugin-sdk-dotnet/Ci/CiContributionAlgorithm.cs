using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk;

/// <summary>Building the answers of <c>ci</c>: steps, provider-native steps and credentials.</summary>
public static class CiSteps
{
    /// <summary>
    /// A provider-native step: a GitHub action (<c>"actions/setup-node@v6"</c>) or an Azure task (<c>"NodeTool@0"</c>).
    /// The inputs are written in the order given. The values of a pipeline are written as quoted YAML strings, except
    /// <c>true</c> and <c>false</c>, so pass strings.
    /// </summary>
    /// <param name="uses">The action or task.</param>
    /// <param name="inputs">The inputs, as name and value pairs.</param>
    /// <returns>The step.</returns>
    public static ActionStep Action(string uses, IEnumerable<KeyValuePair<string, string>>? inputs = null) =>
        new()
        {
            Uses = uses,
            Inputs = [.. (inputs ?? []).Select(pair => new Input { Name = pair.Key, Value = pair.Value })],
        };

    /// <summary>One command a CI phase (<c>setup</c>, <c>audit</c>, <c>postRelease</c>) runs. It blocks unless <paramref name="blocking"/> is false.</summary>
    /// <param name="name">The step's name.</param>
    /// <param name="command">The argv.</param>
    /// <param name="dir">The workspace-relative directory to run in; the workspace root when <c>null</c>.</param>
    /// <param name="blocking">Whether a failure fails the phase.</param>
    /// <param name="cacheFiles">Workspace-relative files whose content decides what an audit step finds (a lockfile): smith skips a passing step while they are unchanged.</param>
    /// <returns>The step.</returns>
    public static Step Command(string name, IEnumerable<string> command, string? dir = null, bool blocking = true, IEnumerable<string>? cacheFiles = null)
    {
        ArgumentNullException.ThrowIfNull(command);
        return new Step
        {
            Name = name,
            Command = [.. command],
            Dir = dir ?? string.Empty,
            Blocking = blocking,
            CacheFiles = [.. cacheFiles ?? []],
        };
    }

    /// <summary>A secret the release phase needs, read from the environment variable <paramref name="env"/>.</summary>
    /// <param name="env">The environment variable.</param>
    /// <param name="description">What the secret is for.</param>
    /// <param name="required">Whether release works without it.</param>
    /// <param name="setup">Whether <c>smith ci setup</c> gets it too (installs from a private registry need it).</param>
    /// <param name="tasks">Whether every step that runs tasks gets it too.</param>
    /// <returns>The credential.</returns>
    public static Credential Credential(string env, string description, bool required = true, bool setup = false, bool tasks = false) =>
        new() { Env = env, Description = description, Required = required, Setup = setup, Tasks = tasks };
}
