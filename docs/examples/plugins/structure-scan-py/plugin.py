#!/usr/bin/env python3
"""A smith structure plugin in Python on monosmith-plugin-sdk whose doctor walks the files of
widget projects: it stops when smith cancels the request and it takes four scans at the same
time. It only speaks the plugin protocol."""

import os
import sys

from monosmith_plugin_sdk import serve
from scan import ScanPlugin

if __name__ == "__main__":
    if os.environ.get("SMITH_PLUGIN") == "1":
        # WIDGET_SCAN_DELAY_MS is a pause before each file, which stands in for a slow disk so a test can
        # cancel a scan in the middle.
        try:
            delay = float(os.environ.get("WIDGET_SCAN_DELAY_MS", "0")) / 1000
        except ValueError:
            delay = 0.0
        serve(ScanPlugin(max(delay, 0.0)))
    print(
        "plugin.py is a smith plugin: smith starts it and speaks the plugin protocol on its stdin and stdout",
        file=sys.stderr,
    )
    sys.exit(1)
