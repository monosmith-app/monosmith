package remotecache

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
)

// cloudImitation is an object store in memory that answers as S3 (path-style)
// or Azure Blob does: objects by path, a paged listing with a continuation
// token or a marker, DELETE, and 403 for credentials it refuses.
type cloudImitation struct {
	mu      sync.Mutex
	objects map[string][]byte // full path, "/bucket/team/ac/k"
	pages   int               // listing requests answered
	deletes int
	refuse  bool // answer 403 to every request
}

const pageSize = 2

func newCloudImitation(t *testing.T, azure bool) (*cloudImitation, *httptest.Server) {
	c := &cloudImitation{objects: map[string][]byte{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		defer c.mu.Unlock()
		allowed := strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=AKID/")
		if azure {
			allowed = r.URL.Query().Get("sig") == "good" && r.Header.Get("x-ms-version") != ""
		}
		if c.refuse || !allowed {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		q := r.URL.Query()
		switch {
		case !azure && r.Method == http.MethodGet && q.Get("list-type") == "2":
			c.pages++
			bucket := strings.TrimSuffix(r.URL.Path, "/")
			keys, next := c.page(bucket+"/", q.Get("prefix"), q.Get("continuation-token"))
			fmt.Fprint(w, "<ListBucketResult>")
			for _, k := range keys {
				fmt.Fprintf(w, "<Contents><Key>%s</Key><Size>1</Size></Contents>", k)
			}
			fmt.Fprintf(w, "<IsTruncated>%v</IsTruncated><NextContinuationToken>%s</NextContinuationToken></ListBucketResult>", next != "", next)
		case azure && r.Method == http.MethodGet && q.Get("comp") == "list":
			c.pages++
			keys, next := c.page(r.URL.Path+"/", q.Get("prefix"), q.Get("marker"))
			fmt.Fprint(w, "<EnumerationResults><Blobs>")
			for _, k := range keys {
				fmt.Fprintf(w, "<Blob><Name>%s</Name></Blob>", k)
			}
			fmt.Fprintf(w, "</Blobs><NextMarker>%s</NextMarker></EnumerationResults>", next)
		case r.Method == http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			c.objects[r.URL.Path] = body
		case r.Method == http.MethodDelete:
			c.deletes++
			if _, ok := c.objects[r.URL.Path]; !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			delete(c.objects, r.URL.Path)
			w.WriteHeader(map[bool]int{true: http.StatusAccepted, false: http.StatusNoContent}[azure])
		default:
			body, ok := c.objects[r.URL.Path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Method == http.MethodGet {
				_, _ = w.Write(body)
			}
		}
	}))
	t.Cleanup(srv.Close)
	return c, srv
}

// page lists the keys under root (the bucket or container path) that start
// with prefix, pageSize at a time; the token is the index to start at.
func (c *cloudImitation) page(root, prefix, token string) ([]string, string) {
	var all []string
	for p := range c.objects {
		if key, ok := strings.CutPrefix(p, root); ok && strings.HasPrefix(key, prefix) {
			all = append(all, key)
		}
	}
	sort.Strings(all)
	start := 0
	if token != "" {
		_, _ = fmt.Sscanf(token, "t%d", &start)
	}
	end := min(start+pageSize, len(all))
	next := ""
	if end < len(all) {
		next = fmt.Sprintf("t%d", end)
	}
	return all[start:end], next
}

func openS3(t *testing.T, srv *httptest.Server) ObjectStore {
	t.Setenv("AWS_ACCESS_KEY_ID", "AKID")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_REGION", "eu-west-1")
	t.Setenv("AWS_ENDPOINT_URL_S3", srv.URL)
	store, err := Open("s3://bucket/team/cache")
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func openAzure(t *testing.T, srv *httptest.Server, sig string) ObjectStore {
	old := azureEndpoint
	azureEndpoint = func(string) string { return srv.URL }
	t.Cleanup(func() { azureEndpoint = old })
	t.Setenv("AZURE_STORAGE_SAS_TOKEN", "?sv=2021-08-06&sp=rwdlc&sig="+sig)
	store, err := Open("azblob://acct/cache/team")
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func fill(t *testing.T, store ObjectStore, names ...string) {
	t.Helper()
	for _, n := range names {
		if err := store.Put(n, []byte("x")); err != nil {
			t.Fatal(err)
		}
	}
}

func listAll(t *testing.T, store ObjectStore, prefix string) []string {
	t.Helper()
	var got []string
	if err := store.List(prefix, func(n string) error { got = append(got, n); return nil }); err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	return got
}

func TestListReadsEveryPageAndDeleteRemoves(t *testing.T) {
	for _, azure := range []bool{false, true} {
		name := map[bool]string{false: "s3", true: "azblob"}[azure]
		t.Run(name, func(t *testing.T) {
			cloud, srv := newCloudImitation(t, azure)
			var store ObjectStore
			if azure {
				store = openAzure(t, srv, "good")
			} else {
				store = openS3(t, srv)
			}
			fill(t, store, "ac/a1", "ac/a2", "ac/a3", "ac/a4", "ac/a5", "cas/b1")
			got := listAll(t, store, "ac/")
			if strings.Join(got, ",") != "ac/a1,ac/a2,ac/a3,ac/a4,ac/a5" {
				t.Fatalf("listed %v", got)
			}
			if cloud.pages != 3 {
				t.Errorf("read %d pages, want 3 of %d", cloud.pages, pageSize)
			}
			for _, n := range got {
				if err := store.Delete(n); err != nil {
					t.Fatal(err)
				}
			}
			if err := store.Delete("ac/a1"); err != nil {
				t.Errorf("deleting what is gone must succeed: %v", err)
			}
			if has, _ := store.Has("ac/a3"); has {
				t.Error("ac/a3 is still there")
			}
			if left := listAll(t, store, ""); strings.Join(left, ",") != "cas/b1" {
				t.Errorf("left %v", left)
			}
		})
	}
}

func TestARefusedListOrDeleteIsARefusal(t *testing.T) {
	cloud, srv := newCloudImitation(t, false)
	store := openS3(t, srv)
	fill(t, store, "ac/a1")
	cloud.refuse = true
	if err := store.List("ac/", func(string) error { return nil }); !Refused(err) {
		t.Errorf("list: %v", err)
	}
	if err := store.Delete("ac/a1"); !Refused(err) {
		t.Errorf("delete: %v", err)
	}
	_, srv2 := newCloudImitation(t, true)
	if err := openAzure(t, srv2, "bad").List("ac/", func(string) error { return nil }); !Refused(err) {
		t.Errorf("azure list with a bad token: %v", err)
	}
}

func TestAPlainStoreCannotListButDeletes(t *testing.T) {
	srv, objects := objectServer(t)
	t.Setenv("SMITH_CACHE_TOKEN", "team-token")
	store, err := Open(srv.URL + "/team")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.List("ac/", func(string) error { return nil }); err != ErrCannotList {
		t.Fatalf("list: %v", err)
	}
	fill(t, store, "ac/a1")
	if err := store.Delete("ac/a1"); err != nil {
		t.Fatal(err)
	}
	if len(objects) != 0 {
		t.Errorf("left %v", objects)
	}
}
