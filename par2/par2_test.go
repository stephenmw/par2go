package par2

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

var ans_MainPkt = RecoverySet{
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

	if rs.SliceSize != ans_MainPkt.SliceSize {
		t.Errorf("Slice size was %d, expected %d\n",
			rs.SliceSize,
			ans_MainPkt.SliceSize,
		)
	}

	if len(rs.FileIds) != len(ans_MainPkt.FileIds) {
		t.Errorf("Length of FileIds was %d, expected %d\n",
			len(rs.FileIds),
			len(ans_MainPkt.FileIds),
		)
	} else {
		for i, _ := range rs.FileIds {
			if !bytes.Equal(rs.FileIds[i][:], ans_MainPkt.FileIds[i][:]) {
				t.Errorf("FileIds[%d] was %s, expected %s\n", i,
					hex.EncodeToString(rs.FileIds[i][:]),
					hex.EncodeToString(ans_MainPkt.FileIds[i][:]),
				)
			}
		}
	}
}

var ans_FileDescPkt = RecoverySet{
	Files: []File{
		File{
			Id:      [16]byte{166, 65, 188, 209, 94, 154, 93, 94, 177, 154, 9, 36, 214, 119, 123, 231},
			Md5:     [16]byte{194, 58, 178, 255, 18, 2, 60, 104, 79, 70, 252, 192, 44, 87, 181, 133},
			Md5_16k: [16]byte{150, 38, 45, 214, 72, 173, 178, 226, 53, 82, 168, 142, 214, 20, 227, 183},
			Size:    725106140,
			Name:    "big_buck_bunny_1080p_h264.mov",
		},
	},
}

func TestParseFileDescPacket(t *testing.T) {
	t.Parallel()

	f, err := os.Open("testdata/only_file.par2")
	if err != nil {
		t.Fatal("Failed to find test data.")
	}

	rs := new(RecoverySet)
	err = rs.ReadRecoveryFile(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(rs.Files) != len(ans_FileDescPkt.Files) {
		t.Fatal("Length of Files slice was %d, expected %d\n",
			len(rs.Files),
			len(ans_FileDescPkt.Files),
		)
	}

	for i, _ := range rs.Files {
		res, ans := rs.Files[i], ans_FileDescPkt.Files[i]

		if !bytes.Equal(res.Id[:], ans.Id[:]) {
			t.Errorf("Files[%d].Id was %s, expected %s\n", i,
				hex.EncodeToString(res.Id[:]),
				hex.EncodeToString(ans.Id[:]),
			)
		}

		if !bytes.Equal(res.Md5[:], ans.Md5[:]) {
			t.Errorf("Files[%d].Md5 was %s, expected %s\n", i,
				hex.EncodeToString(res.Md5[:]),
				hex.EncodeToString(ans.Md5[:]),
			)
		}

		if !bytes.Equal(res.Md5_16k[:], ans.Md5_16k[:]) {
			t.Errorf("Files[%d].Md5_16k was %s, expected %s\n", i,
				hex.EncodeToString(res.Md5_16k[:]),
				hex.EncodeToString(ans.Md5_16k[:]),
			)
		}

		if res.Size != ans.Size {
			t.Errorf("Files[%d].Size was %d, expected %d\n", i,
				res.Size,
				ans.Size,
			)
		}

		if res.Name != ans.Name {
			t.Errorf("Files[%d].Name was `%s`, expected `%s`", i,
				res.Name,
				ans.Name,
			)
		}
	}
}
