package remotecache

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// The GET Object example from the S3 Signature Version 4 documentation.
func TestSignS3MatchesTheAWSExample(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://examplebucket.s3.amazonaws.com/test.txt", nil)
	req.Header.Set("Range", "bytes=0-9")
	SignS3(req, Credentials{AccessKey: "AKIAIOSFODNN7EXAMPLE", SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", Region: "us-east-1"},
		EmptyPayloadHash, time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC))
	got := req.Header.Get("Authorization")
	want := "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-content-sha256;x-amz-date, Signature=f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41"
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
	if !strings.HasPrefix(uriEncode("/a b/ü"), "/a%20b/%C3%BC") {
		t.Fatal(uriEncode("/a b/ü"))
	}
}
