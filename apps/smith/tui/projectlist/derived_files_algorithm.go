package projectlist

// The things a person can do about the derived files of a project.
const (
	// DeriveExplain says who owns the files and why they are in their state.
	DeriveExplain = "explain"
	// DeriveApply writes the files again from what smith derives.
	DeriveApply = "apply"
	// DeriveAdopt keeps the hand edit of a keys-mode file as the new baseline.
	DeriveAdopt = "adopt"
	// DeriveEject makes the files the person's own: smith stops touching them.
	DeriveEject = "eject"
	// DeriveReplace overwrites an edited or different file with what smith derives.
	DeriveReplace = "replace"
)

// StatusWords says in plain words what a status means for the person.
//
// status is one of the Status constants. An unknown status is returned as it is.
func StatusWords(status string) string {
	switch status {
	case StatusCurrent:
		return "up to date"
	case StatusStale:
		return "out of date: derive rewrites it"
	case StatusMissing:
		return "missing: derive writes it back"
	case StatusUnrecorded:
		return "not recorded yet: derive records it"
	case StatusDrift:
		return "edited by hand: derive keeps your edit until you choose"
	case StatusDiverged:
		return "differs from what smith derives: kept until you choose"
	case StatusBroken:
		return "smith cannot edit it: see Explain"
	case StatusEjected:
		return "yours: smith does not touch it"
	case StatusDiffers:
		return "the project's own: it differs from its siblings"
	}
	return status
}

// NeedsDecision is true for a file derive never changes by itself, because
// someone edited it or it differs: the person chooses between keeping it and
// letting smith write it.
func (f DerivedFile) NeedsDecision() bool {
	return f.Status == StatusDrift || f.Status == StatusDiverged || f.Status == StatusBroken
}

// DerivedMatching lists the paths of the derived files of the project for
// which keep is true.
func (p Project) DerivedMatching(keep func(DerivedFile) bool) []string {
	var paths []string
	for _, f := range p.Derived {
		if keep(f) {
			paths = append(paths, f.Path)
		}
	}
	return paths
}

// DeriveArgv is the `smith derive` command line for an action on the derived
// files of the project, or nil when no file of the project takes it: Explain
// every file, Adopt the hand edits of keys-mode files, Eject or Replace the
// files that wait for a choice. Apply is the plain `smith derive`, which acts
// on the whole workspace.
func (p Project) DeriveArgv(action string) []string {
	if action == DeriveApply {
		return []string{"derive"}
	}
	flag, keep := "", func(DerivedFile) bool { return true }
	switch action {
	case DeriveExplain:
		flag = "--explain"
	case DeriveAdopt:
		flag, keep = "--adopt", func(f DerivedFile) bool { return f.Status == StatusDrift && f.Keys }
	case DeriveEject:
		flag, keep = "--eject", DerivedFile.NeedsDecision
	case DeriveReplace:
		flag, keep = "--replace", DerivedFile.NeedsDecision
	default:
		return nil
	}
	paths := p.DerivedMatching(keep)
	if len(paths) == 0 {
		return nil
	}
	argv := []string{"derive"}
	for _, path := range paths {
		argv = append(argv, flag, path)
	}
	return argv
}

// DerivedNeedingAction counts the derived files that derive would write or
// that wait for a choice.
func (p Project) DerivedNeedingAction() int {
	n := 0
	for _, f := range p.Derived {
		if f.Status != StatusCurrent && f.Status != StatusEjected && f.Status != StatusDiffers {
			n++
		}
	}
	return n
}
