package choicedata

import "io/fs"

// MustLoad is Load for the data a plugin embeds in its own binary. That data
// is part of the program, so a file that does not load is a bug in the plugin,
// and MustLoad panics with the error. The plugin's tests load it, so the
// panic never reaches a person.
func MustLoad(fsys fs.FS, dir string) *Set {
	set, err := Load(fsys, dir)
	if err != nil {
		panic("choices: " + err.Error())
	}
	return set
}
