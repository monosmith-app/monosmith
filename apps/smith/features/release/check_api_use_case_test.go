package release

import (
	"context"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// surfaces answers like a plugin: the names per ref, "" being the working tree.
func surfaces(byRef map[string][]string, unreadable ...string) surfaceReader {
	return func(_ context.Context, item Item, ref string) (protocol.APISurfaceResult, bool) {
		if len(unreadable) > 0 && unreadable[0] == item.Project {
			return protocol.APISurfaceResult{}, false
		}
		names, ok := byRef[ref]
		return protocol.APISurfaceResult{Names: names}, ok
	}
}

func tagOfItem(item Item) string { return item.Project + "@" + item.Current }

func TestCheckItemsWarnsOnTheItemWhoseAPIShrank(t *testing.T) {
	plan := &Plan{Items: []Item{
		{Project: "ui", Current: "1.0.0", Next: "1.0.1", Level: "patch"},
		{Project: "other", Current: "1.0.0", Next: "1.0.1", Level: "patch"},
	}}
	read := surfaces(map[string][]string{"ui@1.0.0": {"a", "b"}, "": {"a"}, "other@1.0.0": {"a"}})
	// "other" did not change, so only "ui" gets a warning.
	got := checkItems(context.Background(), plan, tagOfItem, read)
	if len(got) != 1 || !strings.Contains(got[0], "ui: the commits release a patch (1.0.1)") || !strings.Contains(got[0], "lost 1 export (b)") {
		t.Fatalf("warnings = %v", got)
	}
	if len(plan.Items[0].APIWarnings) != 1 || plan.Items[0].APIWarnings[0] != got[0] {
		t.Errorf("the item must carry its warning, got %v", plan.Items[0].APIWarnings)
	}
}

func TestCheckItemsSkipsWhatItCannotCompare(t *testing.T) {
	plan := &Plan{Items: []Item{
		{Project: "new", Current: "", Next: "0.1.0", Level: "minor"},
		{Project: "forced", Current: "1.0.0", Next: "3.0.0", Level: "forced", Forced: true},
		{Project: "blind", Current: "1.0.0", Next: "1.0.1", Level: "patch"},
	}}
	read := surfaces(map[string][]string{"blind@1.0.0": {"a"}, "": {}}, "blind")
	if got := checkItems(context.Background(), plan, tagOfItem, read); len(got) != 0 {
		t.Errorf("no item can be compared, got %v", got)
	}
	for _, item := range plan.Items {
		if len(item.APIWarnings) != 0 {
			t.Errorf("%s must carry no warning, got %v", item.Project, item.APIWarnings)
		}
	}
}

func TestOnlyTheErrorModeTurnsAWarningIntoAnError(t *testing.T) {
	findings := []string{"ui: the commits release a patch (1.0.1)"}
	for _, mode := range []string{APICheckWarn, APICheckOff, ""} {
		if err := apiCheckError(mode, findings); err != nil {
			t.Errorf("mode %q must never fail the release, got %v", mode, err)
		}
	}
	err := apiCheckError(APICheckError, findings)
	if err == nil || !strings.Contains(err.Error(), "ui: the commits release a patch") || !strings.Contains(err.Error(), `api_check = "error"`) {
		t.Errorf("error mode must fail and say why, got %v", err)
	}
	if err := apiCheckError(APICheckError, nil); err != nil {
		t.Errorf("no finding, no error, got %v", err)
	}
}

func TestReportShowsTheAPIWarningInThePreviewComment(t *testing.T) {
	item := Item{Project: "ui", Current: "1.0.0", Next: "1.0.1", Level: "patch", Reason: "1 releasing commit(s)", APIWarnings: []string{"ui: the commits release a patch"}}
	md := Report{DryRun: true, Results: []ReportResult{{Item: item, Status: "planned"}}}.Markdown()
	if !strings.Contains(md, "**Warning, API check:** ui: the commits release a patch") {
		t.Errorf("the preview comment must carry the warning:\n%s", md)
	}
}
