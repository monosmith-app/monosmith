package toolrun

import (
	"path/filepath"
	"testing"
)

func TestWingetAlreadyInstalled(t *testing.T) {
	winget := []string{"winget", "install", "--id", "Google.DartSDK", "-e"}
	cases := []struct {
		name string
		argv []string
		code int
		want bool
	}{
		{"no applicable upgrade", winget, 2316632107, true}, // 0x8A15002B, the owner's case
		{"package already installed", winget, 0x8A150061, true},
		{"another version installed", winget, 0x8A15010D, true},
		{"a newer version installed", winget, 0x8A15010E, true},
		{"upgrade not newer", winget, 0x8A15004F, true},
		{"winget.exe with a path", []string{filepath.Join("x", "winget.exe"), "install"}, 0x8A15002B, true},
		{"a real winget failure", winget, 0x8A150107, false},
		{"success is not this case", winget, 0, false},
		{"another program with the same code", []string{"npm", "install"}, 0x8A15002B, false},
		{"no command", nil, 0x8A15002B, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := WingetAlreadyInstalled(c.argv, c.code); got != c.want {
				t.Fatalf("WingetAlreadyInstalled(%v, %d) = %v, want %v", c.argv, c.code, got, c.want)
			}
		})
	}
}
