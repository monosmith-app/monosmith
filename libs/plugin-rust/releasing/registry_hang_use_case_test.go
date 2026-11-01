package releasing

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// A registry that accepts the connection and never answers ends the lookup at
// the timeout with the skipped reason, and the plugin code keeps running.
func TestAHungRegistryEndsTheLookupWithTheSkippedReason(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	var held []net.Conn
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			held = append(held, c) // accepted and never answered
		}
	}()
	defer func() {
		for _, c := range held {
			_ = c.Close()
		}
	}()
	realClient, realAPI := registryClient, CratesAPI
	defer func() { registryClient, CratesAPI = realClient, realAPI }()
	registryClient, CratesAPI = &http.Client{Timeout: 300 * time.Millisecond}, "http://"+l.Addr().String()

	start := time.Now()
	res := NewestStable("acme-core")
	if took := time.Since(start); took > 3*time.Second {
		t.Fatalf("the lookup took %s", took)
	}
	if !res.Unreachable || res.Detail != protocol.RegistryUnreachableDetail {
		t.Fatalf("want the skipped reason, got %+v", res)
	}
	if again := NewestStable("acme-core"); !again.Unreachable {
		t.Fatalf("a second lookup must also end: %+v", again)
	}
}
