package protocolschema

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckTables compares the methods and error-code tables with the protocol's
// own declarations and says everything that is out of step: a Method or Code
// constant no table row names, a row naming a constant that is gone, a
// capability no method needs, a method that needs one that does not exist.
func CheckTables(docs *SourceDocs) error {
	var problems []string

	declared := map[string]string{} // method name on the wire -> constant
	for _, c := range docs.ConstsWithPrefix("Method") {
		declared[c.Value] = c.Name
	}
	tabled := map[string]bool{}
	for _, spec := range Methods {
		tabled[spec.Method] = true
		if _, ok := declared[spec.Method]; !ok {
			problems = append(problems, fmt.Sprintf("the methods table has %q, which the protocol package does not declare", spec.Method))
		}
	}
	for method, name := range declared {
		if !tabled[method] {
			problems = append(problems, fmt.Sprintf("%s (%q) is missing from Methods in method_table_config.go", name, method))
		}
	}

	fields := map[string]bool{}
	for f := range reflect.TypeFor[protocol.Capabilities]().Fields() {
		fields[f.Name] = true
	}
	used := map[string]bool{}
	for _, spec := range Methods {
		if spec.Capability == "" {
			continue
		}
		if !fields[spec.Capability] {
			problems = append(problems, fmt.Sprintf("method %s needs capability %q, which protocol.Capabilities does not have", spec.Method, spec.Capability))
		}
		used[spec.Capability] = true
	}
	for f := range fields {
		if !used[f] && !modifierCapabilities[f] {
			problems = append(problems, fmt.Sprintf("capability %s is needed by no method in Methods", f))
		}
	}

	known := map[string]bool{}
	for _, spec := range ErrorCodes {
		known[spec.Const] = true
		if _, ok := docs.Const(spec.Const); !ok {
			problems = append(problems, fmt.Sprintf("the error-code table has %s, which the protocol package does not declare", spec.Const))
		}
	}
	for _, c := range docs.ConstsWithPrefix("Code") {
		if !known[c.Name] {
			problems = append(problems, fmt.Sprintf("%s is missing from ErrorCodes in error_code_table_config.go", c.Name))
		}
	}

	if len(problems) == 0 {
		return nil
	}
	slices.Sort(problems)
	return fmt.Errorf("the protocol changed and the schema tables did not follow:\n  %s", strings.Join(problems, "\n  "))
}

// modifierCapabilities are the capabilities that no method needs because they
// change how a method behaves: Channels makes publish take a release channel.
var modifierCapabilities = map[string]bool{"Channels": true}
