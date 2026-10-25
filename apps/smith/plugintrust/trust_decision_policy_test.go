package plugintrust

import "testing"

func fp(command []string, digests ...string) Fingerprint {
	f := Fingerprint{Command: command}
	for i, d := range digests {
		f.Executables = append(f.Executables, Executable{Path: "/p/" + string(rune('a'+i)), SHA256: d})
	}
	return f
}

func TestDecideOnAMachine(t *testing.T) {
	trusted := fp([]string{"node", "p.mjs"}, "aaaa", "bbbb")
	rec := &Record{Workspace: "/w", Plugin: "x", Fingerprint: trusted}
	for name, tc := range map[string]struct {
		facts Facts
		want  Reason
		ok    bool
	}{
		"never trusted":    {Facts{Current: trusted}, ReasonNew, false},
		"same":             {Facts{Recorded: rec, Current: trusted}, "", true},
		"command changed":  {Facts{Recorded: rec, Current: fp([]string{"node", "q.mjs"}, "aaaa", "bbbb")}, ReasonCommandChanged, false},
		"script changed":   {Facts{Recorded: rec, Current: fp([]string{"node", "p.mjs"}, "aaaa", "cccc")}, ReasonFileChanged, false},
		"interpreter":      {Facts{Recorded: rec, Current: fp([]string{"node", "p.mjs"}, "dddd", "bbbb")}, ReasonFileChanged, false},
		"fewer files":      {Facts{Recorded: rec, Current: fp([]string{"node", "p.mjs"}, "aaaa")}, ReasonFileChanged, false},
		"a pin is not own": {Facts{Current: trusted, Pin: "bbbb"}, ReasonNew, false},
	} {
		t.Run(name, func(t *testing.T) {
			v := Decide(tc.facts)
			if v.Trusted != tc.ok || v.Reason != tc.want {
				t.Fatalf("got %+v", v)
			}
		})
	}
}

func TestDecideInCIWantsAMatchingPin(t *testing.T) {
	current := fp([]string{"node", "p.mjs"}, "aaaa", "BBBB")
	rec := &Record{Fingerprint: current}
	for name, tc := range map[string]struct {
		pin  string
		want Reason
		ok   bool
	}{
		"no pin":         {"", ReasonNotPinned, false},
		"wrong pin":      {"cccc", ReasonPinMismatch, false},
		"the program":    {"aaaa", ReasonPinMismatch, false},
		"the script":     {"bbbb", "", true},
		"prefix, spaces": {" sha256:BBBB ", "", true},
	} {
		t.Run(name, func(t *testing.T) {
			// A recorded decision means nothing in CI.
			v := Decide(Facts{Recorded: rec, Current: fp(current.Command, "aaaa", "bbbb"), Pin: tc.pin, InCI: true})
			if v.Trusted != tc.ok || v.Reason != tc.want {
				t.Fatalf("got %+v", v)
			}
		})
	}
	if v := Decide(Facts{Recorded: rec, Current: current, InCI: true}); v.Trusted {
		t.Fatalf("a recorded decision trusted a plugin in CI: %+v", v)
	}
}
