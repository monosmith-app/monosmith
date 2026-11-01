package convert

import (
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// CheckConverted opens the converted workspace and reports the doctor checks
// the plugins hold on it as bugs in smith convert (doctor.ReportCommandBugs).
// A workspace that cannot be opened reports nothing here: the conversion
// already checked that it loads.
//
// Parameters: root is the workspace; version is the running smith's version;
// out is where it reports.
// Returns: the failed checks.
func CheckConverted(root, version string, out io.Writer) []doctor.Finding {
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard, AllowCycles: true})
	if err != nil {
		return nil
	}
	defer s.Close()
	return doctor.ReportCommandBugs(s, "smith convert", out)
}
