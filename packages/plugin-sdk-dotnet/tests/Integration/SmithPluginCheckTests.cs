using MonoSmith.Plugin.Sdk.Tests.Testing;

namespace MonoSmith.Plugin.Sdk.Tests.Integration;

/// <summary>
/// <c>smith plugin check</c> against the C# examples, started the way smith starts any plugin: the program is
/// <c>dotnet</c>, by its path (a bare name is looked up as a plugin, and <c>dotnet</c> is one), and the argument is the
/// assembly.
/// </summary>
public sealed class SmithPluginCheckTests(ExampleSides examples) : IClassFixture<ExampleSides>
{
    private static (int Status, string Output) Check(IReadOnlyList<string> command, params string[] flags) =>
        ProcessConversation.RunAllowingFailure("smith", ["plugin", "check", .. command, .. flags]);

    [IntegrationFact("smith", "dotnet")]
    public void Passes_the_structure_plugin_that_uses_the_SDK()
    {
        var (status, output) = Check(examples.CSharp("structure-layout-cs").CheckCommand);
        Assert.Contains("widget-layout 0.1.0", output, StringComparison.Ordinal);
        Assert.DoesNotContain("✖", output, StringComparison.Ordinal);
        Assert.Contains("✔ a structure plugin does something", output, StringComparison.Ordinal);
        Assert.Contains("✔ shuts down when asked", output, StringComparison.Ordinal);
        Assert.Equal(0, status);
    }

    [IntegrationFact("smith", "dotnet")]
    public void Passes_the_language_plugin_generating_projects_and_discovering_them_again()
    {
        var (status, output) = Check(examples.CSharp("language-widget-cs").CheckCommand, "--generate");
        Assert.Contains("widget 0.1.0", output, StringComparison.Ordinal);
        Assert.DoesNotContain("✖", output, StringComparison.Ordinal);
        foreach (var kind in new[] { "app", "lib", "package" })
        {
            Assert.Contains($"✔ discovers the generated {kind}", output, StringComparison.Ordinal);
        }

        Assert.Equal(0, status);
    }

    [IntegrationFact("smith", "dotnet")]
    public void Passes_the_plugin_that_is_cancellable_and_declares_a_concurrency_and_checks_both()
    {
        var (status, output) = Check(examples.CSharp("structure-scan-cs").CheckCommand);
        Assert.Contains("widget-scan 0.1.0", output, StringComparison.Ordinal);
        Assert.DoesNotContain("✖", output, StringComparison.Ordinal);
        Assert.Contains("✔ accepts 4 requests at the same time", output, StringComparison.Ordinal);
        Assert.Contains("✔ answers a cancelled request with its result or the cancelled error", output, StringComparison.Ordinal);
        Assert.Contains("✔ ignores a cancel for a request it does not know", output, StringComparison.Ordinal);
        Assert.Equal(0, status);
    }

    [IntegrationFact("smith", "dotnet")]
    public void Fails_a_plugin_that_breaks_its_promises()
    {
        var (status, _) = Check([ProcessConversation.DotnetPath, "--this-is-not-a-command"]);
        Assert.NotEqual(0, status);
    }

    [IntegrationFact("smith", "go", "dotnet")]
    public void Prints_the_same_lines_as_the_Go_implementation_of_each_plugin()
    {
        foreach (var (go, csharp, flags) in new[]
        {
            ("structure-layout-go", "structure-layout-cs", Array.Empty<string>()),
            ("structure-scan-go", "structure-scan-cs", []),
            ("language-widget", "language-widget-cs", ["--generate"]),
        })
        {
            var expected = Check(examples.Go(go).Command, flags);
            var actual = Check(examples.CSharp(csharp).CheckCommand, flags);
            Assert.Equal(0, expected.Status);
            Assert.True(actual.Status == 0, actual.Output);
            Assert.Equal(expected.Output.ReplaceLineEndings("\n"), actual.Output.ReplaceLineEndings("\n"));
        }
    }
}
