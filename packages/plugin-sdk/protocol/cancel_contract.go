package protocol

// CancelParams are the params of the cancel notification.
type CancelParams struct {
	// ID is the id of the in-flight request to stop. An id the plugin does
	// not know, or whose request is already answered, is ignored.
	ID int64 `json:"id"`
}
