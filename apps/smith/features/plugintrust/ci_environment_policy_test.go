package plugintrust

import "testing"

func TestInCI(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	for name, tc := range map[string]struct {
		env  map[string]string
		want bool
	}{
		"nothing":        {nil, false},
		"CI":             {map[string]string{"CI": "true"}, true},
		"github actions": {map[string]string{"GITHUB_ACTIONS": "true"}, true},
		"azure":          {map[string]string{"TF_BUILD": "True"}, true},
		"github false":   {map[string]string{"GITHUB_ACTIONS": "false"}, false},
	} {
		if got := InCI(env(tc.env)); got != tc.want {
			t.Errorf("%s: InCI = %v", name, got)
		}
	}
}
