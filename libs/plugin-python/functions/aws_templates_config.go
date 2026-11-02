package functions

// awsTemplates are the files of an AWS Lambda function. template.yaml is the SAM
// template; its CodeUri is the staged build (package.py), where the package sits
// at the root, as Lambda loads it from the zip.
var awsTemplates = map[string]string{
	"[[.Dir]]/template.yaml": `AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31
Description: [[.Name]]

Globals:
  Function:
    Runtime: python[[.PythonVersion]]
    Timeout: 10
    MemorySize: 128

Resources:
  GreetingFunction:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: [[.BuildDir]]/app
      Handler: [[.Import]].functions.greeting_handler.lambda_handler
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
	"[[.Dir]]/events/greeting.json": `[[if eq .Trigger "http"]]{
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
	"[[.Dir]]/start.py": `"""Builds, then runs the function locally with the SAM CLI (sam): an HTTP function as a local
API, any other trigger by invoking it once with the sample event in events/."""

import subprocess
import sys

windows = sys.platform == "win32"
subprocess.check_call([sys.executable, "package.py", "build"])
[[- if eq .Trigger "http"]]
command = ["sam", "local", "start-api"]
[[- else]]
command = ["sam", "local", "invoke", "GreetingFunction", "--event", "events/greeting.json"]
[[- end]]
sys.exit(subprocess.call(command, shell=windows))
`,
	"[[.Dir]]/src/[[.Import]]/functions/greeting_handler.py": `[[if eq .Trigger "http"]]import json
from typing import Any

from ..greeting import greet


def lambda_handler(event: dict[str, Any], context: object) -> dict[str, Any]:
    """Answer GET /hello?name=Ada with a greeting as JSON."""
    name = (event.get("queryStringParameters") or {}).get("name", "world")
    return {
        "statusCode": 200,
        "headers": {"content-type": "application/json"},
        "body": json.dumps({"message": greet(name).message}),
    }
[[else if eq .Trigger "timer"]]from typing import Any

from ..greeting import greet


def lambda_handler(event: dict[str, Any], context: object) -> str:
    """Greet the source of the scheduled event and return the greeting."""
    message = greet(event.get("source", "schedule")).message
    print(message)
    return message
[[else if eq .Trigger "queue"]]from typing import Any

from ..greeting import greet


def lambda_handler(event: dict[str, Any], context: object) -> dict[str, Any]:
    """Greet the body of each SQS message; a message that fails is reported back to be retried."""
    for record in event["Records"]:
        print(greet(record["body"]).message)
    return {"batchItemFailures": []}
[[else]]from typing import Any

from ..greeting import greet


def lambda_handler(event: dict[str, Any], context: object) -> list[str]:
    """Greet the key of each new S3 object."""
    return [greet(record["s3"]["object"]["key"]).message for record in event["Records"]]
[[end]]`,
	"[[.Dir]]/tests/test_greeting_handler.py": `[[if eq .Trigger "http"]]import json

from [[.Import]].functions import lambda_handler


def test_greets_the_name_in_the_query() -> None:
    response = lambda_handler({"queryStringParameters": {"name": "Ada"}}, None)
    assert response["statusCode"] == 200
    assert json.loads(response["body"]) == {"message": "Hello, Ada!"}


def test_greets_the_world_without_a_name() -> None:
    response = lambda_handler({"queryStringParameters": None}, None)
    assert json.loads(response["body"]) == {"message": "Hello, world!"}
[[else if eq .Trigger "timer"]]from [[.Import]].functions import lambda_handler


def test_greets_the_source_of_the_event() -> None:
    assert lambda_handler({"source": "aws.events"}, None) == "Hello, aws.events!"
[[else if eq .Trigger "queue"]]from [[.Import]].functions import lambda_handler


def test_greets_every_message_and_reports_no_failure() -> None:
    event = {"Records": [{"body": "Ada"}, {"body": "Grace"}]}
    assert lambda_handler(event, None) == {"batchItemFailures": []}
[[else]]from [[.Import]].functions import lambda_handler


def test_greets_the_key_of_the_object() -> None:
    event = {"Records": [{"s3": {"object": {"key": "report.csv"}}}]}
    assert lambda_handler(event, None) == ["Hello, report.csv!"]
[[end]]`,
}
