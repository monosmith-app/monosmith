package guide

import (
	"slices"
	"strings"
	"testing"
)

const packageOffersJSON = `[
 {"axis":"framework","name":"react","language":"node","plugin":"vite","roles":["scaffold","toolchain"],"needs":["lint","test"],"kinds":["app"]},
 {"axis":"framework","name":"vue","language":"node","plugin":"vite","roles":["scaffold","toolchain"],"needs":["lint","test"],"kinds":["app"]},
 {"axis":"shell","name":"tauri","language":"node","plugin":"tauri","layer":"shell","roles":["toolchain"],"kinds":["app"]},
 {"axis":"lint","name":"eslint","language":"node","plugin":"eslint","roles":["toolchain"],"default":true},
 {"axis":"test","name":"jest","language":"node","plugin":"jest","roles":["toolchain"]},
 {"axis":"test","name":"vitest","language":"node","plugin":"vitest","roles":["toolchain"],"default":true}
]`

func nodeOffers(t *testing.T) []ChoiceOffer {
	t.Helper()
	offers, err := ParseChoices(packageOffersJSON)
	if err != nil {
		t.Fatal(err)
	}
	return offers
}

func TestAPackageIsOfferedOnlyTheChoicesThatFitIt(t *testing.T) {
	var names []string
	for _, o := range fitting(nodeOffers(t), "package") {
		names = append(names, o.Name)
	}
	if want := []string{"eslint", "jest", "vitest"}; !slices.Equal(names, want) {
		t.Errorf("a package gets %v, want %v", names, want)
	}
	if got := fitting(nodeOffers(t), "app"); len(got) != 6 {
		t.Errorf("an app keeps every choice, got %d", len(got))
	}
}

func TestTheTickedSetIsCheckedBeforeTheFormIsLeft(t *testing.T) {
	offers := nodeOffers(t)
	check := picksCheck("node", offers, map[string]string{"web": "framework"})
	for name, tc := range map[string]struct {
		answers map[string][]string
		value   string
		want    string // "" means accepted
	}{
		"nothing ticked uses the defaults": {map[string][]string{"kind": {"package"}}, "", ""},
		"lint and test on a package":       {map[string][]string{"kind": {"package"}}, "eslint,jest", ""},
		"a framework on a package":         {map[string][]string{"kind": {"package"}}, "react,jest", "react makes only apps, and this is a package: untick it"},
		"a shell alone":                    {map[string][]string{"kind": {"app"}}, "eslint,tauri", "no choice in eslint, tauri writes the project: add one of react, vue, or untick everything to start from the stack's own defaults"},
		"a website with lint and test":     {map[string][]string{"kind": {"app"}, outcomeKey: {"web"}}, "eslint,jest", "needs a framework: tick one of react, vue"},
		"a website with a framework":       {map[string][]string{"kind": {"app"}, outcomeKey: {"web"}}, "react", ""},
		"two frameworks":                   {map[string][]string{"kind": {"app"}}, "react,vue", "react and vue are both on the framework axis: keep one"},
	} {
		got := check(tc.answers, tc.value)
		if tc.want == "" && got != "" || !strings.Contains(got, tc.want) {
			t.Errorf("%s: got %q, want %q", name, got, tc.want)
		}
		if strings.Contains(got, "--with") {
			t.Errorf("%s: the form does not talk of flags: %q", name, got)
		}
	}
}
