package dialog

// aboutLogo is the anvil drawn in block characters.
var aboutLogo = []string{
	"  ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄  ",
	"▄████████████████████████▀",
	" ▀▀▀▀▀▀███████████▀▀▀▀▀▀  ",
	"       ▐█████████▌        ",
	"      ▄███████████▄       ",
	"     ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀      ",
}

// aboutText are the paragraphs about what smith is made for.
var aboutText = []string{
	"One command, smith, runs a monorepo in Node, Go, Python, Rust, .NET, Ruby, PHP, C++, Dart, the JVM and Bicep.",
	"It scaffolds projects, runs their tasks in dependency order from a cache, generates the CI pipeline from what is in the repository, and releases from conventional commits.",
	"This shell never acts by itself: every action ends as a smith command line, printed before it runs.",
}

// aboutDocs is the documentation site.
const aboutDocs = "https://monosmith-app.github.io/monosmith/README.html"
