package protocol

// ConcurrencyUnlimited is the Concurrency value that leaves the number of
// simultaneous requests to the host.
const ConcurrencyUnlimited = -1

// ConcurrencyValid reports whether a declared concurrency is one the protocol
// defines: absent (0), a count (1 and above) or unlimited (-1).
func ConcurrencyValid(n int) bool { return n >= ConcurrencyUnlimited }

// ConcurrencyLimit reads a declared concurrency: how many requests may be in
// flight at once, or unlimited. An absent, zero or invalid value is 1, the
// default of every plugin that says nothing.
func ConcurrencyLimit(n int) (limit int, unlimited bool) {
	switch {
	case n == ConcurrencyUnlimited:
		return 0, true
	case n < 1:
		return 1, false
	default:
		return n, false
	}
}
