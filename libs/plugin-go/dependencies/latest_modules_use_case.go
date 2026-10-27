package dependencies

import (
	"strings"
	"sync"
	"unicode"
)

// ProxyURL is the module proxy smith asks for newest releases (GOPROXY's
// default); tests replace FetchJSON instead.
const ProxyURL = "https://proxy.golang.org"

// escapeModulePath writes a module path the way the proxy protocol wants
// it: every upper-case letter becomes ! and its lower-case form.
func escapeModulePath(path string) string {
	var b strings.Builder
	for _, r := range path {
		if unicode.IsUpper(r) {
			b.WriteByte('!')
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}

// latestModules asks the module proxy for each module's newest release, a
// few at a time. Versions come back without the leading v, as declarations
// carry them.
func latestModules(names []string) (latest, unreachable map[string]string) {
	latest, unreachable = map[string]string{}, map[string]string{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, 8)
	for _, name := range names {
		wg.Add(1)
		slots <- struct{}{}
		go func(name string) {
			defer func() { <-slots; wg.Done() }()
			var info struct{ Version string }
			err := FetchJSON(ProxyURL+"/"+escapeModulePath(name)+"/@latest", &info)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				unreachable[name] = err.Error()
			case info.Version != "":
				latest[name] = strings.TrimPrefix(info.Version, "v")
			}
		}(name)
	}
	wg.Wait()
	return latest, unreachable
}
