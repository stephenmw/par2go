package par2

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

const MAX_FILENAME_LEN = 128

func parsepkt_Main(reader io.Reader, file *os.File, length int64) (updater RecoverySetUpdater, err error) {
	// read slice size
	raw := make([]byte, 12)
	_, err = io.ReadFull(reader, raw[:])
	if err != nil {
		return
	}

	sliceSize := int64(binary.LittleEndian.Uint64(raw[:8]))
	numFiles := binary.LittleEndian.Uint32(raw[8:12])

	// read file-ids
	file_ids := make([][16]byte, 0, numFiles)
	for i := uint32(0); i < numFiles; i++ {
		file_ids = file_ids[:len(file_ids)+1]
		_, err = io.ReadFull(reader, file_ids[len(file_ids)-1][:])
		if err != nil {
			return
		}
	}

	updater = func(r *RecoverySet) error {
		// apply actual changes
		r.FileIds = file_ids
		r.SliceSize = sliceSize

		return nil
	}

	return
}

func parsepkt_FileDesc(reader io.Reader, file *os.File, length int64) (updater RecoverySetUpdater, err error) {
	ret := File{}

	_, err = io.ReadFull(reader, ret.Id[:])
	if err != nil {
		return
	}

	_, err = io.ReadFull(reader, ret.Md5[:])
	if err != nil {
		return
	}

	_, err = io.ReadFull(reader, ret.Md5_16k[:])
	if err != nil {
		return
	}

	// read file size
	var fileSize [8]byte
	_, err = io.ReadFull(reader, fileSize[:])
	if err != nil {
		return
	}
	ret.Size = int64(binary.LittleEndian.Uint64(fileSize[:]))

	var fn_bytes = make([]byte, MAX_FILENAME_LEN)
	n, err := io.ReadAtLeast(reader, fn_bytes, 1)
	if err != nil {
		return
	}
	ret.Name = string(bytes.TrimRight(fn_bytes[:n], "\000"))

	updater = func(r *RecoverySet) error {
		// See if file was parsed already. If it was, ignore new data.
		for _, f := range r.Files {
			if bytes.Equal(f.Id[:], ret.Id[:]) {
				return nil
			}
		}

		// Apply changes
		r.Files = append(r.Files, ret)

		return nil
	}

	return
}

func parsepkt_IFSC(reader io.Reader, file *os.File, length int64) (updater RecoverySetUpdater, err error) {
	var ret FileCheckSums

	// Read file id
	_, err = io.ReadFull(reader, ret.FileId[:])
	if err != nil {
		return
	}

	// Read Checksums
	var raw [20]byte
	for {
		n, err := io.ReadFull(reader, raw[:])
		if err != nil {
			if n == 0 {
				break
			} else {
				return nil, err
			}
		}

		var s SliceChecksum
		copy(s.Md5[:16], raw[0:16])
		copy(s.Crc32[:4], raw[16:20])

		ret.Slices = append(ret.Slices, s)
	}

	updater = func(r *RecoverySet) error {
		// See if file was parsed already. If it was, ignore new data.
		for _, f := range r.IFSC {
			if bytes.Equal(f.FileId[:], ret.FileId[:]) {
				return nil
			}
		}

		// Apply changes
		r.IFSC = append(r.IFSC, ret)

		return nil
	}

	return
}
