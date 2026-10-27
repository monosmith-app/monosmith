#!/usr/bin/env python3
"""A smith structure plugin in Python, standard library only: it adds a
layout-check task to widget projects and makes their lint wait for it.

Started by smith (SMITH_PLUGIN=1) it speaks the plugin protocol on stdin and
stdout. Started with `layout-check`, which is the command of the task it
contributes, it checks the folder it runs in. Anything meant for a human goes
to stderr: stdout carries only protocol frames.
"""
import json
import os
import sys

PROTOCOL_VERSION = 1
JUNK_DRAWERS = {"util", "utils", "helper", "helpers", "common", "shared", "misc"}
SKIPPED = {"dist", "node_modules", ".git"}
workspace_root = ""


def find_junk_drawers(directory, prefix=""):
    """The folders below directory named for nothing, as slash-separated paths."""
    found = []
    for entry in sorted(os.scandir(os.path.join(directory, prefix)), key=lambda e: e.name):
        if not entry.is_dir() or entry.name in SKIPPED:
            continue
        rel = prefix + "/" + entry.name if prefix else entry.name
        if entry.name.lower() in JUNK_DRAWERS:
            found.append(rel)
        found.extend(find_junk_drawers(directory, rel))
    return found


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
        "plugin": {"name": "widget-layout", "version": "0.1.0"},
        "kind": "convention",
        "roles": ["structure"],
        "languages": ["widget"],  # a convention that names none applies to every language
        # Only what the plugin implements; smith never calls anything else.
        "capabilities": {"contribute": True, "doctor": True},
    }


def contribute(params):
    command = [sys.executable, os.path.abspath(__file__), "layout-check"]
    projects = {}
    for project in params["projects"]:
        projects[project["name"]] = {
            "tasks": {
                "layout-check": {
                    "command": command,
                    "cache": True,
                    "description": "fail on folders named util, helpers, common, shared or misc",
                }
            },
            # Ignored for a project that has no lint once every plugin has contributed.
            "dependsOn": {"lint": ["layout-check"]},
        }
    return {"projects": projects}


def doctor(params):
    findings = []
    for project in params["projects"]:
        readme = os.path.join(workspace_root, *project["root"].split("/"), "README.md")
        ok = os.path.exists(readme)
        finding = {"check": project["name"] + " has a README.md", "ok": ok}
        if not ok:
            finding["detail"] = project["root"] + "/README.md does not exist"
            finding["remedy"] = "write " + project["root"] + "/README.md: what the project is for"
            finding["warning"] = True  # reported, but does not fail smith doctor
        findings.append(finding)
    return {"findings": findings}


HANDLERS = {
    "initialize": initialize,
    "contribute": contribute,
    "doctor": doctor,
    "shutdown": lambda params: {},
}


def serve():
    # Use the binary streams so Windows does not translate newlines.
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


def main():
    if os.environ.get("SMITH_PLUGIN") == "1":
        serve()
    elif sys.argv[1:] == ["layout-check"]:
        found = find_junk_drawers(".")
        if found:
            print("folders named for nothing: %s (name them for what they do)" % ", ".join(found), file=sys.stderr)
            sys.exit(1)
        print("no junk-drawer folders")
    else:
        print("usage: plugin.py layout-check (smith runs this program as a plugin)", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
