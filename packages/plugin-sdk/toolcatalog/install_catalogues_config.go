package toolcatalog

// installCatalogues are the pages each install entry is checked against, by
// package manager. An entry is shipped only when its package is found there on
// the date in checked. An entry that cannot be found is removed from the
// install options: the official page option stays.
var installCatalogues = map[string]string{
	"winget":    "https://github.com/microsoft/winget-pkgs",
	"brew":      "https://formulae.brew.sh/formula/",
	"brew-cask": "https://formulae.brew.sh/cask/",
	"apt":       "https://packages.debian.org/ and https://packages.ubuntu.com/",
	"dnf":       "https://packages.fedoraproject.org/",
	"pacman":    "https://archlinux.org/packages/",
	"npm":       "https://www.npmjs.com/",
	"uv":        "https://docs.astral.sh/uv/concepts/python-versions/",
	"fnm":       "https://github.com/Schniz/fnm",
	"rustup":    "https://rust-lang.github.io/rustup/",
}
