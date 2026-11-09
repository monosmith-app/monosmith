package functions

// gcpTemplates are the files of a Google Cloud (Cloud Run) function. The
// Functions Framework loads main.py, which imports the function of the functions
// slice; the framework is a dependency of the app, so no extra tool is installed.
var gcpTemplates = map[string]string{
	"[[.Dir]]/main.py": `"""The entry point the Functions Framework loads."""

from src.[[.Import]].functions import [[.Handler]]

__all__ = ["[[.Handler]]"]
`,
	"[[.Dir]]/requirements.txt": `functions-framework>=3.10
`,
	"[[.Dir]]/start.py": `"""Runs the function locally with the Functions Framework, which the app depends on."""

import subprocess
import sys

command = [
    sys.executable,
    "-m",
    "functions_framework",
    "--target=[[.Handler]]",[[if ne .Trigger "http"]]
    "--signature-type=cloudevent",[[end]]
]
sys.exit(subprocess.call(command))
`,
	"[[.Dir]]/src/[[.Import]]/functions/greeting_handler.py": `[[if eq .Trigger "http"]]import functions_framework
from flask import Request

from ..greeting import greet


@functions_framework.http
def hello(request: Request) -> dict[str, str]:
    """Answer GET /?name=Ada with a greeting as JSON."""
    return {"message": greet(request.args.get("name", "world")).message}
[[else if eq .Trigger "queue"]]import base64

import functions_framework
from cloudevents.http import CloudEvent

from ..greeting import greet


@functions_framework.cloud_event
def process_message(cloud_event: CloudEvent) -> None:
    """Greet the text of a Pub/Sub message (base64 in the event data)."""
    data = cloud_event.data or {}
    text = base64.b64decode(data.get("message", {}).get("data", "")).decode("utf-8")
    print(greet(text or "world").message)
[[else]]import functions_framework
from cloudevents.http import CloudEvent

from ..greeting import greet


@functions_framework.cloud_event
def process_object(cloud_event: CloudEvent) -> None:
    """Greet the name of a new Cloud Storage object."""
    data = cloud_event.data or {}
    print(greet(data.get("name", "object")).message)
[[end]]`,
	"[[.Dir]]/tests/test_greeting_handler.py": `[[if eq .Trigger "http"]]from types import SimpleNamespace
from typing import cast

from flask import Request

from [[.Import]].functions import hello


def test_greets_the_name_in_the_query() -> None:
    request = cast(Request, SimpleNamespace(args={"name": "Ada"}))
    assert hello(request) == {"message": "Hello, Ada!"}


def test_greets_the_world_without_a_name() -> None:
    request = cast(Request, SimpleNamespace(args={}))
    assert hello(request) == {"message": "Hello, world!"}
[[else if eq .Trigger "queue"]]import base64

import pytest
from cloudevents.http import CloudEvent

from [[.Import]].functions import process_message


def test_greets_the_text_of_the_message(capsys: pytest.CaptureFixture[str]) -> None:
    attributes = {"type": "google.cloud.pubsub.topic.v1.messagePublished", "source": "//pubsub"}
    data = {"message": {"data": base64.b64encode(b"Ada").decode()}}
    process_message(CloudEvent(attributes, data))
    assert capsys.readouterr().out == "Hello, Ada!\n"
[[else]]import pytest
from cloudevents.http import CloudEvent

from [[.Import]].functions import process_object


def test_greets_the_name_of_the_object(capsys: pytest.CaptureFixture[str]) -> None:
    attributes = {"type": "google.cloud.storage.object.v1.finalized", "source": "//storage"}
    process_object(CloudEvent(attributes, {"name": "report.csv"}))
    assert capsys.readouterr().out == "Hello, report.csv!\n"
[[end]]`,
}
