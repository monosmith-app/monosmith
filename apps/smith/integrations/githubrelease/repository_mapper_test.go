package githubrelease

import "testing"

func TestParseRepository(t *testing.T) {
	for _, in := range []string{
		"https://github.com/monosmith-app/monosmith.git",
		"https://github.com/monosmith-app/monosmith",
		"https://x-access-token:abc@github.com/monosmith-app/monosmith.git",
		"git@github.com:monosmith-app/monosmith.git",
		"ssh://git@github.com/monosmith-app/monosmith.git",
		"monosmith-app/monosmith",
	} {
		r, err := ParseRepository(in)
		if err != nil || r.String() != "monosmith-app/monosmith" {
			t.Errorf("%s: %v %v", in, r, err)
		}
	}
	for _, in := range []string{"", "https://github.com/onlyowner", "C:/local/bare.git/x/y"} {
		if _, err := ParseRepository(in); err == nil {
			t.Errorf("%q must be rejected", in)
		}
	}
}
