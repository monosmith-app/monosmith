package functions

// azureTemplates are the files of an Azure Functions app (C# isolated worker with
// the ASP.NET Core integration). The function methods take plain arguments
// (a request, a string, a byte array), so the tests call them with no host.
var azureTemplates = map[string]string{
	"[[.Dir]]/[[.ID]].csproj": projectProperties + `
    <OutputType>Exe</OutputType>
    <AzureFunctionsVersion>v4</AzureFunctionsVersion>
  </PropertyGroup>

  <ItemGroup>
    <FrameworkReference Include="Microsoft.AspNetCore.App" />
    <PackageReference Include="Microsoft.Azure.Functions.Worker" Version="` + workerVersion + `" />
    <PackageReference Include="Microsoft.Azure.Functions.Worker.Sdk" Version="` + workerSdkVersion + `" />
    <PackageReference Include="Microsoft.Azure.Functions.Worker.Extensions.Http.AspNetCore" Version="` + httpAspNetCoreVer + `" />
[[- if eq .Trigger "timer"]]
    <PackageReference Include="Microsoft.Azure.Functions.Worker.Extensions.Timer" Version="` + timerVersion + `" />
[[- else if eq .Trigger "queue"]]
    <PackageReference Include="Microsoft.Azure.Functions.Worker.Extensions.Storage.Queues" Version="` + queuesVersion + `" />
[[- else if eq .Trigger "blob"]]
    <PackageReference Include="Microsoft.Azure.Functions.Worker.Extensions.Storage.Blobs" Version="` + blobsVersion + `" />
[[- end]]
  </ItemGroup>

  <ItemGroup>
    <None Update="host.json">
      <CopyToOutputDirectory>PreserveNewest</CopyToOutputDirectory>
    </None>
    <None Update="local.settings.json">
      <CopyToOutputDirectory>PreserveNewest</CopyToOutputDirectory>
      <CopyToPublishDirectory>Never</CopyToPublishDirectory>
    </None>
  </ItemGroup>

  <ItemGroup>
    <Using Include="System.Threading.ExecutionContext" Alias="ExecutionContext" />
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
	"[[.Dir]]/Program.cs": `using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Builder;
using Microsoft.Extensions.Hosting;

var builder = FunctionsApplication.CreateBuilder(args);

builder.ConfigureFunctionsWebApplication();

builder.Build().Run();
`,
	"[[.Dir]]/host.json": `{
  "version": "2.0",
  "logging": {
    "logLevel": {
      "default": "Information"
    }
  }
}
`,
	// The sample is committed; local.settings.json (with real connection strings) is not.
	"[[.Dir]]/local.settings.sample.json": `{
  "IsEncrypted": false,
  "Values": {
    "AzureWebJobsStorage": "UseDevelopmentStorage=true",
    "FUNCTIONS_WORKER_RUNTIME": "dotnet-isolated"
  }
}
`,
	"[[.Dir]]/.gitignore": `local.settings.json
`,
	"[[.Dir]]/Functions/GreetingHandler.cs": `[[if eq .Trigger "http"]]using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using [[.ID]].Greeting;

namespace [[.ID]].Functions;

/// <summary>The function the Functions host calls: it answers an HTTP request with a greeting.</summary>
public class GreetingHandler(ILogger<GreetingHandler> logger)
{
    /// <summary>Answers GET /api/hello?name=Ada with a greeting as JSON.</summary>
    [Function("Hello")]
    public IActionResult Hello([HttpTrigger(AuthorizationLevel.Function, "get")] HttpRequest request)
    {
        logger.LogInformation("HTTP trigger for {Path}", request.Path);
        string name = request.Query["name"].ToString();

        return new OkObjectResult(GreetUseCase.Greet(name.Length > 0 ? name : "world"));
    }
}
[[else if eq .Trigger "timer"]]using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using [[.ID]].Greeting;

namespace [[.ID]].Functions;

/// <summary>The function the Functions host calls: it runs on a schedule and logs a greeting.</summary>
public class GreetingHandler(ILogger<GreetingHandler> logger)
{
    /// <summary>Runs every five minutes (the NCRONTAB in the attribute).</summary>
    [Function("Tick")]
    public void Tick([TimerTrigger("0 */5 * * * *")] TimerInfo timer)
    {
        logger.LogInformation("{Message}", GreetUseCase.Greet(timer.IsPastDue ? "late timer" : "timer").Message);
    }
}
[[else if eq .Trigger "queue"]]using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using [[.ID]].Greeting;

namespace [[.ID]].Functions;

/// <summary>The function the Functions host calls: it greets a message of the Storage queue.</summary>
public class GreetingHandler(ILogger<GreetingHandler> logger)
{
    /// <summary>Greets the text of a message from the queue "items".</summary>
    [Function("ProcessMessage")]
    public void ProcessMessage([QueueTrigger("items", Connection = "AzureWebJobsStorage")] string message)
    {
        logger.LogInformation("{Message}", GreetUseCase.Greet(message).Message);
    }
}
[[else]]using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using [[.ID]].Greeting;

namespace [[.ID]].Functions;

/// <summary>The function the Functions host calls: it greets a new blob.</summary>
public class GreetingHandler(ILogger<GreetingHandler> logger)
{
    /// <summary>Greets the name of a new blob in the container "uploads".</summary>
    [Function("ProcessBlob")]
    public void ProcessBlob([BlobTrigger("uploads/{name}", Connection = "AzureWebJobsStorage")] byte[] content, string name)
    {
        logger.LogInformation("{Message} ({Bytes} bytes)", GreetUseCase.Greet(name).Message, content.Length);
    }
}
[[end]]`,
	"[[.Dir]]/tests/Functions/GreetingHandlerTests.cs": `[[if eq .Trigger "http"]]using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging.Abstractions;
using [[.ID]].Functions;
using GreetingRecord = [[.ID]].Greeting.Greeting;

namespace [[.ID]].Tests.Functions;

public class GreetingHandlerTests
{
    private readonly GreetingHandler handler = new(NullLogger<GreetingHandler>.Instance);

    private string Greet(string query)
    {
        var context = new DefaultHttpContext();
        context.Request.QueryString = new QueryString(query);
        var result = Assert.IsType<OkObjectResult>(handler.Hello(context.Request));

        return Assert.IsType<GreetingRecord>(result.Value).Message;
    }

    [Fact]
    public void GreetsTheNameInTheQuery() => Assert.Equal("Hello, Ada!", Greet("?name=Ada"));

    [Fact]
    public void GreetsTheWorldWithoutAName() => Assert.Equal("Hello, world!", Greet(""));
}
[[else if eq .Trigger "timer"]]using Microsoft.Azure.Functions.Worker;
using [[.ID]].Functions;
using [[.ID]].Tests.Testing;

namespace [[.ID]].Tests.Functions;

public class GreetingHandlerTests
{
    [Fact]
    public void LogsAGreeting()
    {
        var logger = new RecordingLogger<GreetingHandler>();
        new GreetingHandler(logger).Tick(new TimerInfo { IsPastDue = false });

        Assert.Equal("Hello, timer!", Assert.Single(logger.Messages));
    }

    [Fact]
    public void SaysWhenTheTimerIsLate()
    {
        var logger = new RecordingLogger<GreetingHandler>();
        new GreetingHandler(logger).Tick(new TimerInfo { IsPastDue = true });

        Assert.Equal("Hello, late timer!", Assert.Single(logger.Messages));
    }
}
[[else if eq .Trigger "queue"]]using [[.ID]].Functions;
using [[.ID]].Tests.Testing;

namespace [[.ID]].Tests.Functions;

public class GreetingHandlerTests
{
    [Fact]
    public void GreetsTheMessage()
    {
        var logger = new RecordingLogger<GreetingHandler>();
        new GreetingHandler(logger).ProcessMessage("Ada");

        Assert.Equal("Hello, Ada!", Assert.Single(logger.Messages));
    }
}
[[else]]using [[.ID]].Functions;
using [[.ID]].Tests.Testing;

namespace [[.ID]].Tests.Functions;

public class GreetingHandlerTests
{
    [Fact]
    public void GreetsTheBlobName()
    {
        var logger = new RecordingLogger<GreetingHandler>();
        new GreetingHandler(logger).ProcessBlob([1, 2, 3], "report.csv");

        Assert.Equal("Hello, report.csv! (3 bytes)", Assert.Single(logger.Messages));
    }
}
[[end]]`,
}
