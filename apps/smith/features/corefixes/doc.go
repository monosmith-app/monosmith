// Package corefixes writes the fixes of smith's own doctor checks that need
// more than a file's new content: stop tracking generated files, and give a
// project without one a lint task. Each returns a protocol.Fix; doctorfix
// shows and applies it.
package corefixes
