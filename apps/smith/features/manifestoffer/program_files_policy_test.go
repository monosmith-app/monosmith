package manifestoffer

import "testing"

func TestLooksLikeProgram(t *testing.T) {
	cases := map[string]bool{"main.py": true, "Program.cs": true, "src/main.ts": false, "util.ts": false, "server.js": true}
	for file, want := range cases {
		if got := LooksLikeProgram([]string{file}); got != want {
			t.Errorf("LooksLikeProgram(%s) = %v, want %v", file, got, want)
		}
	}
}
