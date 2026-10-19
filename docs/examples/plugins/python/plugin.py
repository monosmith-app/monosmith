#!/usr/bin/env python3
"""A minimal smith language plugin in Python, standard library only.

It teaches smith one made-up project type: every directory that holds a
widget.json is a project, and each gets a "hello" task.

The protocol is JSON-RPC 2.0 on stdin and stdout, each message framed with a
"Content-Length: <bytes>" header, a blank line and the JSON body (the framing
of the Language Server Protocol). Anything for a human goes to stderr.
"""
import json
import os
import sys

PROTOCOL_VERSION = 1
workspace_root = ""


def read_message(stream):
    """Return the next message as a dict, or None when smith closed the pipe."""
    length = None
    while True:
        line = stream.readline()
        if not line:
            return None
        line = line.decode("ascii").strip()
        if line == "":
            break
        name, _, value = line.partition(":")
        if name.lower() == "content-length":
            length = int(value.strip())
    if length is None:
        return None
    return json.loads(stream.read(length))


def write_message(stream, message):
    body = json.dumps(message).encode("utf-8")
    stream.write(b"Content-Length: %d\r\n\r\n" % len(body))
    stream.write(body)
    stream.flush()


def initialize(params):
    global workspace_root
    workspace_root = params["workspaceRoot"]
    if PROTOCOL_VERSION not in params["protocolVersions"]:
        raise RuntimeError("smith speaks %s, this plugin speaks %d"
                           % (params["protocolVersions"], PROTOCOL_VERSION))
    return {
        "protocolVersion": PROTOCOL_VERSION,
        "plugin": {"name": "widget", "version": "0.1.0"},
        "kind": "language",
        "roles": ["language"],  # not also scaffold: this plugin does not generate
        "languages": ["widget"],
        "manifests": ["widget.json"],
        # Only what the plugin implements; smith never calls anything else.
        "capabilities": {"discover": True},
    }


def discover(params):
    projects = []
    for manifest in params["manifests"]:  # workspace-relative, forward slashes
        path = os.path.join(workspace_root, *manifest.split("/"))
        with open(path, encoding="utf-8") as f:
            data = json.load(f)
        if not data.get("name"):
            continue  # not ours
        projects.append({
            "name": data["name"],
            "root": os.path.dirname(manifest).replace("\\", "/"),
            "language": "widget",
            "kind": data.get("kind") or "lib",
            "version": data.get("version", ""),
            "exports": [data["name"]],
            "tasks": {
                "hello": {
                    "command": ["git", "--version"],
                    "inputs": ["widget.json"],
                    "cache": True,
                    "description": "say hello, as a widget does",
                }
            },
        })
    return {"projects": projects}


HANDLERS = {
    "initialize": initialize,
    "discover": discover,
    "shutdown": lambda params: {},
}


def main():
    stdin, stdout = sys.stdin.buffer, sys.stdout.buffer
    while True:
        request = read_message(stdin)
        if request is None:
            return
        if "id" not in request:
            continue  # a notification: no reply
        handler = HANDLERS.get(request["method"])
        response = {"jsonrpc": "2.0", "id": request["id"]}
        if handler is None:
            response["error"] = {"code": -32601, "message": "method not supported: " + request["method"]}
        else:
            try:
                response["result"] = handler(request.get("params") or {})
            except Exception as e:  # -32000: a failure the user must act on
                response["error"] = {"code": -32000, "message": str(e)}
        write_message(stdout, response)
        if request["method"] == "shutdown":
            return


if __name__ == "__main__":
    main()
