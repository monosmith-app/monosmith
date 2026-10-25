namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>
/// A test that starts real processes. Locally it is skipped when a tool it needs is missing; in CI, which has them all, a
/// missing tool fails the test instead.
/// </summary>
[AttributeUsage(AttributeTargets.Method)]
internal sealed class IntegrationFactAttribute : FactAttribute
{
    public IntegrationFactAttribute(params string[] tools)
    {
        if (ProcessConversation.InCi)
        {
            return;
        }

        foreach (var tool in tools)
        {
            var arguments = tool == "dotnet" ? "--version" : tool == "go" ? "version" : "--version";
            if (!ProcessConversation.Runs(tool, arguments))
            {
                Skip = $"{tool} is not available";
                return;
            }
        }
    }
}

/// <summary>A theory that starts real processes; see <see cref="IntegrationFactAttribute"/>.</summary>
[AttributeUsage(AttributeTargets.Method)]
internal sealed class IntegrationTheoryAttribute : TheoryAttribute
{
    public IntegrationTheoryAttribute(params string[] tools)
    {
        if (new IntegrationFactAttribute(tools).Skip is { } reason)
        {
            Skip = reason;
        }
    }
}
