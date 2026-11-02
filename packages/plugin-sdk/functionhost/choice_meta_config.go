package functionhost

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// HostLayer is the layer of the stack (protocol.Choice.Layer) that the function
// hosts and the answers they ask (trigger, runtime) sit on.
const HostLayer = "host"

// meta is the display name and the search words of a host or of an answer it
// asks, by choice name. The three languages share them, so a search for
// "serverless" finds the same things whatever the language.
var meta = map[string]struct {
	title   string
	aliases []string
}{
	Azure:           {"Azure Functions", []string{"serverless", "cloud", "functions", "faas", "microsoft", "azure"}},
	AWS:             {"AWS Lambda", []string{"serverless", "cloud", "functions", "faas", "amazon", "aws"}},
	GCP:             {"Google Cloud Functions", []string{"serverless", "cloud", "functions", "faas", "google", "gcp", "cloud run"}},
	"http-trigger":  {"HTTP trigger", []string{"serverless", "request", "api", "web"}},
	"timer-trigger": {"Timer trigger", []string{"serverless", "schedule", "cron"}},
	"queue-trigger": {"Queue trigger", []string{"serverless", "message", "queue"}},
	"blob-trigger":  {"Blob trigger", []string{"serverless", "storage", "file", "bucket"}},
	"python312":     {"Python 3.12", []string{"serverless", "runtime", "version"}},
	"python313":     {"Python 3.13", []string{"serverless", "runtime", "version"}},
	"python314":     {"Python 3.14", []string{"serverless", "runtime", "version"}},
	"dotnet8":       {".NET 8", []string{"serverless", "runtime", "version", "csharp"}},
	"dotnet10":      {".NET 10", []string{"serverless", "runtime", "version", "csharp"}},
}

// WithMeta fills the title, the aliases and the layer of a host choice or of an
// answer it asks, and returns the choice. A name it does not know stays as it is.
//
// Parameters: c is the choice a language plugin publishes.
func WithMeta(c protocol.Choice) protocol.Choice {
	if m, ok := meta[c.Name]; ok {
		c.Title, c.Aliases, c.Layer = m.title, m.aliases, HostLayer
	}
	return c
}
