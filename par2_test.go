package par2

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

var main_test_answer = RecoverySet{
	FileIds: [][16]byte{
		[16]byte{166, 65, 188, 209, 94, 154, 93, 94, 177, 154, 9, 36, 214, 119, 123, 231},
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

	if rs.SliceSize != main_test_answer.SliceSize {
		t.Error("Slice size incorrect")
	}

	if len(rs.FileIds) != len(main_test_answer.FileIds) {
		t.Errorf("Length of FileIds was %d, expected %d\n",
			len(rs.FileIds),
			len(main_test_answer.FileIds),
		)
	} else {
		for i, _ := range rs.FileIds {
			if !bytes.Equal(rs.FileIds[i][:], main_test_answer.FileIds[i][:]) {
				t.Errorf("FileIds[%d] was %s, expected %s\n", i,
					hex.EncodeToString(rs.FileIds[i][:]),
					hex.EncodeToString(main_test_answer.FileIds[i][:]),
				)
			}
		}
	}
}
