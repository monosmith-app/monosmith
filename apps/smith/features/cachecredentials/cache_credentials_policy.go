package cachecredentials

import (
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Key is the secret the remote cache's objects are signed with.
const Key = "SMITH_CACHE_KEY"

// For returns the credentials a remote cache needs: the key every object is
// signed with, and the backend's own, chosen by the remote's scheme. None is
// required: without the key the remote is off and the run uses the local
// cache, as it does in a fork's pull request, where secrets are withheld.
// Each one is passed to every step that runs tasks.
func For(remote string) []protocol.Credential {
	if remote == "" {
		return nil
	}
	creds := []protocol.Credential{{
		Env: Key, Tasks: true,
		Description: "signs and verifies the remote cache's objects (a random value the team shares: `smith cache key` makes one); without it the remote cache is off",
	}}
	switch {
	case strings.HasPrefix(remote, "azblob://"):
		creds = append(creds, protocol.Credential{
			Env: "AZURE_STORAGE_SAS_TOKEN", Tasks: true,
			Description: "a SAS token for the cache container, with read, write and create permission",
		})
	case strings.HasPrefix(remote, "s3://"):
		creds = append(creds,
			protocol.Credential{Env: "AWS_ACCESS_KEY_ID", Tasks: true, Description: "access key of an identity that can read and write the cache bucket"},
			protocol.Credential{Env: "AWS_SECRET_ACCESS_KEY", Tasks: true, Description: "secret key of that identity"},
			protocol.Credential{Env: "AWS_REGION", Tasks: true, Description: "the bucket's region (for Cloudflare R2: auto)"},
			protocol.Credential{Env: "AWS_ENDPOINT_URL_S3", Tasks: true, Description: "the store's S3 endpoint, for an S3-compatible store such as Cloudflare R2 (https://<account>.r2.cloudflarestorage.com); leave unset for Amazon S3"},
		)
	case strings.HasPrefix(remote, "https://"), strings.HasPrefix(remote, "http://"):
		creds = append(creds, protocol.Credential{
			Env: "SMITH_CACHE_TOKEN", Tasks: true,
			Description: "bearer token the cache server accepts (leave unset for a server that needs none)",
		})
	}
	return creds
}
