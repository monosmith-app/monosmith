package remotecache

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

// s3ListPage is one page of ListObjectsV2.
type s3ListPage struct {
	Contents []struct {
		Key string `xml:"Key"`
	} `xml:"Contents"`
	IsTruncated           bool   `xml:"IsTruncated"`
	NextContinuationToken string `xml:"NextContinuationToken"`
}

// listS3 reads ListObjectsV2 of the bucket at base, page by page, for the
// objects under storePrefix + "/" + prefix, and gives visit their names
// without storePrefix.
func listS3(s httpStore, base, storePrefix, prefix string, visit func(string) error) error {
	full := joinPrefix(storePrefix, prefix)
	token := ""
	for {
		q := url.Values{"list-type": {"2"}, "prefix": {full}}
		if token != "" {
			q.Set("continuation-token", token)
		}
		body, err := s.getPage(strings.TrimSuffix(base, "/")+"/?"+q.Encode(), full)
		if err != nil {
			return err
		}
		var page s3ListPage
		if err := xml.Unmarshal(body, &page); err != nil {
			return fmt.Errorf("the listing of %s could not be read: %w", full, err)
		}
		for _, c := range page.Contents {
			if err := visit(strings.TrimPrefix(c.Key, joinPrefix(storePrefix, ""))); err != nil {
				return err
			}
		}
		if !page.IsTruncated || page.NextContinuationToken == "" {
			return nil
		}
		token = page.NextContinuationToken
	}
}

// azureListPage is one page of List Blobs.
type azureListPage struct {
	Blobs struct {
		Blob []struct {
			Name string `xml:"Name"`
		} `xml:"Blob"`
	} `xml:"Blobs"`
	NextMarker string `xml:"NextMarker"`
}

// listAzure reads List Blobs of the container that starts storePrefix
// ("container/prefix"), page by page, and gives visit the names without the
// prefix. sas authorises each request.
func listAzure(s httpStore, base, storePrefix, sas, prefix string, visit func(string) error) error {
	container, blobPrefix, _ := strings.Cut(storePrefix, "/")
	full := joinPrefix(blobPrefix, prefix)
	marker := ""
	for {
		q := url.Values{"restype": {"container"}, "comp": {"list"}, "prefix": {full}}
		if marker != "" {
			q.Set("marker", marker)
		}
		target := strings.TrimSuffix(base, "/") + "/" + container + "?" + q.Encode() + "&" + sas
		body, err := s.getPage(target, container+"/"+full)
		if err != nil {
			return err
		}
		var page azureListPage
		if err := xml.Unmarshal(body, &page); err != nil {
			return fmt.Errorf("the listing of %s could not be read: %w", full, err)
		}
		for _, b := range page.Blobs.Blob {
			if err := visit(strings.TrimPrefix(b.Name, joinPrefix(blobPrefix, ""))); err != nil {
				return err
			}
		}
		if page.NextMarker == "" {
			return nil
		}
		marker = page.NextMarker
	}
}

// joinPrefix puts a store's prefix before a name: "team/cache" and "ac/" give
// "team/cache/ac/"; an empty store prefix gives the name.
func joinPrefix(storePrefix, name string) string {
	if storePrefix == "" {
		return name
	}
	return storePrefix + "/" + name
}
