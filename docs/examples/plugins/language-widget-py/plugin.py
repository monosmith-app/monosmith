#!/usr/bin/env python3
"""A smith language plugin for "widget" projects (a directory with a widget.json),
in Python on monosmith-plugin-sdk. Started by smith (SMITH_PLUGIN=1) it serves the
plugin protocol on stdin and stdout; started by a person, or by one of the tasks it
contributes, it is a small command line tool."""

import os
import sys

from monosmith_plugin_sdk import serve
from tool import run_tool
from widget import WidgetPlugin

if __name__ == "__main__":
    if os.environ.get("SMITH_PLUGIN") == "1":
        serve(WidgetPlugin())
    try:
        run_tool(sys.argv[1:])
    except ValueError as error:
        print(error, file=sys.stderr)
        sys.exit(1)
