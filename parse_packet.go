package par2

import (
	"bytes"
	"encoding/binary"
	"io"
)

const MAX_FILENAME_LEN = 128

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

func parsepkt_FileDesc(file io.Reader) (updater RecoverySetUpdater, err error) {
	var n int
	_ = n

	ret := File{}

	n, err = file.Read(ret.Id[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}

	n, err = file.Read(ret.Md5[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}

	n, err = file.Read(ret.Md5_16k[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}

	// read file size
	var raw_file_size [8]byte
	n, err = file.Read(raw_file_size[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}
	ret.Size = binary.LittleEndian.Uint64(raw_file_size[:])

	var fn_bytes = make([]byte, MAX_FILENAME_LEN)
	n, err = file.Read(fn_bytes)
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
