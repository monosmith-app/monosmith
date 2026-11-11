package remotecache

import (
	"errors"
	"fmt"
	"net/http"
)

// StatusError is a request the remote store answered with a status other than
// the one expected.
type StatusError struct {
	// Method is the HTTP method of the request; Name is the object or listing it asked for.
	Method, Name string
	// Code is the HTTP status code; Status is the status line ("403 Forbidden").
	Code   int
	Status string
}

// Error implements error: "HEAD ac/1f2e…: 403 Forbidden".
func (e *StatusError) Error() string { return fmt.Sprintf("%s %s: %s", e.Method, e.Name, e.Status) }

func statusError(method, name string, resp *http.Response) error {
	return &StatusError{Method: method, Name: name, Code: resp.StatusCode, Status: resp.Status}
}

// Refused reports whether err is the store refusing the credentials: a 401 or
// a 403 answer.
//
// Parameters: err is an error from an ObjectStore method.
// Returns: true for a refusal, false for any other error and for nil.
func Refused(err error) bool {
	var se *StatusError
	return errors.As(err, &se) && (se.Code == http.StatusUnauthorized || se.Code == http.StatusForbidden)
}
