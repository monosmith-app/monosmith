package pane

import "github.com/monosmith-app/monosmith/apps/smith/tui/theme"

// DialogBody fits the body of a dialog into h rows of w cells: the lines from
// top on, with the "↑ more" and "↓ more" markers (WindowMarked) and the shared
// scrollbar in the last cell when the body does not fit, and the foot lines
// pinned under them, so the keys or buttons a dialog ends with never scroll
// away. Every dialog draws its scrolling part with it, and maps a click with
// BodyLine over the same numbers, so what is drawn and what is clicked agree.
//
// Parameters: th is the theme; body are the lines that scroll; foot are the
// lines that stay at the bottom; w is the width of a row; h is the rows there
// are; top is the first body line wanted.
// Returns: exactly h rows (fewer only when h is below the foot), padded with
// empty rows when the body is short.
func DialogBody(th theme.Theme, body, foot []string, w, h, top int) []string {
	room := max(h-len(foot), 0)
	shown := body
	if len(body) > room {
		top = ClampTop(top, len(body), room)
		shown, _ = WindowMarked(th, body, nil, top, room)
		shown = Scrollbar(th, shown, w, top, len(body))
	}
	out := make([]string, 0, h)
	for _, l := range shown {
		if len(out) >= room {
			break
		}
		out = append(out, l)
	}
	for len(out) < room {
		out = append(out, "")
	}
	return append(out, foot[:min(len(foot), h)]...)
}

// BodyRoom is the rows the scrolling body of a dialog gets: the rows h of the
// dialog's inside less the foot rows pinned under the body.
//
// Parameters: h is the rows inside the dialog; foot is how many rows stay at the bottom.
// Returns: the rows of the body, at least 1.
func BodyRoom(h, foot int) int { return max(h-foot, 1) }

// FollowLines is the first line of a window of h rows over total lines that
// shows the lines from..to-1 (a choice that wraps over several lines), moving
// the window from top as little as it can and never under a "more" marker.
// When the lines are taller than the window, their first line wins.
//
// Parameters: top is where the window starts now; from and to are the lines to
// show, to excluded; total is the body's length; h is the rows of the window.
// Returns: the new first line; 0 when everything fits.
func FollowLines(top, from, to, total, h int) int {
	top = KeepVisible(top, max(to-1, from), total, h)
	return KeepVisible(top, from, total, h)
}

// BodyLine is the body line drawn on row y of a window of h rows that starts
// at top over total lines, as DialogBody draws it: the inverse a click and a
// hover use.
//
// Parameters: y is the row inside the window (0 is its first row); top, total
// and h are as for DialogBody.
// Returns: the body line, or -1 when y is outside the window, past the body or
// on a "more" marker.
func BodyLine(y, top, total, h int) int {
	if y < 0 || y >= h || y >= total {
		return -1
	}
	if total <= h {
		return y
	}
	top = ClampTop(top, total, h)
	if (y == 0 && top > 0) || (y == h-1 && top+h < total) {
		return -1
	}
	return top + y
}
