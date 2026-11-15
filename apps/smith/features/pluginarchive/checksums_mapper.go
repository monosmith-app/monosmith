package pluginarchive

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// FormatChecksums writes digests (name to lower-case hex SHA-256) in the
// format of sha256sum, sorted by name: "<hex>  <name>\n".
func FormatChecksums(sums map[string]string) []byte {
	names := make([]string, 0, len(sums))
	for n := range sums {
		names = append(names, n)
	}
	sort.Strings(names)
	var b bytes.Buffer
	for _, n := range names {
		fmt.Fprintf(&b, "%s  %s\n", sums[n], n)
	}
	return b.Bytes()
}

// ParseChecksums reads what FormatChecksums wrote. A line that is not a
// 64-digit hex digest and a safe name, or a name listed twice, is an error.
func ParseChecksums(raw []byte) (map[string]string, error) {
	sums := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(raw))
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(text) == "" {
			continue
		}
		digest, name, ok := strings.Cut(text, "  ")
		if !ok || !isHexDigest(digest) {
			return nil, fmt.Errorf("%s line %d: expected \"<sha256>  <name>\"", ChecksumsFile, line)
		}
		if err := ValidateName(name); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", ChecksumsFile, line, err)
		}
		if _, dup := sums[name]; dup {
			return nil, fmt.Errorf("%s lists %s twice", ChecksumsFile, name)
		}
		sums[name] = strings.ToLower(digest)
	}
	return sums, sc.Err()
}

func isHexDigest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		isHex := r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
		if !isHex {
			return false
		}
	}
	return true
}
