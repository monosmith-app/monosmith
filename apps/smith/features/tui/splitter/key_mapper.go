package splitter

// The keys of the dividers, by the name bubbletea gives them.
const (
	// KeyNext selects the next divider; one key that works in any terminal.
	KeyNext = `\`
	// KeyPrevious selects the previous divider.
	KeyPrevious = "|"
	// KeyLeft and KeyUp select the previous divider of their axis.
	KeyLeft = "shift+left"
	KeyUp   = "shift+up"
	// KeyRight and KeyDown select the next divider of their axis.
	KeyRight = "shift+right"
	KeyDown  = "shift+down"
	// KeyAdvance goes on to the next divider while one is selected.
	KeyAdvance = "tab"
	// KeyRelease lets go of the selected divider.
	KeyRelease = "esc"
)

// Step is how many cells one press of [ or ] moves a divider.
const Step = 4

// KeyAction says what a key does to the dividers.
type KeyAction int

const (
	// NoKey is a key the dividers do not use.
	NoKey KeyAction = iota
	// CycleKey selects the next or previous divider of any axis.
	CycleKey
	// SelectKey selects the next or previous divider of one axis.
	SelectKey
	// AdvanceKey goes on to the next divider; it counts only while one is selected.
	AdvanceKey
	// ReleaseKey lets go of the selected divider; it counts only while one is selected.
	ReleaseKey
)

// KeyIntent is what a key means: its action, and the axis and direction (-1 is
// previous, 1 is next) it acts in.
type KeyIntent struct {
	Action KeyAction
	Axis   Axis
	Dir    int
}

// IntentOf reads the divider meaning of a key name; a key that is not one of
// the divider keys has the action NoKey.
func IntentOf(name string) KeyIntent {
	switch name {
	case KeyNext:
		return KeyIntent{Action: CycleKey, Dir: 1}
	case KeyPrevious:
		return KeyIntent{Action: CycleKey, Dir: -1}
	case KeyLeft:
		return KeyIntent{Action: SelectKey, Axis: Column, Dir: -1}
	case KeyRight:
		return KeyIntent{Action: SelectKey, Axis: Column, Dir: 1}
	case KeyUp:
		return KeyIntent{Action: SelectKey, Axis: Row, Dir: -1}
	case KeyDown:
		return KeyIntent{Action: SelectKey, Axis: Row, Dir: 1}
	case KeyAdvance:
		return KeyIntent{Action: AdvanceKey, Dir: 1}
	case KeyRelease:
		return KeyIntent{Action: ReleaseKey}
	}
	return KeyIntent{}
}
