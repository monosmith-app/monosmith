// Command schemagen writes the JSON Schema of the plugin protocol,
// schema/protocol.schema.json, from the Go types in the protocol package.
// Run it from the SDK module's root (packages/plugin-sdk) after any change to
// the protocol:
//
//	go run ./schemagen
//
// It fails when the protocol has a method, an error code or a capability the
// schema's tables do not describe yet. With -check it writes nothing and fails
// when the committed file is not what it would write.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocolschema"
)

func main() {
	protocolDir := flag.String("protocol", "protocol", "the directory of the protocol package")
	out := flag.String("out", filepath.Join("schema", "protocol.schema.json"), "the schema file to write")
	check := flag.Bool("check", false, "write nothing; fail when the file is not up to date")
	flag.Parse()
	if err := run(*protocolDir, *out, *check); err != nil {
		fmt.Fprintln(os.Stderr, "schemagen:", err)
		os.Exit(1)
	}
}

func run(protocolDir, out string, check bool) error {
	schema, err := protocolschema.Build(protocolDir)
	if err != nil {
		return err
	}
	current, readErr := os.ReadFile(out)
	if readErr == nil && bytes.Equal(bytes.ReplaceAll(current, []byte("\r\n"), []byte("\n")), schema) {
		fmt.Println(out, "is up to date")
		return nil
	}
	if check {
		return fmt.Errorf("%s is not up to date: run `go run ./schemagen` in packages/plugin-sdk", out)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, schema, 0o644); err != nil {
		return err
	}
	fmt.Println("wrote", out)
	return nil
}
