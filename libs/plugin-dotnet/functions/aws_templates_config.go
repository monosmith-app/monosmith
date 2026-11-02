package functions

// awsTemplates are the files of an AWS Lambda function: a class library with the
// Lambda serializer, a SAM template that points at the published folder, and a
// sample event. The handler takes the event type of the trigger, so the tests call it
// with no Lambda host.
var awsTemplates = map[string]string{
	"[[.Dir]]/[[.ID]].csproj": projectProperties + `
    <GenerateRuntimeConfigurationFiles>true</GenerateRuntimeConfigurationFiles>
    <AWSProjectType>Lambda</AWSProjectType>
    <CopyLocalLockFileAssemblies>true</CopyLocalLockFileAssemblies>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Amazon.Lambda.Core" Version="` + lambdaCoreVersion + `" />
    <PackageReference Include="Amazon.Lambda.Serialization.SystemTextJson" Version="` + lambdaJsonVersion + `" />
[[- if eq .Trigger "http"]]
    <PackageReference Include="Amazon.Lambda.APIGatewayEvents" Version="` + lambdaGatewayVersion + `" />
[[- else if eq .Trigger "timer"]]
    <PackageReference Include="Amazon.Lambda.CloudWatchEvents" Version="` + lambdaEventsVersion + `" />
[[- else if eq .Trigger "queue"]]
    <PackageReference Include="Amazon.Lambda.SQSEvents" Version="` + lambdaSQSVersion + `" />
[[- else]]
    <PackageReference Include="Amazon.Lambda.S3Events" Version="` + lambdaS3Version + `" />
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
    <PackageReference Include="Amazon.Lambda.TestUtilities" Version="` + lambdaTestVersion + `" />
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
	// CodeUri is the folder dotnet publish writes (smith run package), where the
	// assembly sits at the root, as Lambda loads it from the zip.
	"[[.Dir]]/template.yaml": `AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Description: [[.Name]]

Globals:
  Function:
    Runtime: [[.Runtime]]
    Timeout: 10
    MemorySize: 256

Resources:
  GreetingFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: [[.PublishDir]]
      Handler: [[.ID]]::[[.ID]].Functions.GreetingHandler::Handle
      Events:
[[- if eq .Trigger "http"]]
        Hello:
          Type: Api
          Properties:
            Path: /hello
            Method: get
[[- else if eq .Trigger "timer"]]
        Tick:
          Type: Schedule
          Properties:
            Schedule: rate(5 minutes)
[[- else if eq .Trigger "queue"]]
        Items:
          Type: SQS
          Properties:
            Queue: !GetAtt ItemsQueue.Arn
            BatchSize: 1
[[- else]]
        Upload:
          Type: S3
          Properties:
            Bucket: !Ref UploadsBucket
            Events: s3:ObjectCreated:*
[[- end]]
[[- if eq .Trigger "queue"]]

  ItemsQueue:
    Type: AWS::SQS::Queue
[[- else if eq .Trigger "blob"]]

  UploadsBucket:
    Type: AWS::S3::Bucket
[[- end]]
[[- if eq .Trigger "http"]]

Outputs:
  HelloApi:
    Description: The local and deployed URL of the function
    Value: !Sub https://${ServerlessRestApi}.execute-api.${AWS::Region}.amazonaws.com/Prod/hello
[[- end]]
`,
	// sam local invoke feeds the function the sample event; the http trigger uses sam local start-api.
	"[[.Dir]]/Events/greeting.json": `[[if eq .Trigger "http"]]{
  "httpMethod": "GET",
  "path": "/hello",
  "queryStringParameters": { "name": "Ada" }
}
[[else if eq .Trigger "timer"]]{
  "source": "aws.events",
  "detail-type": "Scheduled Event",
  "detail": {}
}
[[else if eq .Trigger "queue"]]{
  "Records": [
    { "messageId": "1", "body": "Ada" }
  ]
}
[[else]]{
  "Records": [
    { "s3": { "bucket": { "name": "uploads" }, "object": { "key": "report.csv", "size": 3 } } }
  ]
}
[[end]]`,
	"[[.Dir]]/.gitignore": `.aws-sam/
`,
	"[[.Dir]]/Functions/GreetingHandler.cs": `[[if eq .Trigger "http"]]using System.Text.Json;
using Amazon.Lambda.APIGatewayEvents;
using Amazon.Lambda.Core;
using [[.ID]].Greeting;

[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.SystemTextJson.DefaultLambdaJsonSerializer))]

namespace [[.ID]].Functions;

/// <summary>The function Lambda calls: it answers an API Gateway request with a greeting.</summary>
public class GreetingHandler
{
    /// <summary>Answers GET /hello?name=Ada with a greeting as JSON.</summary>
    public APIGatewayProxyResponse Handle(APIGatewayProxyRequest request, ILambdaContext context)
    {
        string name = "world";
        if (request.QueryStringParameters is { } query && query.TryGetValue("name", out var value))
        {
            name = value;
        }

        return new APIGatewayProxyResponse
        {
            StatusCode = 200,
            Headers = new Dictionary<string, string> { ["content-type"] = "application/json" },
            Body = JsonSerializer.Serialize(GreetUseCase.Greet(name)),
        };
    }
}
[[else if eq .Trigger "timer"]]using Amazon.Lambda.CloudWatchEvents.ScheduledEvents;
using Amazon.Lambda.Core;
using [[.ID]].Greeting;

