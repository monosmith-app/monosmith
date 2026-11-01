package apisurface

import (
	"reflect"
	"testing"
)

func TestExportedNames(t *testing.T) {
	files := map[string]string{
		"lib.go": `package lib

const Max = 3
var hidden = 1
type Client struct { Name string; secret int }
type reader interface{ Read() }
type Doer interface { Do(); done() }
func Dial() {}
func helper() {}
func (c *Client) Close() {}
func (c *Client) flush() {}
func (r reader) Exported() {}
`,
		"lib_test.go":          "package lib\nfunc TestX() {}\n",
		"client/client.go":     "package client\nfunc Open() {}\n",
		"internal/deep/x.go":   "package deep\nfunc Secret() {}\n",
		"cmd/tool/main.go":     "package main\nfunc Run() {}\n",
		"broken.go":            "package lib\nfunc (",
		"testdata/sample/s.go": "package sample\nfunc S() {}\n",
	}
	want := []string{"Client", "Client.Close", "Client.Name", "Dial", "Doer", "Doer.Do", "Max", "client:Open"}
	if got := ExportedNames(files); !reflect.DeepEqual(got, want) {
		t.Fatalf("names = %v\nwant    %v", got, want)
	}
}
