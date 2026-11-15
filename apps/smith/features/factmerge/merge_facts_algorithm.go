package factmerge

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Merge merges layers into one fact per key. For each key the narrowest scope
// that sets it wins, whatever the order of layers, and wider scopes are not
// looked at. Layers of the winning scope that set the key to different values
// are a conflict, and Merge returns an error that names the key, both sources
// and both values; equal values agree. The values of layers are not changed.
func Merge(layers []Layer) (map[string]Fact, error) {
	merged := map[string]Fact{}
	// Narrowest first: the first layer to set a key fixes its scope.
	ordered := slices.SortedStableFunc(slices.Values(layers), func(a, b Layer) int { return int(b.Scope) - int(a.Scope) })
	for _, l := range ordered {
		for key, value := range l.Facts {
			prior, set := merged[key]
			switch {
			case !set:
				merged[key] = Fact{Value: value, Scope: l.Scope, Source: l.Source}
			case l.Scope < prior.Scope:
				// A narrower scope already won.
			case sameValue(prior.Value, value):
				if !strings.Contains(prior.Source, l.Source) {
					prior.Source += " and " + l.Source
					merged[key] = prior
				}
			default:
				return nil, fmt.Errorf("fact %q is set to two values at the same scope (%s): %s says %s, %s says %s; set it in smith.toml [facts] to decide",
					key, l.Scope, prior.Source, show(prior.Value), l.Source, show(value))
			}
		}
	}
	return merged, nil
}

// Values returns only the values of merged facts, the shape a request carries.
// It returns nil for no facts.
func Values(merged map[string]Fact) map[string]any {
	if len(merged) == 0 {
		return nil
	}
	out := make(map[string]any, len(merged))
	for k, f := range merged {
		out[k] = f.Value
	}
	return out
}

func sameValue(a, b any) bool {
	ja, errA := json.Marshal(a)
	jb, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return reflect.DeepEqual(a, b)
	}
	return string(ja) == string(jb)
}

func show(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}
