package updatenotice

import "testing"

func TestInstallCommand(t *testing.T) {
	cases := []struct{ exe, goos, want string }{
		{`C:\Users\a\AppData\Roaming\npm\node_modules\monosmith\bin\smith.exe`, "windows", "npm i -g monosmith@latest"},
		{"/usr/lib/node_modules/monosmith/bin/smith", "linux", "npm i -g monosmith@latest"},
		{"/home/a/.dotnet/tools/smith", "linux", "dotnet tool update --global monosmith"},
		{"/home/a/.local/share/pipx/venvs/monosmith/bin/smith", "linux", "pipx upgrade monosmith"},
		{"/home/a/.venv/lib/python3.12/site-packages/monosmith/smith", "linux", "pip install --upgrade monosmith"},
		{"/home/a/.local/bin/smith", "linux", "curl -fsSL https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.sh | sh"},
		{`C:\Users\a\.local\bin\smith.exe`, "windows", "irm https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.ps1 | iex"},
		{"/opt/smith", "linux", "see " + ReleasesURL},
	}
	for _, c := range cases {
		if got := InstallCommand(c.exe, c.goos); got != c.want {
			t.Errorf("InstallCommand(%q) = %q, want %q", c.exe, got, c.want)
		}
	}
}
