package par2

import (
	"testing"
	"os"
	"encoding/hex"
)

var fileid, _ = hex.DecodeString("12")
var main_test_answer = RecoverySet {
	FileIds: [][16]byte{[16]byte{},
		//[16]byte("\xa6\x41\xbc\xd1\x5e\x9a\x5d\x5e\xb1\x9a\x09\x24\xd6\x77\x7b\xe7"),
	},
	SliceSize: 362644,
}

func TestParseMainPacket(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/only_main.par2")
	if err != nil {
		t.Fatal("Failed to find test data.")
	}

	rs := new(RecoverySet)
	err = rs.ReadRecoveryFile(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(rs.FileIds) != len(main_test_answer.FileIds) {
		t.Error("Length of FileIds incorrect")
	}

	if rs.SliceSize != main_test_answer.SliceSize {
		t.Error("Slice size incorrect")
	}
}
