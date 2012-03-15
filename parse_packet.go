package par2

import (
	"encoding/binary"
	"io"
)

func parsepkt_Main(file io.Reader) (updater RecoverySetUpdater, err error) {
	var n int
	_ = n

	// read slice size
	var raw_slice_size [8]byte
	n, err = file.Read(raw_slice_size[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}
	slice_size := binary.LittleEndian.Uint64(raw_slice_size[:])

	// read number of files in recovery set.
	var raw_num_files [4]byte
	n, err = file.Read(raw_num_files[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}
	num_files := binary.LittleEndian.Uint32(raw_num_files[:])

	// read file-ids
	file_ids := make([][16]byte, 0, num_files)
	for i := uint32(0); i < num_files; i++ {
		file_ids = file_ids[:len(file_ids)+1]
		n, err = file.Read(file_ids[len(file_ids)-1][:])
		if err != nil {
			if err == io.EOF {
				err = ErrUnexpectedEndOfPacket
			}
			return
		}
	}

	updater = func(r *RecoverySet) error {
		// apply actual changes
		r.FileIds = file_ids
		r.SliceSize = slice_size

		return nil
	}

	return
}
