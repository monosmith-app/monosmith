package remotesync

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
)

// listingStore is an object store in memory that can list, and that refuses
// deletes when refuse is set.
type listingStore struct {
	mu      sync.Mutex
	objects map[string]bool
	refuse  bool
	fail    string
}

func (s *listingStore) Get(string) ([]byte, bool, error) { return nil, false, nil }
func (s *listingStore) Put(name string, _ []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[name] = true
	return nil
}
func (s *listingStore) Has(name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.objects[name], nil
}
func (s *listingStore) Describe() string { return "memory" }
func (s *listingStore) List(prefix string, visit func(string) error) error {
	s.mu.Lock()
	var names []string
	for n := range s.objects {
		if strings.HasPrefix(n, prefix) {
			names = append(names, n)
		}
	}
	s.mu.Unlock()
	sort.Strings(names)
	for _, n := range names {
		if err := visit(n); err != nil {
			return err
		}
	}
	return nil
}
func (s *listingStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.refuse {
		return &remotecache.StatusError{Method: "DELETE", Name: name, Code: 403, Status: "403 Forbidden"}
	}
	if name == s.fail {
		return errors.New("500 Internal Server Error")
	}
	delete(s.objects, name)
	return nil
}

func newListingStore(names ...string) *listingStore {
	s := &listingStore{objects: map[string]bool{}}
	for _, n := range names {
		s.objects[n] = true
	}
	return s
}

func TestFindObjectsListsEntriesAndBlobsOnly(t *testing.T) {
	store := newListingStore("ac/1", "cas/2", "cas/3", "other/x")
	got, err := FindObjects(Remote{Store: store}, nil)
	if err != nil || !got.Listed || strings.Join(got.Names, ",") != "ac/1,cas/2,cas/3" {
		t.Fatalf("found %+v, %v", got, err)
	}
}

func TestClearRemoteDeletesEverythingFound(t *testing.T) {
	store := newListingStore("ac/1", "cas/2", "cas/3")
	remote := Remote{Store: store, URL: "s3://bucket/team?secret=1"}
	objects, _ := FindObjects(remote, nil)
	var last int
	result, err := ClearRemote(context.Background(), remote, objects, func(done int) { last = done })
	if err != nil || result.Deleted != 3 || result.Found != 3 || last != 3 || len(store.objects) != 0 || result.Store != "s3://bucket/team" {
		t.Fatalf("result %+v, %v, left %v", result, err, store.objects)
	}
}

func TestClearRemoteCountsAFailureAndGoesOn(t *testing.T) {
	store := newListingStore("ac/1", "cas/2", "cas/3")
	store.fail = "cas/2"
	objects, _ := FindObjects(Remote{Store: store}, nil)
	result, err := ClearRemote(context.Background(), Remote{Store: store}, objects, nil)
	if err != nil || result.Deleted != 2 || result.Failed != 1 || len(result.Errors) != 1 {
		t.Fatalf("result %+v, %v", result, err)
	}
}

func TestClearRemoteStopsAtARefusalAndRefusesAReadOnlyRemote(t *testing.T) {
	store := newListingStore("ac/1", "cas/2")
	store.refuse = true
	objects, _ := FindObjects(Remote{Store: store}, nil)
	if _, err := ClearRemote(context.Background(), Remote{Store: store}, objects, nil); !remotecache.Refused(err) {
		t.Fatalf("err = %v", err)
	}
	if _, err := ClearRemote(context.Background(), Remote{Store: store, ReadOnly: true}, objects, nil); err != ErrReadOnly {
		t.Fatalf("read-only: %v", err)
	}
}

func TestFindObjectsOnAStoreThatCannotListNamesTheLocalEntries(t *testing.T) {
	_, srv := newFakeStore(t)
	remote := connect(t, plan(t, srv))
	got, err := FindObjects(remote, localWith(t, 2))
	if err != nil || got.Listed || len(got.Names) != 6 {
		t.Fatalf("found %+v, %v", got, err)
	}
}
