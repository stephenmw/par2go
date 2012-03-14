package par2

import (
	"encoding/binary"
	"io"
)

func parse_main_packet(file io.Reader) RecoverySetUpdater {
	var errs []error

	// read slice size
	var raw_slice_size [8]byte
	file.Read(raw_slice_size[:])
	slice_size := binary.LittleEndian.Uint64(raw_slice_size[:])

	// read number of files in recovery set.
	var raw_num_files [4]byte
	file.Read(raw_num_files[:])
	num_files := binary.LittleEndian.Uint32(raw_num_files[:])

	// read file-ids
	file_ids := make([][16]byte, 0, num_files)
	for i := uint32(0); i < num_files; i++ {
		file_ids = file_ids[:len(file_ids)+1]
		_, _ = file.Read(file_ids[len(file_ids)-1][:])
	}

	var ret RecoverySetUpdater = func(r *RecoverySet) error {
		if len(errs) > 0 {
			// if we ran into errors, we should not be updating r
			return errs[0]
		}

		// apply actual changes
		r.FileIds = file_ids
		r.SliceSize = slice_size

		return nil
	}

	return ret
}
