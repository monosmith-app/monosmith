package doctorfix

import "testing"

func TestDisplayCommand(t *testing.T) {
	for _, tc := range []struct {
		argv []string
		want string
	}{
		{[]string{`C:\Users\me\AppData\smith.exe`, "adopt", "--init"}, "smith adopt --init"},
		{[]string{"/usr/local/bin/smith", "agent", "install"}, "smith agent install"},
		{[]string{"npm", "install"}, "npm install"},
		{nil, ""},
	} {
		if got := displayCommand(tc.argv); got != tc.want {
			t.Errorf("displayCommand(%v) = %q, want %q", tc.argv, got, tc.want)
		}
	}
}