[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.SystemTextJson.DefaultLambdaJsonSerializer))]

namespace [[.ID]].Functions;

/// <summary>The function Lambda calls: it runs on the schedule of template.yaml.</summary>
public class GreetingHandler
{
    /// <summary>Greets the source of the scheduled event and returns the greeting.</summary>
    public string Handle(ScheduledEvent scheduledEvent, ILambdaContext context) =>
        GreetUseCase.Greet(scheduledEvent.Source).Message;
}
[[else if eq .Trigger "queue"]]using Amazon.Lambda.Core;
using Amazon.Lambda.SQSEvents;
using [[.ID]].Greeting;

[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.SystemTextJson.DefaultLambdaJsonSerializer))]

namespace [[.ID]].Functions;

/// <summary>The function Lambda calls: it greets the messages of an SQS queue.</summary>
public class GreetingHandler
{
    /// <summary>Greets the body of each SQS message and returns the greetings.</summary>
    public List<string> Handle(SQSEvent sqsEvent, ILambdaContext context) =>
        sqsEvent.Records.Select(record => GreetUseCase.Greet(record.Body).Message).ToList();
}
[[else]]using Amazon.Lambda.Core;
using Amazon.Lambda.S3Events;
using [[.ID]].Greeting;

[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.SystemTextJson.DefaultLambdaJsonSerializer))]

namespace [[.ID]].Functions;

/// <summary>The function Lambda calls: it greets the new objects of an S3 bucket.</summary>
public class GreetingHandler
{
    /// <summary>Greets the key of each new S3 object and returns the greetings.</summary>
    public List<string> Handle(S3Event s3Event, ILambdaContext context) =>
        s3Event.Records.Select(record => GreetUseCase.Greet(record.S3.Object.Key).Message).ToList();
}
[[end]]`,
	"[[.Dir]]/tests/Functions/GreetingHandlerTests.cs": `using System.Text;
using Amazon.Lambda.Serialization.SystemTextJson;
using Amazon.Lambda.TestUtilities;
[[- if eq .Trigger "http"]]
using System.Text.Json;
using Amazon.Lambda.APIGatewayEvents;
[[- else if eq .Trigger "timer"]]
using Amazon.Lambda.CloudWatchEvents.ScheduledEvents;
[[- else if eq .Trigger "queue"]]
using Amazon.Lambda.SQSEvents;
[[- else]]
using Amazon.Lambda.S3Events;
[[- end]]
using [[.ID]].Functions;

namespace [[.ID]].Tests.Functions;

/// <summary>The handler gets the event the way Lambda hands it over: parsed from JSON by the serializer.</summary>
public class GreetingHandlerTests
{
    private static T Event<T>(string json) =>
        new DefaultLambdaJsonSerializer().Deserialize<T>(new MemoryStream(Encoding.UTF8.GetBytes(json)));
[[if eq .Trigger "http"]]
    private static string Message(string json)
    {
        var response = new GreetingHandler().Handle(Event<APIGatewayProxyRequest>(json), new TestLambdaContext());
        Assert.Equal(200, response.StatusCode);

        return JsonDocument.Parse(response.Body).RootElement.GetProperty("Message").GetString()!;
    }

    [Fact]
    public void GreetsTheNameInTheQuery() =>
        Assert.Equal("Hello, Ada!", Message("""{ "httpMethod": "GET", "queryStringParameters": { "name": "Ada" } }"""));

    [Fact]
    public void GreetsTheWorldWithoutAName() =>
        Assert.Equal("Hello, world!", Message("""{ "httpMethod": "GET" }"""));
[[else if eq .Trigger "timer"]]
    [Fact]
    public void GreetsTheSourceOfTheEvent() =>
        Assert.Equal("Hello, aws.events!", new GreetingHandler().Handle(Event<ScheduledEvent>("""{ "source": "aws.events" }"""), new TestLambdaContext()));
[[else if eq .Trigger "queue"]]
    [Fact]
    public void GreetsEveryMessage()
    {
        var sqsEvent = Event<SQSEvent>("""{ "Records": [ { "body": "Ada" }, { "body": "Grace" } ] }""");

        Assert.Equal(new[] { "Hello, Ada!", "Hello, Grace!" }, new GreetingHandler().Handle(sqsEvent, new TestLambdaContext()));
    }
[[else]]
    [Fact]
    public void GreetsTheKeyOfTheObject()
    {
        var s3Event = Event<S3Event>("""{ "Records": [ { "s3": { "object": { "key": "report.csv" } } } ] }""");

        Assert.Equal("Hello, report.csv!", Assert.Single(new GreetingHandler().Handle(s3Event, new TestLambdaContext())));
    }
[[end]]}
`,
}
