package screenreader

import "testing"

func TestWantedFollowsTheEnvironmentVariable(t *testing.T) {
	cases := map[string]bool{"": false, "0": false, "false": false, "No": false, "off": false, "1": true, "true": true, "yes": true, "anything": true}
	for value, want := range cases {
		got := Wanted(func(name string) string {
			if name != Variable {
				t.Fatalf("looked at %q", name)
			}
			return value
		})
		if got != want {
			t.Errorf("%s=%q: %v, want %v", Variable, value, got, want)
		}
	}
}
