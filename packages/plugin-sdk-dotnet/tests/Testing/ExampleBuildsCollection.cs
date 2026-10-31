namespace MonoSmith.Plugin.Sdk.Tests.Testing;

/// <summary>
/// The test classes that build the C# examples. They share the examples' <c>obj</c> folders, so xUnit must not run them
/// at the same time: two parallel builds of one project race on the apphost copy (<c>MSB3030</c>).
/// </summary>
[CollectionDefinition(Name)]
public sealed class ExampleBuildsCollection
{
    /// <summary>The name that a test class gives to <see cref="CollectionAttribute"/> to join the collection.</summary>
    public const string Name = "Example builds";
}
