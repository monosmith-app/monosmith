package convert

import "errors"

// ErrFindings is a conversion that finished and kept its changes, but whose
// closing checks found something (the lint, doctor). The report above lists
// each finding with its remedy; the command exits with code 2.
var ErrFindings = errors.New("the conversion finished with findings")
