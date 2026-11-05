// Package pathfinder is the browsable list that sits under a text field where
// the person types a path. It shows the folders and files of the folder being
// typed, narrows them as the text grows, and completes a name with Tab. The
// text field stays the owner of the value: a plain typed path always works,
// the list only helps.
//
// The host calls Sync with the text after every change, gives the list the
// keys it understands (Handle), and draws it with Rows. A folder is read only
// when the typed text enters it, and never more than a few thousand names, so
// a huge folder or a deep tree does not slow the screen.
package pathfinder
