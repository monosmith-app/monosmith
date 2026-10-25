using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Integration;

/// <summary>The protocol types, the plugin interfaces and the host calls are generated from the protocol's JSON Schema; they must be current.</summary>
public sealed class GeneratedProtocolTests
{
    [IntegrationFact("node")]
    public void The_generated_files_are_what_the_schema_produces()
    {
        var packageRoot = Path.Combine(ProcessConversation.RepositoryRoot, "packages", "plugin-sdk-dotnet");
        var (status, output) = ProcessConversation.RunAllowingFailure("node", [Path.Combine(packageRoot, "Generator", "generate-protocol.handler.mjs"), "--check"], packageRoot);
        Assert.True(status == 0, output);
    }
}
