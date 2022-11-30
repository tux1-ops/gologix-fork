package gologix

import (
	"fmt"
	"testing"
)

func TestPath(t *testing.T) {
	var tests = []struct {
		path string
		want []byte
	}{
		{
			"1,0,2,172.25.58.11,1,1",
			[]byte{0x01, 0x00, 0x12, 0x0C, 0x31, 0x37, 0x32, 0x2E, 0x32, 0x35, 0x2E, 0x35, 0x38, 0x2E, 0x31, 0x31, 0x01, 0x01},
		},
		{
			"1,0,32,2,36,1",
			[]byte{0x01, 0x00, 0x20, 0x02, 0x24, 0x01},
		},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("path: %s", tt.path)
		t.Run(testname, func(t *testing.T) {
			res, err := ParsePath(tt.path)
			if err != nil {
				t.Errorf("Error in pathgen for %s. %v", tt.path, err)
			}
			if !check_bytes(res, tt.want) {
				t.Errorf("Wrong Value for result.  \nWanted %v. \nGot    %v", tt.want, res)
			}
		})
	}

}

func check_bytes(s0, s1 []byte) bool {
	if len(s1) != len(s0) {
		return false
	}
	for i := range s0 {
		if s0[i] != s1[i] {
			return false
		}

	}
	return true
}

func TestPathBuild(t *testing.T) {
	tests := []struct {
		name string
		path []Byteable
		want []byte
	}{
		{
			name: "connection manager only",
			path: []Byteable{CIPObject_ConnectionManager},
			want: []byte{0x20, 0x06},
		},
		{
			name: "connection manager instance 1",
			path: []Byteable{CIPObject_ConnectionManager, CIPInstance(1)},
			want: []byte{0x20, 0x06, 0x24, 0x01},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := BuildPath(tt.path...)
			if err != nil {
				t.Errorf("Problem building path. %v", err)
			}
			if !check_bytes(have.Bytes(), tt.want) {
				t.Errorf("ResultMismatch.\n Have %v\n Want %v\n", have.Bytes(), tt.want)
			}
		})
	}

}
