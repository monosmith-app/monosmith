package cachesetup

// Answer is one option of a question, with what it does in plain words.
type Answer struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Hint  string `json:"hint"`
}

// Question is one decision of the setup, for an interface that asks it (the
// terminal of `smith cache setup` asks the same ones).
type Question struct {
	Key     string   `json:"key"`
	Prompt  string   `json:"prompt"`
	Answers []Answer `json:"answers,omitempty"`
	// Free is true when the answer is typed text.
	Free bool `json:"free,omitempty"`
	// After is the answer that makes this question appear ("" for always).
	After string `json:"after,omitempty"`
}

// The storage kinds, as the scheme of `[cache] remote`.
const (
	KindAzure = "azblob"
	KindS3    = "s3"
	KindHTTPS = "https"
)

// Questions are the decisions of the setup in the order they come. The
// choices inside a cloud (subscription, account, container, folder, bucket)
// are listed by the cloud's own program at the time, so they appear here as
// questions without fixed answers.
func Questions() []Question {
	return []Question{
		{Key: "kind", Prompt: "What kind of remote storage? Google Cloud Storage is not supported.", Answers: []Answer{
			{Label: "Azure Blob", Value: KindAzure, Hint: "a container in an Azure storage account (azblob://account/container/prefix)"},
			{Label: "S3", Value: KindS3, Hint: "a bucket in Amazon S3 or an S3-compatible store (s3://bucket/prefix)"},
			{Label: "HTTPS server", Value: KindHTTPS, Hint: "any server that takes PUT, GET and HEAD (https://host/path)"},
		}},
		{Key: "subscription", After: KindAzure, Prompt: "Which Azure subscription? (az account list)"},
		{Key: "account", After: KindAzure, Prompt: "Which storage account? (az storage account list)"},
		{Key: "container", After: KindAzure, Prompt: "Which container? (az storage container list)"},
		{Key: "bucket", After: KindS3, Prompt: "Which bucket? (aws s3api list-buckets)"},
		{Key: "folder", Prompt: "Which base folder (prefix) inside it? Then the sub-folder for this workspace, named after it."},
		{Key: "address", After: KindHTTPS, Prompt: "The address of the cache server", Free: true},
		{Key: "write", Prompt: "Write [cache] remote to smith.toml? The exact change is shown first.", Answers: []Answer{
			{Label: "Yes, write it", Value: "yes", Hint: "only the setting changes; comments stay"},
			{Label: "No, change nothing", Value: "no", Hint: "the address is printed so you can set it yourself"},
		}},
		{Key: "key", Prompt: "Make the shared key (SMITH_CACHE_KEY) now? It is copied, never shown.", Answers: []Answer{
			{Label: "Yes, make and copy a key", Value: "yes", Hint: "the key stays in memory while the screen is open"},
			{Label: "No, I have one", Value: "no", Hint: "`smith cache key` makes one later"},
		}},
	}
}
