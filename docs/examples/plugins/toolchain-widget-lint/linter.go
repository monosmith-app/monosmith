package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// defaultRequire is what every widget must say about itself when the
// project's settings do not choose.
var defaultRequire = []string{"description", "license"}

// runLint is the lint task's command line:
//
//	smith-plugin-widget-lint lint <field,field,...>
//
// It reads ./widget.json and fails, naming each field that is empty.
func runLint(args []string) error {
	if len(args) != 2 || args[0] != "lint" {
		return fmt.Errorf("usage: smith-plugin-widget-lint lint <field,field,...> (smith runs this program as a plugin)")
	}
	raw, err := os.ReadFile("widget.json")
	if err != nil {
		return err
	}
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return fmt.Errorf("widget.json: %w", err)
	}
	var missing []string
	for _, field := range strings.Split(args[1], ",") {
		if value, _ := manifest[field].(string); strings.TrimSpace(value) == "" {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("widget.json must set %s", strings.Join(missing, ", "))
	}
	fmt.Println("widget.json sets", args[1])
	return nil
}
