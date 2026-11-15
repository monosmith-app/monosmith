package linuxpackages

import (
	"bytes"

	"github.com/google/rpmpack"
)

// RPM is the .rpm holding binary at /usr/bin/<p.Binary>.
func RPM(p Package, binary []byte) ([]byte, error) {
	summary := p.Summary
	if summary == "" {
		summary = p.Name
	}
	r, err := rpmpack.NewRPM(rpmpack.RPMMetaData{
		Name:        p.Name,
		Summary:     summary,
		Description: summary,
		Version:     packageVersion(p.Version),
		Release:     rpmRelease,
		Arch:        rpmArches[p.Arch],
		OS:          "linux",
		URL:         p.Homepage,
		Packager:    p.Maintainer,
		Group:       "Development/Tools",
		Licence:     p.License,
		Compressor:  "gzip",
		BuildTime:   stamp,
	})
	if err != nil {
		return nil, err
	}
	r.AddFile(rpmpack.RPMFile{Name: "/usr/bin/" + p.Binary, Body: binary, Mode: 0o100755, Owner: "root", Group: "root", MTime: uint32(stamp.Unix())})
	var out bytes.Buffer
	if err := r.Write(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
