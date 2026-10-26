package adoptflow

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// CountFiles is how many files the repository in dir has (tracked, and
// untracked ones that are not ignored): what reading it looks through.
func CountFiles(dir string) (int, error) {
	files, err := gitclient.New(dir).Files()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// ReadingLine is the progress line of reading a repository of n files, in the
// "… name" form the dialog shows as the step that is running.
func ReadingLine(n int) string {
	switch n {
	case 0:
		return "… reading the repository"
	case 1:
		return "… reading 1 file"
	}
	return "… reading " + groupDigits(n) + " files"
}

// groupDigits writes n with a comma between thousands: 12480 is "12,480".
func groupDigits(n int) string {
	s := strconv.Itoa(n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}
