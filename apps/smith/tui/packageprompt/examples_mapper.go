package packageprompt

// ExampleNames are package names a person of that language would recognize,
// to show in the empty box. An unknown language gets the node examples, which
// are the most common.
func ExampleNames(language string) string {
	switch language {
	case "python":
		return "requests  pydantic>=2  pytest"
	case "go":
		return "github.com/spf13/cobra  golang.org/x/sync@latest"
	case "rust":
		return "serde  tokio  clap"
	case "dotnet":
		return "Newtonsoft.Json  Serilog"
	case "jvm":
		return "com.google.guava:guava  org.slf4j:slf4j-api"
	case "dart":
		return "http  provider  intl"
	case "ruby":
		return "rails  rspec  nokogiri"
	case "php":
		return "monolog/monolog  guzzlehttp/guzzle"
	}
	return "zod  react@19  @types/node"
}
