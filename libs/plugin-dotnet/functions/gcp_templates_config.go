package functions

// gcpTemplates are the files of a Google Cloud (Cloud Run) function: a project that
// references Google.Cloud.Functions.Hosting, which supplies the entry point, and
// a class that implements the Functions Framework interface of the trigger. The
// framework is a package, so dotnet run is the local tool.
var gcpTemplates = map[string]string{
	"[[.Dir]]/[[.ID]].csproj": projectProperties + `
    <OutputType>Exe</OutputType>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Google.Cloud.Functions.Hosting" Version="` + gcpHostingVersion + `" />
[[- if or (eq .Trigger "queue") (eq .Trigger "blob")]]
    <PackageReference Include="Google.Events.Protobuf" Version="` + gcpEventsVersion + `" />
[[- end]]
  </ItemGroup>
` + projectTail,
	"[[.Dir]]/tests/[[.ID]].Tests.csproj": `<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>[[.TestFramework]]</TargetFramework>
    <IsPackable>false</IsPackable>
    <IsTestProject>true</IsTestProject>
  </PropertyGroup>

  <ItemGroup>
    <FrameworkReference Include="Microsoft.AspNetCore.App" />
    <PackageReference Include="Microsoft.NET.Test.Sdk" Version="[[.TestSdkVersion]]" />
    <PackageReference Include="xunit" Version="[[.XunitVersion]]" />
    <PackageReference Include="xunit.runner.visualstudio" Version="[[.XunitRunnerVersion]]" />
  </ItemGroup>

  <ItemGroup>
    <Using Include="Xunit" />
    <ProjectReference Include="../[[.ID]].csproj" />
  </ItemGroup>

</Project>
`,
	"[[.Dir]]/Functions/GreetingHandler.cs": `[[if eq .Trigger "http"]]using Google.Cloud.Functions.Framework;
using Microsoft.AspNetCore.Http;
using [[.ID]].Greeting;

namespace [[.ID]].Functions;

/// <summary>The function the Functions Framework calls: it answers an HTTP request with a greeting.</summary>
public class GreetingHandler : IHttpFunction
{
    /// <summary>Answers GET /?name=Ada with a greeting as JSON.</summary>
    public async Task HandleAsync(HttpContext context)
    {
        string name = context.Request.Query["name"].ToString();
        await context.Response.WriteAsJsonAsync(GreetUseCase.Greet(name.Length > 0 ? name : "world"));
    }
}
[[else if eq .Trigger "queue"]]using CloudNative.CloudEvents;
using Google.Cloud.Functions.Framework;
using Google.Events.Protobuf.Cloud.PubSub.V1;
using Microsoft.Extensions.Logging;
using [[.ID]].Greeting;

namespace [[.ID]].Functions;

/// <summary>The function the Functions Framework calls: it greets a Pub/Sub message.</summary>
public class GreetingHandler(ILogger<GreetingHandler> logger) : ICloudEventFunction<MessagePublishedData>
{
    /// <summary>Greets the text of the message.</summary>
    public Task HandleAsync(CloudEvent cloudEvent, MessagePublishedData data, CancellationToken cancellationToken)
    {
        string text = data.Message?.TextData ?? "";
        logger.LogInformation("{Message}", GreetUseCase.Greet(text.Length > 0 ? text : "world").Message);

        return Task.CompletedTask;
    }
}
[[else]]using CloudNative.CloudEvents;
using Google.Cloud.Functions.Framework;
using Google.Events.Protobuf.Cloud.Storage.V1;
using Microsoft.Extensions.Logging;
using [[.ID]].Greeting;

namespace [[.ID]].Functions;

/// <summary>The function the Functions Framework calls: it greets a new Cloud Storage object.</summary>
public class GreetingHandler(ILogger<GreetingHandler> logger) : ICloudEventFunction<StorageObjectData>
{
    /// <summary>Greets the name of the object.</summary>
    public Task HandleAsync(CloudEvent cloudEvent, StorageObjectData data, CancellationToken cancellationToken)
    {
        logger.LogInformation("{Message}", GreetUseCase.Greet(data.Name).Message);

        return Task.CompletedTask;
    }
}
[[end]]`,
	"[[.Dir]]/tests/Functions/GreetingHandlerTests.cs": `[[if eq .Trigger "http"]]using System.Text.Json;
using Microsoft.AspNetCore.Http;
using [[.ID]].Functions;

namespace [[.ID]].Tests.Functions;

public class GreetingHandlerTests
{
    private static async Task<string> Message(string query)
    {
        var context = new DefaultHttpContext();
        context.Request.QueryString = new QueryString(query);
        context.Response.Body = new MemoryStream();
        await new GreetingHandler().HandleAsync(context);
        context.Response.Body.Position = 0;
        using var json = await JsonDocument.ParseAsync(context.Response.Body);

        return json.RootElement.GetProperty("message").GetString()!;
    }

    [Fact]
    public async Task GreetsTheNameInTheQuery() => Assert.Equal("Hello, Ada!", await Message("?name=Ada"));

    [Fact]
    public async Task GreetsTheWorldWithoutAName() => Assert.Equal("Hello, world!", await Message(""));
}
[[else if eq .Trigger "queue"]]using CloudNative.CloudEvents;
using Google.Events.Protobuf.Cloud.PubSub.V1;
using [[.ID]].Functions;
using [[.ID]].Tests.Testing;

namespace [[.ID]].Tests.Functions;

public class GreetingHandlerTests
{
    [Fact]
    public async Task GreetsTheTextOfTheMessage()
    {
        var logger = new RecordingLogger<GreetingHandler>();
        var data = new MessagePublishedData { Message = new PubsubMessage { TextData = "Ada" } };
        await new GreetingHandler(logger).HandleAsync(new CloudEvent(), data, CancellationToken.None);

        Assert.Equal("Hello, Ada!", Assert.Single(logger.Messages));
    }
}
[[else]]using CloudNative.CloudEvents;
using Google.Events.Protobuf.Cloud.Storage.V1;
using [[.ID]].Functions;
using [[.ID]].Tests.Testing;

namespace [[.ID]].Tests.Functions;

public class GreetingHandlerTests
{
    [Fact]
    public async Task GreetsTheNameOfTheObject()
    {
        var logger = new RecordingLogger<GreetingHandler>();
        await new GreetingHandler(logger).HandleAsync(new CloudEvent(), new StorageObjectData { Name = "report.csv" }, CancellationToken.None);

        Assert.Equal("Hello, report.csv!", Assert.Single(logger.Messages));
    }
}
[[end]]`,
}
