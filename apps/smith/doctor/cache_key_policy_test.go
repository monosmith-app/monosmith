package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheRemoteCacheKeyIsOnlyAWarning(t *testing.T) {
	none := func(string) string { return "" }
	if got := cacheKeyFinding("", none); len(got) != 0 {
		t.Errorf("no remote, no finding: %+v", got)
	}
	got := cacheKeyFinding("azblob://acme/smith-cache/main", none)
	if len(got) != 1 || got[0].OK || !got[0].Warning {
		t.Fatalf("a remote without its key warns and does not fail: %+v", got)
	}
	if got := cacheKeyFinding("azblob://acme/smith-cache/main", func(string) string { return "$(SMITH_CACHE_KEY)" }); got[0].OK {
		t.Error("an unexpanded Azure macro is no key")
	}
	if got := cacheKeyFinding("azblob://acme/smith-cache/main", func(name string) string {
		if name == "SMITH_CACHE_KEY" {
			return "a-shared-secret"
		}
		return ""
	}); !got[0].OK {
		t.Errorf("with the key set the check passes: %+v", got)
	}
}

func TestCacheKeyFindingNamesTheProblemOnlyWhenItFails(t *testing.T) {
	failed := cacheKeyFinding("s3://b", func(string) string { return "" })[0]
	if failed.OK || failed.Problem != "SMITH_CACHE_KEY is not set, so the remote cache s3://b is off" {
		t.Errorf("failed: %+v", failed)
	}
	passed := cacheKeyFinding("s3://b", func(string) string { return "k" })[0]
	if !passed.OK || passed.Problem != "" {
		t.Errorf("passed: %+v", passed)
	}
}

func TestAnAzureBlobRemoteWithAKeyButNoSASTokenWarns(t *testing.T) {
	env := map[string]string{"SMITH_CACHE_KEY": "abc"}
	findings := cacheKeyFinding("azblob://acct/cache/p", func(k string) string { return env[k] })
	var token *protocol.Finding
	for i := range findings {
		if findings[i].Check == "the remote cache has its SAS token" {
			token = &findings[i]
		}
	}
	if token == nil || token.OK || !token.Warning || token.Remedy == "" {
		t.Fatalf("a missing SAS token is a warning with a remedy: %+v", findings)
	}
	env["AZURE_STORAGE_SAS_TOKEN"] = "sv=2024&sig=x"
	for _, f := range cacheKeyFinding("azblob://acct/cache/p", func(k string) string { return env[k] }) {
		if !f.OK {
			t.Fatalf("with the key and the token set nothing is wrong: %+v", f)
		}
	}
}

func TestOtherRemotesAndAMissingKeyAddNoSASFinding(t *testing.T) {
	none := func(string) string { return "" }
	for _, f := range cacheKeyFinding("azblob://acct/cache/p", none) {
		if f.Check == "the remote cache has its SAS token" {
			t.Fatal("without the key the key finding says it: no second warning")
		}
	}
	key := func(k string) string {
		if k == "SMITH_CACHE_KEY" {
			return "abc"
		}
		return ""
	}
	for _, f := range cacheKeyFinding("s3://bucket/p", key) {
		if f.Check == "the remote cache has its SAS token" {
			t.Fatal("only an Azure Blob remote needs a SAS token")
		}
	}
}
