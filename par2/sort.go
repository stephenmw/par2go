package par2

import (
	"bytes"
)

// Sort a File slice by the Id field
type fileById []File

func (xs fileById) Len() int {
	return len(xs)
}

func (xs fileById) Less(i, j int) bool {
	return bytes.Compare(xs[i].Id[:], xs[j].Id[:]) < 0
}

func (xs fileById) Swap(i, j int) {
	xs[i], xs[j] = xs[j], xs[i]
}

// Sort a FileCheckSums slice by the FileId field
type ifscByFileId []FileCheckSums

func (xs ifscByFileId) Len() int {
	return len(xs)
}

func (xs ifscByFileId) Less(i, j int) bool {
	return bytes.Compare(xs[i].FileId[:], xs[j].FileId[:]) < 0
}

func (xs ifscByFileId) Swap(i, j int) {
	xs[i], xs[j] = xs[j], xs[i]
}
