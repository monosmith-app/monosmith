package cachesetup

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Item is one thing a cloud lists: a subscription, an account, a container, a
// bucket or a folder.
type Item struct {
	// Label is what the person reads; Value is what the next command takes.
	Label, Value string
	// Hint says more, in plain words.
	Hint string
}

// Listing is one read-only command and how to read its output. The setup
// shows Argv before it runs it.
type Listing struct {
	// Argv is the command, program first.
	Argv []string
	// Parse turns the command's standard output into the choices.
	Parse func(out string) ([]Item, error)
}

// Line is the command as a person would type it, quoting what needs it.
func (l Listing) Line() string { return CommandLine(l.Argv) }

// CommandLine joins arguments for display, quoting those with spaces.
func CommandLine(argv []string) string {
	parts := make([]string, len(argv))
	for i, a := range argv {
		if strings.ContainsAny(a, " \t\"'") {
			a = `"` + strings.ReplaceAll(a, `"`, `\"`) + `"`
		}
		parts[i] = a
	}
	return strings.Join(parts, " ")
}

// AzureSubscriptions lists the subscriptions the signed-in account can see.
func AzureSubscriptions() Listing {
	return Listing{Argv: []string{"az", "account", "list", "--output", "json"}, Parse: func(out string) ([]Item, error) {
		var rows []struct {
			Name, ID, State string
			IsDefault       bool `json:"isDefault"`
		}
		if err := json.Unmarshal([]byte(out), &rows); err != nil {
			return nil, fmt.Errorf("could not read the subscription list: %w", err)
		}
		var items []Item
		for _, r := range rows {
			hint := r.ID
			if r.IsDefault {
				hint += " (current)"
			}
			items = append(items, Item{Label: r.Name, Value: r.ID, Hint: hint})
		}
		return items, nil
	}}
}

// AzureAccounts lists the storage accounts of a subscription. An account with a
// hierarchical namespace has real folders; the others show a folder once
// something is written under it.
func AzureAccounts(subscription string) Listing {
	return Listing{Argv: append([]string{"az", "storage", "account", "list"}, subArgs(subscription, "--output", "json")...), Parse: func(out string) ([]Item, error) {
		var rows []struct {
			Name  string
			IsHNS bool `json:"isHnsEnabled"`
		}
		if err := json.Unmarshal([]byte(out), &rows); err != nil {
			return nil, fmt.Errorf("could not read the storage account list: %w", err)
		}
		var items []Item
		for _, r := range rows {
			hint := "blob storage: a folder exists once a file is written under it"
			if r.IsHNS {
				hint = "hierarchical namespace: real folders"
			}
			items = append(items, Item{Label: r.Name, Value: r.Name, Hint: hint})
		}
		return items, nil
	}}
}

// AzureContainers lists the containers of a storage account.
func AzureContainers(subscription, account string) Listing {
	return Listing{Argv: append([]string{"az", "storage", "container", "list", "--account-name", account}, subArgs(subscription, "--auth-mode", "login", "--output", "json")...), Parse: namesOf}
}

// AzureFolders lists the folders (blob prefixes) directly under a prefix.
func AzureFolders(subscription, account, container, prefix string) Listing {
	argv := append([]string{"az", "storage", "blob", "list", "--account-name", account, "--container-name", container}, subArgs(subscription)...)
	if prefix != "" {
		argv = append(argv, "--prefix", strings.TrimSuffix(prefix, "/")+"/")
	}
	argv = append(argv, "--delimiter", "/", "--auth-mode", "login", "--output", "json")
	return Listing{Argv: argv, Parse: func(out string) ([]Item, error) {
		items, err := namesOf(out)
		var folders []Item
		for _, it := range items {
			if strings.HasSuffix(it.Value, "/") {
				name := strings.TrimSuffix(it.Value, "/")
				folders = append(folders, Item{Label: name[strings.LastIndex(name, "/")+1:], Value: name})
			}
		}
		return folders, err
	}}
}

// namesOf reads a JSON list of objects with a name.
func namesOf(out string) ([]Item, error) {
	var rows []struct{ Name string }
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		return nil, fmt.Errorf("could not read the list: %w", err)
	}
	var items []Item
	for _, r := range rows {
		items = append(items, Item{Label: r.Name, Value: r.Name})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return items, nil
}

