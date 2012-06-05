package par2

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"github.com/stephenmw/par2go/par2/iotools"
	"io"
	"io/ioutil"
	"os"
)

var packet_seq = []byte("PAR2\000PKT")

var (
	ErrUnexpectedEndOfPacket = errors.New("par2: unexpected end of packet")
	ErrUnknownPktType        = errors.New("par2: unknown packet type found")
)

// PacketParsers are functions that parse the internal structure of packets and
// return a RecoverySetUpdater. The function takes an io.Reader that contains
// ONLY the internal packet data, no headers and no extra data.
type PacketParser func(io.Reader) (RecoverySetUpdater, error)

// RecoverySetUpdaters are closures that update a RecoverySet. They are
// returned by PacketParsers.
type RecoverySetUpdater func(*RecoverySet) error

type RecoverySet struct {
	SliceSize uint64
	FileIds   [][16]byte
	Files     []File
}

type File struct {
	Id         [16]byte
	Md5        [16]byte
	Md5_16k    [16]byte
	Size       uint64
	Name       string
	SliceMd5   [][16]byte
	SliceCrc32 [][4]byte
}

func (r *RecoverySet) ReadRecoveryFile(file io.ReadSeeker) error {
	n, _ := file.Seek(0, os.SEEK_CUR)
	f := iotools.NewOffsetReader(bufio.NewReader(file), n)
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
func (r *RecoverySet) readNextPacket(file iotools.OffsetReader) (err error) {
	err = seekNextPacket(file)
	if err != nil {
		return
	}

	// read packet size
	var raw_pkt_size [8]byte
	_, err = io.ReadFull(file, raw_pkt_size[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}
	pkt_size := binary.LittleEndian.Uint64(raw_pkt_size[:])

	// read md5 of packet
	var md5sum [16]byte
	_, err = io.ReadFull(file, md5sum[:])
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}

	// wrap file in a md5 calculator and limit the amount that is readable
	hasher := md5.New()
	pkt_reader := io.TeeReader(io.LimitReader(file, int64(pkt_size)-32), hasher)

	var setid [16]byte
	_, err = io.ReadFull(pkt_reader, setid[:])
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}

	var pkt_type = make([]byte, 16)
	_, err = io.ReadFull(pkt_reader, pkt_type)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}
	pkt_type = bytes.TrimRight(pkt_type, "\000")

	var parser PacketParser
	switch string(pkt_type) {
	case "PAR 2.0\000Main":
		parser = parsepkt_Main
	case "PAR 2.0\000FileDesc":
		parser = parsepkt_FileDesc
	}
	if parser == nil {
		return ErrUnknownPktType
	}

	updater, err := parser(pkt_reader)
	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrUnexpectedEndOfPacket
		}
		return err
	}

	// empty pkt_reader
	io.Copy(ioutil.Discard, pkt_reader)

	if bytes.Equal(md5sum[:], hasher.Sum([]byte{})) {
		err := updater(r)
		return err
	}

	return nil
}
