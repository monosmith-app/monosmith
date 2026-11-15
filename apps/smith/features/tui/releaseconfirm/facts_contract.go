package releaseconfirm

// TestRun is how the last test run of the session ended.
type TestRun int

// The ways the last test run can have ended.
const (
	// TestsNone means no test run happened in this session.
	TestsNone TestRun = iota
	// TestsPassed means the last test run passed.
	TestsPassed
	// TestsFailed means the last test run failed.
	TestsFailed
)

// Facts are what the shell knows about the workspace when it asks to release.
type Facts struct {
	// DoctorKnown is false until the doctor has checked the workspace;
	// DoctorFindings is how many checks found a problem.
	DoctorKnown    bool
	DoctorFindings int
	// Tests is how the last test run ended.
	Tests TestRun
	// GitKnown is false when the repository was not read. Branch is the current
	// branch; Remote receives the tags and HasRemote says it exists; Upstream is
	// true when the branch follows a remote branch, and Ahead counts the commits
	// that the remote branch lacks.
	GitKnown  bool
	Branch    string
	Remote    string
	HasRemote bool
	Upstream  bool
	Ahead     int
	// NoPush is true when the command line keeps the tags on this computer.
	NoPush bool
}

// Check is one line of the list shown before the release: Warn is true when it
// is not what a release needs. A warning never blocks: the person decides.
type Check struct {
	Text string
	Warn bool
}
