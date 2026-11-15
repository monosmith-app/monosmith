package releasesign

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"aead.dev/minisign"
)

// KeyEnv holds the minisign private key a release signs with: the text of an
// unencrypted key file. Unset, a release signs nothing.
const KeyEnv = "SMITH_SIGNING_KEY"

// Suffix is added to a signed file's name to name its signature.
const Suffix = ".minisig"

// Sign writes a minisign signature beside every checksums.txt among a
// release's assets, so a launcher trusts an archive only when its checksum is
// in a list the publisher signed (ADR 0101). It returns the assets with the
// signatures added. With no key, or no checksums.txt, it changes nothing.
//
// Parameters: root is the workspace root; assets are the workspace-relative
// files the release attaches; key is the private key text (KeyEnv); tag names
// the release in the signature's trusted comment.
// Returns the assets to attach, or an error for a key that cannot be read or
// a file that cannot be signed.
func Sign(root string, assets []string, key, tag string) ([]string, error) {
	if strings.TrimSpace(key) == "" {
		return assets, nil
	}
	var private minisign.PrivateKey
	if err := private.UnmarshalText([]byte(strings.TrimSpace(key))); err != nil {
		return nil, fmt.Errorf("%s is not an unencrypted minisign private key: %w", KeyEnv, err)
	}
	out := slices.Clone(assets)
	for _, asset := range assets {
		if path.Base(asset) != "checksums.txt" || slices.Contains(assets, asset+Suffix) {
			continue
		}
		file := filepath.Join(root, filepath.FromSlash(asset))
		message, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		signature := minisign.SignWithComments(private, message, "smith release "+tag, "signed by smith release")
		if err := os.WriteFile(file+Suffix, signature, 0o644); err != nil {
			return nil, err
		}
		out = append(out, asset+Suffix)
	}
	return out, nil
}
