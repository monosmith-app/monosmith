package functions

import (
	"maps"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// Templates are the files a function app adds to the .NET app's own (the
// greeting slice), with [[ ]] delimiters; they replace the app's csproj, its tests
// project and its Program.cs. The data has Dir, Name, ID, Stack, TestSdkVersion and
// the fields of Data.
func Templates(host string) map[string]string {
	files := maps.Clone(commonTemplates)
	switch host {
	case functionhost.Azure:
		maps.Copy(files, azureTemplates)
		maps.Copy(files, recordingLogger)
	case functionhost.AWS:
		maps.Copy(files, awsTemplates)
	case functionhost.GCP:
		maps.Copy(files, gcpTemplates)
		maps.Copy(files, recordingLogger)
	}
	return files
}

// projectProperties is the property group every function csproj starts with: the
// workspace's Directory.Build.props sets the nullable and implicit-usings
// switches, and turns warnings into errors. The audit warnings about a package's
// known vulnerability stay warnings, so a new advisory never breaks a build.
const projectProperties = `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>[[.Framework]]</TargetFramework>
    <RootNamespace>[[.ID]]</RootNamespace>
    <AssemblyName>[[.ID]]</AssemblyName>
    <IsPackable>false</IsPackable>
    <Version>0.0.0</Version>
    <SmithKind>app</SmithKind>
[[- if .Stack]]
    <SmithStack>[[.Stack]]</SmithStack>
[[- end]]
    <WarningsNotAsErrors>NU1901;NU1902;NU1903;NU1904;NETSDK1138</WarningsNotAsErrors>`

// projectTail closes the csproj: the tests project lives beside the code.
const projectTail = `
  <!-- The tests project lives beside the code, in its own project. -->
  <ItemGroup>
    <Compile Remove="tests/**" />
    <None Remove="tests/**" />
  </ItemGroup>

</Project>
`

// commonTemplates are the files of every host.
var commonTemplates = map[string]string{
	"[[.Dir]]/README.md": `# [[.Name]]

A [[.HostLabel]] app in C# on [[.Framework]], made by smith. It answers [[.TriggerLabel]].

- ` + "`smith test`" + ` runs the unit tests of the function: they need no cloud tool.
- ` + "`smith run start`" + ` runs it locally with [[.LocalTool]].
- ` + "`smith run package`" + ` publishes the app into dist/dotnet/publish/[[.ID]]/release.
[[- if eq .Host "azure-functions"]]
- ` + "`smith pipeline deploy-azure-function`" + ` adds a pipeline that deploys the app to an Azure Function App.
- Copy local.settings.sample.json to local.settings.json (git ignores it) for the settings of your machine.
- Create the Function App with the .NET isolated worker on [[.Framework]].
[[- else if eq .Host "aws-lambda"]]
- template.yaml is the SAM template (Runtime [[.Runtime]]); Events/greeting.json is the sample event.
[[- else]]
- Deploy it with ` + "`gcloud run deploy --source .`" + ` or ` + "`gcloud functions deploy`" + ` and the runtime [[.Runtime]].
[[- end]]
`,
}

// recordingLogger is the logger the tests of Azure and Google use to see what a function
// logged, with no logging package. Lambda handlers return their result instead.
var recordingLogger = map[string]string{
	// Test doubles live in a Testing folder (the slice rules).
	"[[.Dir]]/tests/Testing/RecordingLogger.cs": `using Microsoft.Extensions.Logging;

namespace [[.ID]].Tests.Testing;

/// <summary>A logger that keeps the messages, so a test can read what a function logged.</summary>
public sealed class RecordingLogger<T> : ILogger<T>
{
    public List<string> Messages { get; } = [];

    public IDisposable? BeginScope<TState>(TState state)
        where TState : notnull => null;

    public bool IsEnabled(LogLevel logLevel) => true;

    public void Log<TState>(LogLevel logLevel, EventId eventId, TState state, Exception? exception, Func<TState, Exception?, string> formatter) =>
        Messages.Add(formatter(state, exception));
}
`,
}
