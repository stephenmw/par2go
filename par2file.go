package main

import (
	"os"
	"io"
	"bytes"
	"fmt"
	"crypto/md5"
	"encoding/binary"
)

var packet_seq = []byte("PAR2\000PKT")

type Packet struct {
	io.ReadSeeker
}

type RecoverySet struct {
	slice_size uint64
	num_files uint32
	file_ids [][16]byte
}

func (r *RecoverySet) ReadRecoveryFile(file io.ReadSeeker) error {
	for {
		err := r.readNextPacket(file)
		if err != nil {
			return nil
		}
	}

	return nil
}

// seekNextPacket reads from file until it has read the magic packet sequence.
// The error from the reader is returned. If no error is found, a packet was. 
func seekNextPacket(file io.Reader) error {
	for pos:=0; pos<len(packet_seq); {
		var b [1]byte

		read, err := file.Read(b[:1])

		if err != nil || read <= 0 {
			return err
		}

		switch b[0] {
		case packet_seq[pos]:
			pos++
		case packet_seq[0]:
			pos = 1
		default:
			pos = 0
		}
	}

	return nil
}

// readNextPacket finds and then reads the next packet in file.
func (r *RecoverySet) readNextPacket(file io.ReadSeeker) error {
	err := seekNextPacket(file)
	if err != nil {
		return err
	}

	// read packet size
	var raw_pkt_size [8]byte
	file.Read(raw_pkt_size[:])
	pkt_size := binary.LittleEndian.Uint64(raw_pkt_size[:])

	// read md5 of packet
	var md5sum [16]byte
	file.Read(md5sum[:])

	// wrap file in a md5 calculator and limit the amount that is readable
	hasher := md5.New()
	pkt_reader := io.TeeReader(io.LimitReader(file, int64(pkt_size)-32), hasher)

	var setid [16]byte
	pkt_reader.Read(setid[:])

	var pkt_type = make([]byte, 16)
	pkt_reader.Read(pkt_type)
	pkt_type = bytes.TrimRight(pkt_type, "\000")

	// empty pkt_reader
	var buffer [32*1024]byte
	for {
		_, err := pkt_reader.Read(buffer[:])
		if err != nil {
			break
		}
	}

	var hash_good bool
	if bytes.Equal(md5sum[:], hasher.Sum([]byte{})) {
		hash_good = true
	}

	if hash_good {
		fmt.Printf("%s good\n", string(md5sum[:]))
	} else{
		fmt.Printf("Set %s: %s %d\n", string(setid[:]), string(pkt_type), pkt_size)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please give a single par2 file to read.")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := new(RecoverySet)
	err = r.ReadRecoveryFile(file)

	fmt.Println(err)
	pos, _ := file.Seek(0, os.SEEK_CUR)
	fmt.Println(pos)
}
