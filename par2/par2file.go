package par2

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

var packet_seq = []byte("PAR2\000PKT")

var (
	ErrUnexpectedEndOfPacket = errors.New("par2: unexpected end of packet")
	ErrUnknownPktType        = errors.New("par2: packet of unknown type found")
	ErrMismatchedRecoverySet = errors.New("par2: packet for unknown recovery set found")
)

// PacketParsers are functions that parse the internal structure of packets and
// return a RecoverySetUpdater. The function takes an io.Reader that contains
// ONLY the internal packet data, no headers and no extra data.
type PacketParser func(io.Reader, *os.File, int64) (RecoverySetUpdater, error)

// RecoverySetUpdaters are closures that update a RecoverySet. They are
// returned by PacketParsers.
type RecoverySetUpdater func(*RecoverySet) error

type RecoverySet struct {
	Id        [16]byte
	SliceSize int64
	FileIds   [][16]byte
	Files     []*File
	IFSC      []*FileCheckSums
}

type File struct {
	Id      [16]byte
	Md5     [16]byte
	Md5_16k [16]byte
	Size    int64
	Name    string

	// actual data of file
	Fp io.ReaderAt
}

// CheckFile returns a list of corrupted par2 slices in a file. The first slice
// is zero
func (r RecoverySet) CheckFile(id [16]byte) (corrupt []int, err error) {
	f := r.fileById(id)
	c := r.ifscByFileId(id)
	if f == nil || c == nil {
		panic("better error handling needed")
	}
	slices := c.Slices

	hasher := md5.New()
	sum := make([]byte, md5.Size)

	for i, checksum := range slices {
		hasher.Reset()
		slice := io.NewSectionReader(f.Fp, r.SliceSize*int64(i), r.SliceSize)
		n, err := io.Copy(hasher, slice)
		if err != nil {
			return corrupt, err
		}

		zero := []byte{0}
		for i := int64(0); i < r.SliceSize-n; i++ {
			hasher.Write(zero)
		}

		if !bytes.Equal(checksum.Md5[:], hasher.Sum(sum[:0])) {
			corrupt = append(corrupt, i)
			fmt.Printf("damn it %s %s\n",
				hex.EncodeToString(checksum.Md5[:]),
				hex.EncodeToString(sum),
			)
		}
	}

	return
}

type FileCheckSums struct {
	FileId [16]byte
	Slices []SliceChecksum
}

type SliceChecksum struct {
	Md5   [16]byte
	Crc32 [4]byte
}

// ReadRecoveryFile parses a par2 file.
func (r *RecoverySet) ReadRecoveryFile(file *os.File) (errors []error) {
	defer r.sort()

	for {
		err := r.readNextPacket(file)
		if err == io.EOF {
			return
		} else if err != nil {
			errors = append(errors, err)
		}
	}

	panic("unreachable")
}

// seekNextPacket reads from file until it has read the magic packet sequence.
// The error from the reader is returned. If no error is found, a packet was. 
func seekNextPacket(file io.Reader) error {
	for pos := 0; pos < len(packet_seq); {
		var b [1]byte

		read, err := file.Read(b[:1])

		//BUG(smw): if read <= 0 and err == nil, nil is returned
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
func (r *RecoverySet) readNextPacket(file *os.File) (err error) {
	err = seekNextPacket(file)
	if err != nil {
		return
	}

	raw := make([]byte, 56)
	_, err = io.ReadFull(file, raw)
	if err != nil {
		if err == io.EOF {
			err = ErrUnexpectedEndOfPacket
		}
		return
	}

	packetSize := int64(binary.LittleEndian.Uint64(raw[0:8]))
	packetMd5 := raw[8:24]
	recoverySetId := raw[24:40]
	packetType := string(bytes.TrimRight(raw[40:56], "\000"))

	if r.Id == [16]byte{} {
		copy(r.Id[:], recoverySetId[:])
	}

	if !bytes.Equal(recoverySetId, r.Id[:]) {
		return ErrMismatchedRecoverySet
	}

	// wrap file in a md5 calculator and limit the amount that is readable
	hasher := md5.New()
	hasher.Write(raw[24:]) // hash starts at beginning of recoverysetid
	pkt_reader := io.TeeReader(io.LimitReader(file, int64(packetSize)-64), hasher)

	var parser PacketParser
	switch packetType {
	case "PAR 2.0\000Main":
		parser = parsepkt_Main
	case "PAR 2.0\000FileDesc":
		parser = parsepkt_FileDesc
	case "PAR 2.0\000IFSC":
		parser = parsepkt_IFSC
	}
	if parser == nil {
		return ErrUnknownPktType
	}

	updater, err := parser(pkt_reader, file, packetSize)

	if err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrUnexpectedEndOfPacket
		}
		return err
	}

	// empty pkt_reader
	io.Copy(ioutil.Discard, pkt_reader)

	if bytes.Equal(packetMd5, hasher.Sum([]byte{})) {
		err = updater(r)
		return
	}

	return nil
}

func (r *RecoverySet) sort() {
	sort.Sort(fileById(r.Files))
	sort.Sort(ifscByFileId(r.IFSC))
}

func (r *RecoverySet) fileById(id [16]byte) *File {
	for _, f := range r.Files {
		if f.Id == id {
			return f
		}
	}

	return nil
}

func (r *RecoverySet) ifscByFileId(id [16]byte) *FileCheckSums {
	for _, c := range r.IFSC {
		if c.FileId == id {
			return c
		}
	}

	return nil
}

// IsComplete returns true if all metadata/checksum information has been found.
func (r *RecoverySet) IsComplete() bool {
	if len(r.FileIds) == 0 {
		return false
	}

	if len(r.Files) != len(r.FileIds) || len(r.IFSC) != len(r.FileIds) {
		return false
	}

	for i := range r.FileIds {
		if r.Files[i].Id != r.FileIds[i] || r.IFSC[i].FileId != r.FileIds[i] {
			return false
		}
	}

	return true
}
