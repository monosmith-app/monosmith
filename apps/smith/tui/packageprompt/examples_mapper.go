package packageprompt

// ExampleNames are package names a person of that language would recognize,
// to show in the empty box. An unknown language gets the node examples, which
// are the most common.
func ExampleNames(language string) string {
	switch language {
	case "python":
		return "requests  pydantic>=2  pytest" // words:literal
	case "go":
		return "github.com/spf13/cobra  golang.org/x/sync@latest" // words:literal
	case "rust":
		return "serde  tokio  clap" // words:literal
	case "dotnet":
		return "Newtonsoft.Json  Serilog" // words:literal
	case "jvm":
		return "com.google.guava:guava  org.slf4j:slf4j-api" // words:literal
	case "dart":
		return "http  provider  intl" // words:literal
	case "ruby":
		return "rails  rspec  nokogiri" // words:literal
	case "php":
		return "monolog/monolog  guzzlehttp/guzzle" // words:literal
	}
	return "zod  react@19  @types/node" // words:literal
}
