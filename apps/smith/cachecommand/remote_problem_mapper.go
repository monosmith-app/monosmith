package cachecommand

import (
	"errors"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
)

// remoteProblem turns an error of the remote cache into the problem and the
// step that clears it.
func remoteProblem(err error, remote string) error {
	store := remotesync.StoreName(remote)
	switch {
	case errors.Is(err, remotesync.ErrNoRemote):
		return problem.New("no remote cache is set for this workspace", "run `smith cache setup` to set one up")
	case errors.Is(err, remotesync.ErrNoKey):
		return problem.New("the remote cache "+store+" is off: SMITH_CACHE_KEY is not set", "set SMITH_CACHE_KEY to the team's key (`smith cache key` makes a new one)")
	case errors.Is(err, remotesync.ErrReadOnly):
		return problem.New("the remote cache "+store+" is read-only here", "unset SMITH_CACHE_READ_ONLY, or set remote_read_only = false under [cache] in smith.toml", "a pull-request build never writes the remote: push from the main branch")
	case remotecache.Refused(err):
		var se *remotecache.StatusError
		errors.As(err, &se)
		return problem.New("the remote cache "+store+" refused the credentials (it answered "+se.Status+")", credentialsRemedy(remote))
	}
	return problem.New("could not reach the remote cache "+store+": "+err.Error(), "check the network and run the command again", "`smith cache status` shows the remote settings")
}

// credentialsRemedy names the credentials a store reads.
func credentialsRemedy(remote string) string {
	switch {
	case strings.HasPrefix(remote, "s3://"):
		return "check AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_REGION: the key needs read, write and list rights on the bucket (without list, S3 answers 403 for a missing object)"
	case strings.HasPrefix(remote, "azblob://"):
		return "check AZURE_STORAGE_SAS_TOKEN: it must not be expired and needs read, write, delete, create and list rights (`smith cache setup` makes one)"
	}
	return "check SMITH_CACHE_TOKEN: the server must accept it"
}

// deleteRightsRemedy says how to get a credential that may list and delete.
func deleteRightsRemedy(remote string) string {
	switch {
	case strings.HasPrefix(remote, "s3://"):
		return "give the key of AWS_ACCESS_KEY_ID s3:ListBucket and s3:DeleteObject on the bucket"
	case strings.HasPrefix(remote, "azblob://"):
		return "issue a new AZURE_STORAGE_SAS_TOKEN with list and delete rights (`smith cache setup` makes one with read, write, delete, create and list)"
	}
	return "the server must accept DELETE with SMITH_CACHE_TOKEN"
}
