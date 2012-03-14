package par2

import (
	"./extendedio"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

var packet_seq = []byte("PAR2\000PKT")

// PacketParsers are functions that parse the internal structure of packets and
// return a RecoverySetUpdater. The function takes an io.Reader that contains
// ONLY the internal packet data, no headers and no extra data.
type PacketParser func(io.Reader) RecoverySetUpdater

// RecoverySetUpdaters are closures that update a RecoverySet. They are
// returned by PacketParsers.
type RecoverySetUpdater func(*RecoverySet) error

type RecoverySet struct {
	SliceSize uint64
	FileIds   [][16]byte
	Files     []File
}

type File struct {
	Md5     [16]byte
	Md5_16k [16]byte
	Size    uint64
	Name    string
}

func (r *RecoverySet) ReadRecoveryFile(file io.ReadSeeker) error {
	n, _ := file.Seek(0, os.SEEK_CUR)
	f := extendedio.NewOffsetReader(bufio.NewReader(file), n)
	for {
		err := r.readNextPacket(f)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// seekNextPacket reads from file until it has read the magic packet sequence.
// The error from the reader is returned. If no error is found, a packet was. 
func seekNextPacket(file io.Reader) error {
	for pos := 0; pos < len(packet_seq); {
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
func (r *RecoverySet) readNextPacket(file extendedio.OffsetReader) error {
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

	var parser PacketParser
	switch string(pkt_type) {
	case "PAR 2.0\000Main":
		parser = parse_main_packet
	}
	updater := parser(pkt_reader)

	// empty pkt_reader
	var buffer [1024]byte
	for {
		_, err := pkt_reader.Read(buffer[:])
		if err != nil {
			break
		}
	}

	if bytes.Equal(md5sum[:], hasher.Sum([]byte{})) {
		err := updater(r)
		return err
	} else {
		return nil
	}

	panic("unreachable")
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
