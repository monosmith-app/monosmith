package session

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func remoteSession(t *testing.T) *Session {
	t.Helper()
	cfg := workspace.Config{}
	cfg.Cache.Remote = "https://cache.example.invalid/smith"
	cfg.Cache.Dir = t.TempDir()
	return &Session{Root: t.TempDir(), Config: cfg}
}

func clearCacheEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range append([]string{"SMITH_CACHE_REMOTE", "SMITH_CACHE_READ_ONLY", "GITHUB_EVENT_NAME", "BUILD_REASON", "CI_PIPELINE_SOURCE"}, cacheEnvironment...) {
		t.Setenv(name, "")
	}
}

func TestAnEmptyCacheKeyLeavesTheRunWorkingWithTheRemoteOff(t *testing.T) {
	// GitHub passes a fork's pull request an empty secret.
	clearCacheEnvironment(t)
	storage, err := remoteSession(t).Storage()
	if err != nil {
		t.Fatal(err)
	}
	if _, local := storage.(*taskcache.LocalStorage); !local {
		t.Errorf("storage is %T, want the local store alone", storage)
	}
}

func TestAnUnexpandedAzureMacroLeavesTheRemoteOff(t *testing.T) {
	// Azure Pipelines hands a fork's build the variable's own name.
	clearCacheEnvironment(t)
	t.Setenv("SMITH_CACHE_KEY", "$(SMITH_CACHE_KEY)")
	t.Setenv("SMITH_CACHE_TOKEN", "$(SMITH_CACHE_TOKEN)")
	storage, err := remoteSession(t).Storage()
	if err != nil {
		t.Fatal(err)
	}
	if _, local := storage.(*taskcache.LocalStorage); !local {
		t.Errorf("storage is %T, want the local store alone", storage)
	}
}

func TestPullRequestsReadTheRemoteAndNeverWriteIt(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"github": {"GITHUB_EVENT_NAME": "pull_request"},
		"azure":  {"BUILD_REASON": "PullRequest"},
		"gitlab": {"CI_PIPELINE_SOURCE": "merge_request_event"},
	} {
		clearCacheEnvironment(t)
		t.Setenv("SMITH_CACHE_KEY", "a-shared-secret")
		for key, value := range env {
			t.Setenv(key, value)
		}
		storage, err := remoteSession(t).Storage()
		if err != nil {
			t.Fatal(err)
		}
		tiered, ok := storage.(*remotecache.Tiered)
		if !ok || !tiered.ReadOnly {
			t.Errorf("%s: storage is %T (read only: %v), want a read-only remote", name, storage, ok && tiered.ReadOnly)
		}
	}
}

func TestThePushToTheReleaseBranchWritesTheRemote(t *testing.T) {
	clearCacheEnvironment(t)
	t.Setenv("SMITH_CACHE_KEY", "a-shared-secret")
	t.Setenv("GITHUB_EVENT_NAME", "push")
	storage, err := remoteSession(t).Storage()
	if err != nil {
		t.Fatal(err)
	}
	if tiered, ok := storage.(*remotecache.Tiered); !ok || tiered.ReadOnly {
		t.Errorf("storage is %T, want a writable remote", storage)
	}
}
