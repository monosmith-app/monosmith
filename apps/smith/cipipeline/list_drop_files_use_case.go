package cipipeline

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// dropFile is a file the pack phase left in the drop directory.
type dropFile struct {
	Path string
	Size int64
}

// listDropFiles lists what is in the drop directory, workspace-relative.
func listDropFiles(root string) []dropFile {
	var files []dropFile
	base := filepath.Join(root, filepath.FromSlash(DropDir))
	_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		files = append(files, dropFile{Path: filepath.ToSlash(rel), Size: info.Size()})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

// dropTable is one row per packed file with its size.
func dropTable(files []dropFile) resultTable {
	t := resultTable{Headers: []string{"Packed file", "Size"}, Right: []bool{false, true}}
	for _, f := range files {
		t.Rows = append(t.Rows, tableRow{Cells: []string{f.Path, formatSize(f.Size)}})
	}
	return t
}

// formatSize is a size as a reader says it: 812 B, 4.2 KB, 12.4 MB.
func formatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	size, suffix := float64(n)/unit, "KB"
	for _, next := range []string{"MB", "GB"} {
		if size < unit {
			break
		}
		size, suffix = size/unit, next
	}
	return fmt.Sprintf("%.1f %s", size, suffix)
}

// totalSize adds the sizes.
func totalSize(files []dropFile) int64 {
	var n int64
	for _, f := range files {
		n += f.Size
	}
	return n
}
