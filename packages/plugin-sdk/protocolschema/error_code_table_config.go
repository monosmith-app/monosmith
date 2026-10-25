package protocolschema

// ErrorCodeSpec says what an error code of the protocol means to a plugin.
type ErrorCodeSpec struct {
	// Const is the name of the constant in the protocol package. The code
	// itself is read from there, so it cannot differ from the Go SDK's.
	Const       string
	Description string
}

// ErrorCodes are the codes of the error member of a response. Build checks the
// list against the Code constants of the protocol package.
var ErrorCodes = []ErrorCodeSpec{
	{"CodeParseError", "The message was not valid JSON. The response has a null id."},
	{"CodeInvalidRequest", "The request is not valid, or arrived out of order: every request before initialize is refused with this code."},
	{"CodeMethodNotFound", "The plugin does not implement the method, or did not advertise the capability. The message is \"method not supported: <method>\"."},
	{"CodeInvalidParams", "The params do not have the shape the method expects."},
	{"CodeInternalError", "The plugin failed in a way that is not the user's to act on, for example a result that could not be encoded."},
	{"CodePluginFailure", "The method ran and failed for a reason the user must act on (a missing toolchain, a bad manifest). The message is shown to the user as it is, so it should say what to do. An error a method of the Go SDK returns becomes this code."},
}
