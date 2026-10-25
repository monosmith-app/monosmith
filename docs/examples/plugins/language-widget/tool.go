package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// runTool is the command line the plugin's tasks call: the tasks run this
// same executable, so the example needs no other tool installed.
//
//	smith-plugin-widget check            validate ./widget.json
//	smith-plugin-widget build <out-dir>  write <out-dir>/widget.txt
func runTool(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: smith-plugin-widget check | build <out-dir> (smith runs this program as a plugin)")
	}
	m, err := readManifest(manifestFile)
	if err != nil {
		return err
	}
	switch args[0] {
	case "check":
		if err := m.validate(); err != nil {
			return err
		}
		fmt.Printf("%s %s is valid\n", m.Name, m.Version)
	case "build":
		if len(args) != 2 {
			return fmt.Errorf("build takes the output directory")
		}
		if err := os.MkdirAll(args[1], 0o755); err != nil {
			return err
		}
		out := filepath.Join(args[1], "widget.txt")
		if err := os.WriteFile(out, []byte(m.Name+" "+m.Version+"\n"), 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", filepath.ToSlash(out))
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}
