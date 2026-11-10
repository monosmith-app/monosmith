# Posting 2.11.1: findings

Source: a real run. Screens show an IP address: do not publish them. The in-flight state of a slow request and a delete confirmation were not captured.

## Layout
- Top bar: method dropdown, URL, status pill, latency strip and Send.
- Left: Collection, with Requests and History tabs.
- Right top: request tabs (Headers, Body, Path, Query, Auth, Info, Scripts, Options).
- Right bottom: response tabs (Body, Headers, Cookies, Scripts, Trace) with a status pill, size and time.
- The focused pane has a bright border.

## Navigation
`Ctrl+O` opens a jump overlay. `Ctrl+T` method, `Ctrl+J` send, `Ctrl+S` save, `Ctrl+N` new. `Ctrl+P` opens the command palette.

## Discoverability
- The footer shows keys and changes with focus. It gained `/ Search  d Dupe  Delete` when the collection was focused.
- The palette lists commands with a description each.
- F1 opens "Focused Widget Help" and all keybindings, with "Press ESC to dismiss".

## Feedback
- An orange toast sits bottom-right with a title and a message, for example "Request URL is missing an 'http://' or 'https://' protocol" and "All connection attempts failed".
- Status pills: "200 OK" green, "500 Internal Server Error" orange.
- A JSON body has line numbers. "No headers" shows a hatched empty fill.
- Bug: after a failed send the status badge kept the `200` of the previous request.

## Mouse
A click on a collection row selects the request. A click on Send sends. Tab clicks and hover were not verified.

## What smith should copy
A context-sensitive footer. A palette with a description per command. Focused-widget help. The toast style for errors. A status pill by result class. Empty-state text with a hatch fill.

## What to avoid
A stale status badge. A footer that is cut off on a narrow terminal (it stopped at `sh`).
