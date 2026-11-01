package apisurface

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// prefixes remembers, by workspace root, where the root is below the top of
// its repository, which does not change during a run.
var prefixes sync.Map

// batchReader reads many files of one git revision through a single
// `git cat-file --batch` process. A git process per file costs 50 to 100 ms
// on Windows, and an API check reads several files of every package twice.
type batchReader struct {
	root, ref, prefix string
	cmd               *exec.Cmd
	in                io.WriteCloser
	out               *bufio.Reader
	failed            bool
}

// newBatchReader prepares a reader for ref. The git process starts on the
// first Read.
func newBatchReader(root, ref string) *batchReader {
	return &batchReader{root: root, ref: ref}
}

func (b *batchReader) start() error {
	if cached, ok := prefixes.Load(b.root); ok {
		b.prefix = cached.(string)
	} else {
		prefix, _, err := toolrun.Capture(b.root, []string{"git", "rev-parse", "--show-prefix"})
		if err != nil {
			return err
		}
		b.prefix = strings.TrimSpace(prefix)
		prefixes.Store(b.root, b.prefix)
	}
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = b.root
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	b.cmd, b.in, b.out = cmd, in, bufio.NewReader(stdout)
	return nil
}

// Read implements Source. A path the revision does not hold, and any failure
// of the git process, answer false.
func (b *batchReader) Read(rel string) (string, bool) {
	if b.failed {
		return "", false
	}
	if b.cmd == nil {
		if err := b.start(); err != nil {
			b.failed = true
			return "", false
		}
	}
	if _, err := fmt.Fprintf(b.in, "%s:%s%s\n", b.ref, b.prefix, rel); err != nil {
		b.failed = true
		return "", false
	}
	header, err := b.out.ReadString('\n')
	if err != nil {
		b.failed = true
		return "", false
	}
	fields := strings.Fields(header)
	// "<oid> <type> <size>" for an object, "<name> missing" for none.
	if len(fields) != 3 || fields[1] != "blob" {
		return "", false
	}
	size, err := strconv.Atoi(fields[2])
	if err != nil {
		b.failed = true
		return "", false
	}
	body := make([]byte, size+1) // the content and the newline that ends it
	if _, err := io.ReadFull(b.out, body); err != nil {
		b.failed = true
		return "", false
	}
	return string(body[:size]), true
}

// Close ends the git process.
func (b *batchReader) Close() error {
	if b.cmd == nil {
		return nil
	}
	_ = b.in.Close()
	return b.cmd.Wait()
}
