package scaffolding

// Test packages the tests project pins (those of the SDK 10 xunit
// template); bump them deliberately.
const (
	TestSdkVersion      = "17.14.1"
	XunitVersion        = "2.9.3"
	XunitRunnerVersion  = "3.1.4"
	TargetFramework     = "net10.0"
	directoryBuildProps = "Directory.Build.props"
)

// rootTemplates are written once, by the first .NET project.
var rootTemplates = map[string]string{
	directoryBuildProps: `<Project>
  <!-- Every build goes to the workspace's dist/ (smith): bin/ and obj/ of
       every project live under dist/dotnet, never in a source tree. Every
       project inherits the target framework below; change it here. -->
  <PropertyGroup>
    <UseArtifactsOutput>true</UseArtifactsOutput>
    <ArtifactsPath>$(MSBuildThisFileDirectory)dist/dotnet</ArtifactsPath>
    <TargetFramework>` + TargetFramework + `</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
  </PropertyGroup>
</Project>
`,
}

// The greeting slice, a worked example of a vertical slice in C#: a folder
// per outcome, a type per role, the tests mirrored in the tests project.
var sliceTemplates = map[string]string{
	"[[.Dir]]/Greeting/Greeting.cs": `namespace [[.ID]].Greeting;

/// <summary>A greeting addressed to someone.</summary>
public sealed record Greeting(string Message);
`,
	"[[.Dir]]/Greeting/GreetUseCase.cs": `namespace [[.ID]].Greeting;

/// <summary>Greets someone by name.</summary>
public static class GreetUseCase
{
    public static Greeting Greet(string name) => new($"Hello, {name}!");
}
`,
	"[[.Dir]]/tests/[[.ID]].Tests.csproj": `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
[[- if .ExplicitFramework]]
    <TargetFramework>` + TargetFramework + `</TargetFramework>
[[- end]]
    <IsPackable>false</IsPackable>
    <IsTestProject>true</IsTestProject>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.NET.Test.Sdk" Version="` + TestSdkVersion + `" />
    <PackageReference Include="xunit" Version="` + XunitVersion + `" />
    <PackageReference Include="xunit.runner.visualstudio" Version="` + XunitRunnerVersion + `" />
  </ItemGroup>

  <ItemGroup>
    <Using Include="Xunit" />
    <ProjectReference Include="../[[.ID]].csproj" />
  </ItemGroup>

</Project>
`,
	"[[.Dir]]/tests/Greeting/GreetUseCaseTests.cs": `using [[.ID]].Greeting;

namespace [[.ID]].Tests.Greeting;

public class GreetUseCaseTests
{
    [Fact]
    public void GreetsByName() => Assert.Equal("Hello, world!", GreetUseCase.Greet("world").Message);
}
`,
}

var projectTemplate = map[string]string{
	"[[.Dir]]/[[.ID]].csproj": `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
[[- if .ExplicitFramework]]
    <TargetFramework>` + TargetFramework + `</TargetFramework>
[[- end]]
[[- if eq .Kind "app"]]
    <OutputType>Exe</OutputType>
[[- end]]
    <RootNamespace>[[.ID]]</RootNamespace>
    <AssemblyName>[[.ID]]</AssemblyName>
[[- if eq .Kind "package"]]
    <PackageId>[[.ID]]</PackageId>
    <Description>[[.Name]]</Description>
    <PackageLicenseExpression>MIT</PackageLicenseExpression>
[[- else]]
    <IsPackable>false</IsPackable>
[[- end]]
    <Version>0.0.0</Version>
    <SmithKind>[[.Kind]]</SmithKind>
[[- if .Stack]]
    <SmithStack>[[.Stack]]</SmithStack>
[[- end]]
  </PropertyGroup>

  <!-- The tests project lives beside the code, in its own project. -->
  <ItemGroup>
    <Compile Remove="tests/**" />
    <None Remove="tests/**" />
  </ItemGroup>

</Project>
`,
}

var appTemplates = map[string]string{
	"[[.Dir]]/Program.cs": `using [[.ID]].Greeting;

Console.WriteLine(GreetUseCase.Greet("world").Message);
`,
}
