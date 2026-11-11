package cachecommand

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
)

// pushBanner is the first line of `smith cache push`.
func pushBanner(store string, dryRun bool) string {
	if dryRun {
		return fmt.Sprintf("Asking %s which local entries it lacks (dry run: nothing is uploaded).\n", store)
	}
	return fmt.Sprintf("Uploading the local entries that %s lacks.\n", store)
}

// pushStepLine is the line of one entry: "  ✓ web:build  uploaded (3 files, 1.2 MB)".
// An entry the remote holds already has none ("").
func pushStepLine(step remotesync.PushStep) string {
	task := step.Task
	if task == "" {
		task = step.Key[:min(len(step.Key), 12)]
	}
	size := fmt.Sprintf("%d %s, %s", step.Blobs, plural(step.Blobs, "file", "files"), formatSize(step.Bytes))
	switch step.Outcome {
	case remotesync.StepUploaded:
		return fmt.Sprintf("  %s %s  uploaded (%s)\n", marks.Passed, task, size)
	case remotesync.StepWouldUpload:
		return fmt.Sprintf("  ○ %s  would upload (%s)\n", task, size)
	case remotesync.StepFailed:
		return fmt.Sprintf("  %s %s  not uploaded: %s\n", marks.Failed, task, step.Error)
	}
	return ""
}

// pushClosing is the summary line of `smith cache push`, and the remedy when
// entries failed.
func pushClosing(r remotesync.PushResult) string {
	var b strings.Builder
	there := fmt.Sprintf("%d already on the remote", r.Present)
	size := fmt.Sprintf("%d %s, %s", r.Blobs, plural(r.Blobs, "file", "files"), formatSize(r.Bytes))
	switch {
	case r.Entries == 0:
		fmt.Fprintf(&b, "%s The local cache holds no entry: there is nothing to upload.\n", marks.Passed)
	case r.DryRun:
		fmt.Fprintf(&b, "%s Dry run: %d of %d %s would be uploaded to %s (%s); %s. Nothing was uploaded.\n",
			marks.Passed, r.Uploaded, r.Entries, plural(r.Entries, "entry", "entries"), r.Store, size, there)
		if r.Uploaded > 0 {
			b.WriteString("→ run smith cache push to upload them\n")
		}
	case r.Failed > 0:
		fmt.Fprintf(&b, "%s Uploaded %d of %d %s to %s (%s); %s; %d failed.\n",
			marks.Failed, r.Uploaded, r.Entries, plural(r.Entries, "entry", "entries"), r.Store, size, there, r.Failed)
	case r.Uploaded == 0:
		fmt.Fprintf(&b, "%s %s holds all %d local %s: nothing to upload.\n", marks.Passed, r.Store, r.Entries, plural(r.Entries, "entry", "entries"))
	default:
		fmt.Fprintf(&b, "%s Uploaded %d %s to %s (%s); %s.\n",
			marks.Passed, r.Uploaded, plural(r.Uploaded, "entry", "entries"), r.Store, size, there)
	}
	return b.String()
}
