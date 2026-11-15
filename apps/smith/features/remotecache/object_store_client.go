package remotecache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ObjectStore reads and writes whole objects by name.
type ObjectStore interface {
	// Get returns an object; found is false when it does not exist.
	Get(name string) (data []byte, found bool, err error)
	// Put writes an object, replacing one of the same name.
	Put(name string, data []byte) error
	// Has reports whether an object exists, without reading it.
	Has(name string) (bool, error)
	// Describe names the store for messages.
	Describe() string
	// List calls visit with the name of every object whose name starts with
	// prefix ("ac/", "cas/"), names as Get takes them. It reads the listing
	// page by page. A store that cannot list returns ErrCannotList.
	List(prefix string, visit func(name string) error) error
	// Delete removes an object; an object that does not exist is a success.
	Delete(name string) error
}

// ErrCannotList is returned by List on a plain https:// store, which takes
// PUT, GET, HEAD and DELETE but has no listing.
var ErrCannotList = errors.New("a plain https:// store cannot list its objects")

// httpStore is an object store reached over HTTP: a plain server (PUT, GET,
// HEAD, DELETE under a base URL), S3 or Azure Blob, which differ in the URL of
// an object, how a request is authorised and how a listing is read.
type httpStore struct {
	describe  string
	objectURL func(name string) string
	authorise func(req *http.Request, payloadHash string)
	// list reads the listing of names under a prefix; nil when the store has none.
	list   func(s httpStore, prefix string, visit func(name string) error) error
	client *http.Client
}

func (s httpStore) Describe() string { return s.describe }

func (s httpStore) do(method, name string, body []byte) (*http.Response, error) {
	return s.doURL(method, s.objectURL(name), body)
}

func (s httpStore) doURL(method, target string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	hash := EmptyPayloadHash
	if body != nil {
		hash = sha256Hex(body)
		req.ContentLength = int64(len(body))
	}
	s.authorise(req, hash)
	resp, err := s.client.Do(req)
	return resp, withoutQuery(err)
}

// withoutQuery drops the query of the URL a transport error names: on Azure
// it holds the SAS token, which must never reach a message.
func withoutQuery(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) {
		if i := strings.IndexByte(ue.URL, '?'); i >= 0 {
			ue.URL = ue.URL[:i]
		}
	}
	return err
}

func (s httpStore) Get(name string) ([]byte, bool, error) {
	resp, err := s.do(http.MethodGet, name, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = resp.Body.Close() }()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, false, nil
	case resp.StatusCode != http.StatusOK:
		return nil, false, statusError(http.MethodGet, name, resp)
	}
	data, err := io.ReadAll(resp.Body)
	return data, err == nil, err
}

func (s httpStore) Has(name string) (bool, error) {
	resp, err := s.do(http.MethodHead, name, nil)
	if err != nil {
		return false, err
	}
	_ = resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}
	return false, statusError(http.MethodHead, name, resp)
}

func (s httpStore) Put(name string, data []byte) error {
	resp, err := s.do(http.MethodPut, name, data)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return statusError(http.MethodPut, name, resp)
	}
	return nil
}

func (s httpStore) Delete(name string) error {
	resp, err := s.do(http.MethodDelete, name, nil)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || (resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return nil
	}
	return statusError(http.MethodDelete, name, resp)
}

func (s httpStore) List(prefix string, visit func(name string) error) error {
	if s.list == nil {
		return ErrCannotList
	}
	return s.list(s, prefix, visit)
}

// getPage reads one page of a listing; what names the listing in an error.
func (s httpStore) getPage(target, what string) ([]byte, error) {
	resp, err := s.doURL(http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, statusError("LIST", what, resp)
	}
	return io.ReadAll(resp.Body)
}

// azureEndpoint is the Blob service of an account; tests point it at a local server.
var azureEndpoint = func(account string) string { return "https://" + account + ".blob.core.windows.net" }

// Open returns the store a remote URL names:
//
//	https://host/path   a server taking PUT, GET and HEAD under the path;
//	                    SMITH_CACHE_TOKEN, when set, is sent as a bearer token
//	s3://bucket/prefix  S3 (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY,
//	                    AWS_SESSION_TOKEN, AWS_REGION; AWS_ENDPOINT_URL_S3 for
//	                    an S3-compatible store, addressed path-style)
//	azblob://account/container/prefix
//	                    Azure Blob, authorised by AZURE_STORAGE_SAS_TOKEN
func Open(remote string) (ObjectStore, error) {
	u, err := url.Parse(remote)
	if err != nil {
		return nil, fmt.Errorf("[cache] remote %q: %w", remote, err)
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	prefix := strings.Trim(u.Path, "/")
	join := func(base, name string) string {
		if prefix != "" {
			name = prefix + "/" + name
		}
		return strings.TrimSuffix(base, "/") + "/" + name
	}
	switch u.Scheme {
	case "http", "https":
		token := os.Getenv("SMITH_CACHE_TOKEN")
		base := u.Scheme + "://" + u.Host
		return httpStore{describe: remote, client: client, objectURL: func(n string) string { return join(base, n) }, authorise: func(req *http.Request, _ string) {
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		}}, nil
	case "s3":
		creds := Credentials{AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"), SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"), SessionToken: os.Getenv("AWS_SESSION_TOKEN"), Region: os.Getenv("AWS_REGION")}
		if creds.Region == "" {
			creds.Region = "us-east-1"
		}
		if creds.AccessKey == "" || creds.SecretKey == "" {
			return nil, fmt.Errorf("[cache] remote %s needs AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY", remote)
		}
		base := "https://" + u.Host + ".s3." + creds.Region + ".amazonaws.com"
		if endpoint := os.Getenv("AWS_ENDPOINT_URL_S3"); endpoint != "" {
			base = strings.TrimSuffix(endpoint, "/") + "/" + u.Host
		}
		return httpStore{describe: remote, client: client, objectURL: func(n string) string { return join(base, n) }, authorise: func(req *http.Request, hash string) {
			SignS3(req, creds, hash, time.Now())
		}, list: func(s httpStore, p string, visit func(string) error) error {
			return listS3(s, base, prefix, p, visit)
		}}, nil
	case "azblob":
		sas := strings.TrimPrefix(os.Getenv("AZURE_STORAGE_SAS_TOKEN"), "?")
		if sas == "" {
			return nil, fmt.Errorf("[cache] remote %s needs AZURE_STORAGE_SAS_TOKEN", remote)
		}
		base := azureEndpoint(u.Host)
		return httpStore{describe: remote, client: client, objectURL: func(n string) string { return join(base, n) + "?" + sas }, authorise: func(req *http.Request, _ string) {
			req.Header.Set("x-ms-version", "2021-08-06")
			if req.Method == http.MethodPut {
				req.Header.Set("x-ms-blob-type", "BlockBlob")
			}
		}, list: func(s httpStore, p string, visit func(string) error) error {
			return listAzure(s, base, prefix, sas, p, visit)
		}}, nil
	}
	return nil, fmt.Errorf("[cache] remote %q: use https://, s3:// or azblob://", remote)
}
