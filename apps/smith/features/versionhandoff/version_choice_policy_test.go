package versionhandoff

import "testing"

func TestPinInReadsTheWorkspacePinAsTheLaunchersDo(t *testing.T) {
	cases := map[string]string{
		"[workspace]\nsmith = \"2.50.0\"\n":                       "2.50.0",
		"[workspace]\r\nsmith = '3.0.0-beta.1' # pinned\r\n":      "3.0.0-beta.1",
		"[ workspace ]\nname = \"x\"\nsmith=\"2.1.0\"":            "2.1.0",
		"[workspace]\nsmith = \"latest\"\n":                       "",
		"[workspace]\nsmith = \"^2\"\n":                           "",
		"[project]\nsmith = \"2.50.0\"\n":                         "",
		"# [workspace]\nsmith = \"2.50.0\"\n":                     "",
		"[workspace]\n# smith = \"2.50.0\"\n":                     "",
		"[workspace]\nname = \"x\"\n[release]\nsmith = \"1.0.0\"": "",
		"": "",
	}
	for text, want := range cases {
		if got := PinIn(text); got != want {
			t.Errorf("PinIn(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestChooseVersionPutsSmithVersionBeforeThePin(t *testing.T) {
	pin := func() string { return "2.40.0" }
	cases := []struct {
		env, want string
		bad       bool
	}{
		{"", "2.40.0", false},
		{"  ", "2.40.0", false},
		{"2.51.0", "2.51.0", false},
		{" 3.0.0-rc.1 ", "3.0.0-rc.1", false},
		{"latest", "", false},
		{"next", "", true},
		{"2.5", "", true},
	}
	for _, c := range cases {
		got, err := ChooseVersion(c.env, pin)
		if got != c.want || (err != nil) != c.bad {
			t.Errorf("ChooseVersion(%q) = %q, %v; want %q, error %v", c.env, got, err, c.want, c.bad)
		}
	}
	if _, err := ChooseVersion("2.5", pin); err == nil || err.Error() != `SMITH_VERSION is "2.5": use a version such as 2.50.0, or latest` {
		t.Errorf("the message differs from the launchers: %v", err)
	}
}

func TestCompareVersionsSortsAsTheLaunchersDo(t *testing.T) {
	cases := []struct {
		a, b string
		sign int
	}{
		{"2.50.0", "2.50.0", 0},
		{"2.9.0", "2.10.0", -1},
		{"3.0.0", "2.99.99", 1},
		{"3.0.0-beta.1", "3.0.0", -1},
		{"3.0.0", "3.0.0-beta.1", 1},
		{"3.0.0-beta.1", "3.0.0-beta.2", -1},
		{"3.0.0-alpha", "3.0.0-beta", -1},
	}
	for _, c := range cases {
		got := CompareVersions(c.a, c.b)
		if (got > 0) != (c.sign > 0) || (got < 0) != (c.sign < 0) {
			t.Errorf("CompareVersions(%s, %s) = %d, want sign %d", c.a, c.b, got, c.sign)
		}
	}
}
