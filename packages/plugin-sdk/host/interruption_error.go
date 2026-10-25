package host

import "errors"

// interruption is the error of a call the host gave up on, because its context
// was cancelled or its time ran out. It unwraps to the context error
// (context.Canceled or context.DeadlineExceeded) and, when the plugin
// answered after being asked to stop, to the plugin's answer as well, so
// errors.Is and errors.As see both.
type interruption struct {
	message string
	causes  []error
	killed  bool
}

func (e *interruption) Error() string   { return e.message }
func (e *interruption) Unwrap() []error { return e.causes }

// WasKilled reports whether a call ended with the host killing the plugin
// because it did not answer in time. The plugin's process, and every other
// call in flight to it, is gone.
func WasKilled(err error) bool {
	var i *interruption
	return errors.As(err, &i) && i.killed
}
