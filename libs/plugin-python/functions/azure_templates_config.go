package functions

// azureTemplates are the files of an Azure Functions app (Python v2 model).
// function_app.py is the file the host loads; it registers the blueprint of the
// functions slice. The tests call the plain function, so no host is needed.
var azureTemplates = map[string]string{
	"[[.Dir]]/function_app.py": `"""The entry point the Azure Functions host loads (the Python v2 programming model)."""

import azure.functions as func

from src.[[.Import]].functions import blueprint

app = func.FunctionApp(http_auth_level=func.AuthLevel.FUNCTION)
app.register_functions(blueprint)
`,
	"[[.Dir]]/host.json": `{
  "version": "2.0",
  "extensionBundle": {
    "id": "Microsoft.Azure.Functions.ExtensionBundle",
    "version": "[4.*, 5.0.0)"
  }
}
`,
	"[[.Dir]]/requirements.txt": `azure-functions>=1.25
`,
	// The sample is committed; local.settings.json (with real connection strings) is not.
	"[[.Dir]]/local.settings.sample.json": `{
  "IsEncrypted": false,
  "Values": {
    "AzureWebJobsStorage": "UseDevelopmentStorage=true",
    "FUNCTIONS_WORKER_RUNTIME": "python"
  }
}
`,
	"[[.Dir]]/.gitignore": `local.settings.json
`,
	"[[.Dir]]/start.py": `"""Starts the Functions host (Azure Functions Core Tools: func) with the workspace venv."""

import os
import subprocess
import sys

os.environ["languageWorkers__python__defaultExecutablePath"] = sys.executable
settings = "local.settings.json"
if not os.path.exists(settings) and os.path.exists("local.settings.sample.json"):
    print("No local.settings.json: copy local.settings.sample.json to it for your settings.")
sys.exit(subprocess.call(["func", "start"], shell=sys.platform == "win32"))
`,
	"[[.Dir]]/src/[[.Import]]/functions/greeting_handler.py": `[[if eq .Trigger "http"]]import json

import azure.functions as func

from ..greeting import greet

blueprint = func.Blueprint()


def hello(req: func.HttpRequest) -> func.HttpResponse:
    """Answer GET /api/hello?name=Ada with a greeting as JSON."""
    greeting = greet(req.params.get("name", "world"))
    return func.HttpResponse(json.dumps({"message": greeting.message}), mimetype="application/json")


blueprint.route(route="hello", methods=["GET"])(hello)
[[else if eq .Trigger "timer"]]import logging

import azure.functions as func

from ..greeting import greet

blueprint = func.Blueprint()


def tick(timer: func.TimerRequest) -> str:
    """Log a greeting every five minutes (the NCRONTAB below)."""
    message = greet("late timer" if timer.past_due else "timer").message
    logging.info(message)
    return message


blueprint.timer_trigger(schedule="0 */5 * * * *", arg_name="timer", run_on_startup=False)(tick)
[[else if eq .Trigger "queue"]]import logging

import azure.functions as func

from ..greeting import greet

blueprint = func.Blueprint()


def process_message(message: func.QueueMessage) -> str:
    """Greet the text of a message from the Storage queue "items"."""
    greeting = greet(message.get_body().decode("utf-8")).message
    logging.info(greeting)
    return greeting


blueprint.queue_trigger(arg_name="message", queue_name="items", connection="AzureWebJobsStorage")(
    process_message
)
[[else]]import logging

import azure.functions as func

from ..greeting import greet

blueprint = func.Blueprint()


def process_blob(blob: func.InputStream) -> str:
    """Greet the name of a new blob in the container "uploads"."""
    message = f"{greet(blob.name or 'blob').message} ({blob.length} bytes)"
    logging.info(message)
    return message


blueprint.blob_trigger(arg_name="blob", path="uploads/{name}", connection="AzureWebJobsStorage")(
    process_blob
)
[[end]]`,
	"[[.Dir]]/tests/test_greeting_handler.py": `[[if eq .Trigger "http"]]import json

import azure.functions as func

from [[.Import]].functions import hello


def test_greets_the_name_in_the_query() -> None:
    request = func.HttpRequest(method="GET", url="/api/hello", params={"name": "Ada"}, body=b"")
    assert json.loads(hello(request).get_body()) == {"message": "Hello, Ada!"}


def test_greets_the_world_without_a_name() -> None:
    request = func.HttpRequest(method="GET", url="/api/hello", params={}, body=b"")
    assert json.loads(hello(request).get_body()) == {"message": "Hello, world!"}
[[else if eq .Trigger "timer"]]from types import SimpleNamespace
from typing import cast

import azure.functions as func

from [[.Import]].functions import tick


def test_logs_a_greeting() -> None:
    timer = cast(func.TimerRequest, SimpleNamespace(past_due=False))
    assert tick(timer) == "Hello, timer!"


def test_says_when_the_timer_is_late() -> None:
    timer = cast(func.TimerRequest, SimpleNamespace(past_due=True))
    assert tick(timer) == "Hello, late timer!"
[[else if eq .Trigger "queue"]]import azure.functions as func

from [[.Import]].functions import process_message


def test_greets_the_message() -> None:
    assert process_message(func.QueueMessage(body=b"Ada")) == "Hello, Ada!"
[[else]]from types import SimpleNamespace
from typing import cast

import azure.functions as func

from [[.Import]].functions import process_blob


def test_greets_the_blob_name() -> None:
    blob = cast(func.InputStream, SimpleNamespace(name="uploads/report.csv", length=3))
    assert process_blob(blob) == "Hello, uploads/report.csv! (3 bytes)"
[[end]]`,
}
