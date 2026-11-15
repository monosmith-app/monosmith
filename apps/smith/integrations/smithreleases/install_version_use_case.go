package smithreleases

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"aead.dev/minisign"
)

// archiveSystems and archiveProcessors are Go's names as the archives spell them.
var (
	archiveSystems    = map[string]string{"darwin": "darwin", "linux": "linux", "windows": "windows"}
	archiveProcessors = map[string]string{"amd64": "amd64", "arm64": "arm64"}
)

// ArchiveName is the release archive of one version for a system:
// monosmith_<version>_<os>_<arch>.zip.
//
// Parameters: version is exact; goos and goarch are Go's names.
// Returns: the name, or an error for a system smith is not built for.
func ArchiveName(version, goos, goarch string) (string, error) {
	system, okSystem := archiveSystems[goos]
	processor, okProcessor := archiveProcessors[goarch]
	if !okSystem || !okProcessor {
		return "", fmt.Errorf("smith is not built for %s on %s", goos, goarch)
	}
	return "monosmith_" + version + "_" + system + "_" + processor + ".zip", nil
}

// ProgramPath is where one version's program lives under the smith home.
//
// Parameters: home is the smith home; version is exact; goos is Go's name.
// Returns: the path, whether it exists or not.
func ProgramPath(home, version, goos string) string {
	name := "smith"
	if goos == "windows" {
		name = "smith.exe"
	}
	return filepath.Join(home, "versions", version, name)
}

// Check says why an archive is not trusted, or "" when its checksum is in a
// checksums.txt that publicKey signed.
//
// Parameters: archive and name are the archive's bytes and file name;
// checksums and signature are checksums.txt and checksums.txt.minisig;
// publicKey is the key in minisign's base64 text.
// Returns: the reason, or "".
func Check(archive []byte, name string, checksums, signature []byte, publicKey string) string {
	var key minisign.PublicKey
	if err := key.UnmarshalText([]byte(publicKey)); err != nil {
		return "the release key cannot be read: " + err.Error()
	}
	if !minisign.Verify(key, checksums, signature) {
		return "the signature of checksums.txt does not match smith's release key"
	}
	listed := ""
	for _, line := range strings.Split(string(checksums), "\n") {
		if fields := strings.Fields(line); len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			listed = strings.ToLower(fields[0])
		}
	}
	if listed == "" {
		return "checksums.txt does not list " + name
	}
	sum := sha256.Sum256(archive)
	if actual := hex.EncodeToString(sum[:]); actual != listed {
		return fmt.Sprintf("%s does not match its checksum (got %s, listed %s)", name, actual, listed)
	}
	return ""
}

// Install downloads one version, checks it against the signed checksums.txt,
// and puts its program in <home>/versions/<version>. Nothing is written
// unless every check passes, and the folder appears in one rename.
//
// Parameters: home is the smith home; version is exact; goos and goarch are
// Go's names; publicKey is the release key (PublicKey).
// Returns: the program's path, or an error when a file cannot be downloaded
// or does not pass the checks.
func Install(ctx context.Context, home, version, goos, goarch, publicKey string) (string, error) {
	name, err := ArchiveName(version, goos, goarch)
	if err != nil {
		return "", err
	}
	archive, err := File(ctx, version, name)
	if err != nil {
		return "", err
	}
	checksums, err := File(ctx, version, "checksums.txt")
	if err != nil {
		return "", err
	}
	signature, err := File(ctx, version, "checksums.txt.minisig")
	if err != nil {
		return "", err
	}
	if why := Check(archive, name, checksums, signature, publicKey); why != "" {
		return "", fmt.Errorf("smith %s was not installed: %s", version, why)
	}
	target := ProgramPath(home, version, goos)
	program, err := unzipProgram(archive, filepath.Base(target))
	if err != nil {
		return "", fmt.Errorf("%s: %w", name, err)
	}
	staging := filepath.Join(home, "versions", fmt.Sprintf(".%s-%d", version, os.Getpid()))
	_ = os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(staging, filepath.Base(target)), program, 0o755); err != nil {
		return "", err
	}
	_ = os.RemoveAll(filepath.Dir(target))
	if err := os.Rename(staging, filepath.Dir(target)); err != nil {
		return "", err
	}
	return target, nil
}

// unzipProgram is the program's bytes from a release archive.
func unzipProgram(archive []byte, program string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		if path.Base(f.Name) != program {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = rc.Close() }()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("holds no %s", program)
}
