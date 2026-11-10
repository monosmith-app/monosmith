// Package buildrefs finds the files that name a folder a tool builds into, so
// a doctor fix that moves the build can refuse, with the file named, when the
// move would break one of them (a Dockerfile that copies bin/Release, a
// pipeline that uploads target/release).
package buildrefs
