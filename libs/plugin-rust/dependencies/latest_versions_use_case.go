package dependencies

import (
	"sync"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Lookup asks crates.io for a crate's newest stable version; tests replace it.
var Lookup = releasing.NewestStable

// Latest asks crates.io for each crate's newest stable release, a few at a
// time.
func Latest(_ string, params protocol.LatestParams) (protocol.LatestResult, error) {
	res := protocol.LatestResult{Latest: map[string]string{}, Unreachable: map[string]string{}}
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, 8)
	for _, name := range params.Names {
		wg.Add(1)
		slots <- struct{}{}
		go func(name string) {
			defer func() { <-slots; wg.Done() }()
			r := Lookup(name)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case r.Unreachable:
				res.Unreachable[name] = r.Detail
			case r.Version != "":
				res.Latest[name] = r.Version
			}
		}(name)
	}
	wg.Wait()
	return res, nil
}
