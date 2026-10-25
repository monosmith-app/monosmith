#!/usr/bin/env python3
"""A smith structure plugin in Python on monosmith-plugin-sdk: it adds a layout-check
task to widget projects and makes their lint wait for it.

Started by smith (SMITH_PLUGIN=1) it serves the plugin protocol on stdin and stdout.
Started with `layout-check`, which is the command of the task it contributes, it checks
the folder it runs in. The SDK frames the messages, derives the capabilities from the
methods of the plugin (layout.py) and keeps stdout for the protocol: anything meant
for a human goes to stderr (`print(..., file=sys.stderr)`, or a plain print once serving).
"""

import os
import sys

from layout import LayoutPlugin, find_junk_drawers
from monosmith_plugin_sdk import serve

if __name__ == "__main__":
    if os.environ.get("SMITH_PLUGIN") == "1":
        serve(LayoutPlugin())
    elif sys.argv[1:2] == ["layout-check"]:
        found = find_junk_drawers(".")
        if found:
            print(f"folders named for nothing: {', '.join(found)} (name them for what they do)", file=sys.stderr)
            sys.exit(1)
        print("no junk-drawer folders")
    else:
        print("usage: plugin.py layout-check (smith runs this program as a plugin)", file=sys.stderr)
        sys.exit(1)
