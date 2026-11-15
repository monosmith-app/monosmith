package remotecache

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Credentials sign S3 requests (AWS Signature Version 4).
type Credentials struct {
	AccessKey, SecretKey, SessionToken, Region string
}

// EmptyPayloadHash is the SHA-256 of no content.
const EmptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// uriEncode escapes a path the way SigV4 canonicalises it: every byte but
// unreserved characters and "/".
func uriEncode(path string) string {
	var b strings.Builder
	for _, c := range []byte(path) {
		switch {
		case 'A' <= c && c <= 'Z', 'a' <= c && c <= 'z', '0' <= c && c <= '9', c == '-', c == '_', c == '.', c == '~', c == '/':
			b.WriteByte(c)
		default:
			b.WriteString("%" + strings.ToUpper(hex.EncodeToString([]byte{c})))
		}
	}
	return b.String()
}

// SignS3 adds the SigV4 headers to a request for the s3 service, its payload
// hash given (EmptyPayloadHash for GET and HEAD).
func SignS3(req *http.Request, c Credentials, payloadHash string, now time.Time) {
	stamp := now.UTC().Format("20060102T150405Z")
	day := stamp[:8]
	req.Header.Set("x-amz-date", stamp)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	if c.SessionToken != "" {
		req.Header.Set("x-amz-security-token", c.SessionToken)
	}
	headers := map[string]string{"host": req.URL.Host}
	for name, values := range req.Header {
		headers[strings.ToLower(name)] = strings.TrimSpace(strings.Join(values, ","))
	}
	var names []string
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	var canonicalHeaders strings.Builder
	for _, n := range names {
		canonicalHeaders.WriteString(n + ":" + headers[n] + "\n")
	}
	signed := strings.Join(names, ";")
	query := req.URL.Query()
	var keys []string
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var pairs []string
	for _, k := range keys {
		for _, v := range query[k] {
			pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	path := req.URL.EscapedPath()
	if unescaped, err := url.PathUnescape(path); err == nil {
		path = uriEncode(unescaped)
	}
	if path == "" {
		path = "/"
	}
	canonical := strings.Join([]string{req.Method, path, strings.Join(pairs, "&"), canonicalHeaders.String(), signed, payloadHash}, "\n")
	scope := day + "/" + c.Region + "/s3/aws4_request"
	toSign := "AWS4-HMAC-SHA256\n" + stamp + "\n" + scope + "\n" + sha256Hex([]byte(canonical))
	key := hmacSHA256([]byte("AWS4"+c.SecretKey), day)
	key = hmacSHA256(key, c.Region)
	key = hmacSHA256(key, "s3")
	key = hmacSHA256(key, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(key, toSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+c.AccessKey+"/"+scope+", SignedHeaders="+signed+", Signature="+signature)
}
