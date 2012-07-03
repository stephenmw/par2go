package par2

import (
	"bytes"
	"sort"
)

// Sort a File slice by the Id field
type fileById []*File

func (xs fileById) Len() int {
	return len(xs)
}

func (xs fileById) Less(i, j int) bool {
	return bytes.Compare(xs[i].Id[:], xs[j].Id[:]) < 0
}

func (xs fileById) Swap(i, j int) {
	xs[i], xs[j] = xs[j], xs[i]
}

// Sort a File slice by filename
type fileByName []*File

func (xs fileByName) Len() int {
	return len(xs)
}

func (xs fileByName) Less(i, j int) bool {
	return xs[i].Name < xs[j].Name
}

func (xs fileByName) Swap(i, j int) {
	xs[i], xs[j] = xs[j], xs[i]
}

// FilesSortedByName returns a new slice of file pointers sorted by their name.
func FilesSortedByName(in []*File) (ret []*File) {
	ret = make([]*File, len(in))
	copy(ret, in)

	sort.Sort(fileByName(ret))

	return
}

// Sort a FileCheckSums slice by the FileId field
type ifscByFileId []*FileCheckSums

func (xs ifscByFileId) Len() int {
	return len(xs)
}

func (xs ifscByFileId) Less(i, j int) bool {
	return bytes.Compare(xs[i].FileId[:], xs[j].FileId[:]) < 0
}

func (xs ifscByFileId) Swap(i, j int) {
	xs[i], xs[j] = xs[j], xs[i]
}
