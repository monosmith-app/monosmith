using MonoSmith.Plugin.Sdk.Protocol;

namespace MonoSmith.Plugin.Sdk.Testing;

/// <summary>One check: its name, whether it held, and, when it did not, what is wrong.</summary>
/// <param name="Check">The check's name.</param>
/// <param name="Ok">Whether it held.</param>
/// <param name="Detail">What is wrong; <c>null</c> for a check that held.</param>
public sealed record CheckResult(string Check, bool Ok, string? Detail = null);

/// <summary>Every check run against one plugin.</summary>
public sealed class ConformanceReport(Peer plugin)
{
    private readonly List<CheckResult> _results = [];

    /// <summary>Who was checked.</summary>
    public Peer Plugin { get; } = plugin;

    /// <summary>The checks, in the order they ran.</summary>
    public IReadOnlyList<CheckResult> Results => _results;

    /// <summary>Whether every check passed.</summary>
    public bool Ok => _results.All(result => result.Ok);

    /// <summary>The failed checks, one per line: <c>check: detail</c>.</summary>
    /// <returns>The lines; empty when every check passed.</returns>
    public string Failures() => string.Join('\n', _results.Where(result => !result.Ok).Select(result => $"{result.Check}: {result.Detail}"));

    /// <summary>Records a check. The detail is kept only for a check that failed.</summary>
    /// <param name="check">The check's name.</param>
    /// <param name="ok">Whether it held.</param>
    /// <param name="detail">What is wrong, when it did not hold.</param>
    public void Add(string check, bool ok, string detail) => _results.Add(ok ? new CheckResult(check, true) : new CheckResult(check, false, detail));

    /// <summary>The report as <c>smith plugin check --json</c> prints it: <c>{"plugin": ..., "results": [...]}</c>.</summary>
    /// <returns>The JSON text.</returns>
    public string ToJson()
    {
        var writer = new JsonBuilder().BeginObject();
        writer.Property("plugin");
        Plugin.Write(writer);
        writer.Property("results").BeginArray();
        foreach (var result in _results)
        {
            writer.BeginObject();
            writer.Property("check").String(result.Check);
            writer.Property("ok").Boolean(result.Ok);
            if (result.Detail is not null)
            {
                writer.Property("detail").String(result.Detail);
            }

            writer.EndObject();
        }

        writer.EndArray();
        return writer.EndObject().ToString();
    }
}
