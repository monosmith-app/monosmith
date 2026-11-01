package pipeline

import (
	"encoding/json"
	"fmt"
)

// NpmRunner runs npm with the arguments and returns what it printed on
// stdout; npm audit exits non-zero whenever it finds anything, so the exit
// status carries nothing the gate reads.
type NpmRunner func(args ...string) string

// AuditNpm is the npm audit gate: false (fail) only for an advisory in the
// production tree (`--omit=dev`) at moderate or above whose fix npm can
// apply. Advisories that reach only dev dependencies, which no consumer
// receives, and those nobody can fix are printed as notes and pass. A report
// that is not JSON passes with the reason printed: a broken audit service
// must not stop a release. The dry run of `npm audit fix` runs only when
// something would block, to confirm there is a change to make.
func AuditNpm(npm NpmRunner, log func(string)) bool {
	log("Auditing dependencies with npm (asks the registry, this can take a minute)...")
	var whole, shipped AuditReport
	out := npm("audit", "--json")
	if err := json.Unmarshal([]byte(out), &whole); err != nil {
		log("npm audit produced no JSON - not blocking on a broken audit.")
		return true
	}
	if err := json.Unmarshal([]byte(npm("audit", "--omit=dev", "--json")), &shipped); err != nil {
		log("npm audit --omit=dev produced no JSON - not blocking on a broken audit.")
		return true
	}
	all := ClassifyAudit(whole)
	production := ClassifyAudit(shipped)
	for _, n := range all.Notes {
		log(n)
	}
	for _, a := range all.Blocking {
		if _, ok := shipped.Vulnerabilities[a.Name]; !ok {
			log(fmt.Sprintf("  note [%s] %s - reaches only dev dependencies, which no consumer receives, so it does not block", a.Severity, a.Name))
		}
	}
	if len(production.Blocking) == 0 {
		log(fmt.Sprintf("npm audit - %d advisory(ies), none actionable at moderate or above in production dependencies.", all.Total))
		return true
	}
	if PlanIsEmpty(npm("audit", "fix", "--omit=dev", "--dry-run", "--json")) {
		for _, a := range production.Blocking {
			log(fmt.Sprintf("  note [%s] %s - npm reports a fix, but npm audit fix has nothing to apply", a.Severity, a.Name))
		}
		log(fmt.Sprintf("npm audit - %d production advisory(ies); %d reported as fixable, but no patched release is reachable, so there is nothing to do here.", production.Total, len(production.Blocking)))
		return true
	}
	for _, a := range production.Blocking {
		log(fmt.Sprintf("  BLOCKING [%s] %s - fix available", a.Severity, a.Name))
	}
	log("Each has a published fix. Add a targeted overrides entry in package.json rather than npm audit fix --force.")
	return false
}
