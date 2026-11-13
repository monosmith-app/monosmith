package dependencies

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestLatest(t *testing.T) {
	realLookup, realFetch := Lookup, FetchJSON
	defer func() { Lookup, FetchJSON = realLookup, realFetch }()
	Lookup = func(_, name string) protocol.PublishedResult {
		switch name {
		case "zod":
			return protocol.PublishedResult{Version: "3.23.8"}
		case "offline":
			return protocol.PublishedResult{Unreachable: true, Detail: "ETIMEDOUT"}
		}
		return protocol.PublishedResult{}
	}
	FetchJSON = func(_ string, v any) error {
		return json.Unmarshal([]byte(`[{"version":"v27.0.0","lts":false},{"version":"v26.1.0","lts":"Krypton"},{"version":"v24.9.0","lts":"Jod"}]`), v)
	}
	res, err := Latest("/ws", protocol.LatestParams{Names: []string{"zod", "offline", "missing"}, Tools: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Latest["zod"] != "3.23.8" || res.Unreachable["offline"] != "ETIMEDOUT" || len(res.Latest) != 1 {
		t.Fatalf("%+v", res)
	}
	if len(res.Tools) != 1 || res.Tools[0].Latest != "26.1.0" || res.Tools[0].Pinned != "24" || !res.Tools[0].Newer {
		t.Fatalf("tools %+v", res.Tools)
	}
	FetchJSON = func(string, any) error { return errors.New("offline") }
	res, _ = Latest("/ws", protocol.LatestParams{Tools: true})
	if res.Tools[0].Latest != "" || res.Tools[0].Newer {
		t.Fatalf("an unreachable feed reports no latest: %+v", res.Tools)
	}
}