// S3Buckets lists the buckets the signed-in identity can see.
func S3Buckets() Listing {
	return Listing{Argv: []string{"aws", "s3api", "list-buckets", "--output", "json"}, Parse: func(out string) ([]Item, error) {
		var doc struct{ Buckets []struct{ Name string } }
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			return nil, fmt.Errorf("could not read the bucket list: %w", err)
		}
		var items []Item
		for _, b := range doc.Buckets {
			items = append(items, Item{Label: b.Name, Value: b.Name})
		}
		return items, nil
	}}
}

// S3Folders lists the folders (key prefixes) directly under a prefix.
func S3Folders(bucket, prefix string) Listing {
	argv := []string{"aws", "s3api", "list-objects-v2", "--bucket", bucket}
	if prefix != "" {
		argv = append(argv, "--prefix", strings.TrimSuffix(prefix, "/")+"/")
	}
	argv = append(argv, "--delimiter", "/", "--output", "json")
	return Listing{Argv: argv, Parse: func(out string) ([]Item, error) {
		if strings.TrimSpace(out) == "" {
			return nil, nil // aws prints nothing for an empty bucket
		}
		var doc struct{ CommonPrefixes []struct{ Prefix string } }
		if err := json.Unmarshal([]byte(out), &doc); err != nil {
			return nil, fmt.Errorf("could not read the folder list: %w", err)
		}
		var items []Item
		for _, p := range doc.CommonPrefixes {
			name := strings.TrimSuffix(p.Prefix, "/")
			items = append(items, Item{Label: name[strings.LastIndex(name, "/")+1:], Value: name})
		}
		return items, nil
	}}
}

// NeedsSignIn reads a failed cloud command's error text and tells whether the
// cause is that nobody is signed in, and which command signs in.
func NeedsSignIn(program, stderr string) (login []string, yes bool) {
	text := strings.ToLower(stderr)
	switch program {
	case "az":
		for _, m := range []string{"az login", "please run 'az login'", "aadsts", "not logged in"} {
			if strings.Contains(text, m) {
				return []string{"az", "login"}, true
			}
		}
	case "aws":
		for _, m := range []string{"unable to locate credentials", "expiredtoken", "sso login", "invalidclienttokenid", "could not be found"} {
			if strings.Contains(text, m) {
				return []string{"aws", "sso", "login"}, true
			}
		}
	}
	return nil, false
}

// SASChoice is one way to make the container's SAS token.
type SASChoice struct {
	Label, Value, Hint string
	// MaxDays is the longest validity the way allows; 0 means no limit.
	MaxDays int
}

// SASChoices are the two ways to sign a container SAS token with az.
func SASChoices() []SASChoice {
	return []SASChoice{
		{Label: "Signed in as you (user delegation)", Value: "login", MaxDays: 7,
			Hint: "no account key is used; Azure limits the token to 7 days, so you renew it every week"},
		{Label: "Signed with the account key", Value: "key", MaxDays: 0,
			Hint: "the token can last months; it is signed with the account key, so rotating that key voids the token"},
	}
}

// GenerateSAS is the command that makes a SAS token for a container with read,
// write, delete, create and list permission (delete for `smith cache clear
// --remote`).
//
// Parameters: way is a SASChoice value; expiry is when the token ends.
func GenerateSAS(subscription, account, container, way string, expiry time.Time) []string {
	argv := append([]string{"az", "storage", "container", "generate-sas", "--account-name", account, "--name", container}, subArgs(subscription,
		"--permissions", "rwdcl", "--expiry", expiry.UTC().Format("2006-01-02T15:04Z"), "--auth-mode", way)...)
	if way == "login" {
		argv = append(argv, "--as-user")
	}
	return append(argv, "--output", "tsv")
}

// subArgs puts --subscription before the rest when there is a subscription.
func subArgs(subscription string, rest ...string) []string {
	if subscription == "" {
		return rest
	}
	return append([]string{"--subscription", subscription}, rest...)
}
