using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Integration;

/// <summary>The C# examples are plugins that are tested with their own tests project; the SDK runs those too.</summary>
[Collection(ExampleBuildsCollection.Name)]
public sealed class ExampleTests
{
    [IntegrationTheory("dotnet")]
    [InlineData("language-widget-cs")]
    [InlineData("structure-layout-cs")]
    [InlineData("structure-scan-cs")]
    public void Passes_its_own_tests(string example)
    {
        var tests = Path.Combine(ProcessConversation.Examples, example, "tests");
        var output = ProcessConversation.Run("dotnet", ["test", tests, "-nologo", "-v", "q"]);
        Assert.Contains("Passed!", output, StringComparison.Ordinal);
    }
}
