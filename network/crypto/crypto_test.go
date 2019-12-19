package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
)

var commonTestData = []byte{0x74, 0x65, 0x73, 0x74}
var tests = []struct {
	decryptedData      []byte
	key                uint32
	encryptedData      []byte
	ecc, ec0, ec1, ec2 uint16
}{
	{
		commonTestData,
		0,
		[]byte{0x46, 0x53, 0x28, 0x5E},
		0x2976, 0x06ea, 0x0215, 0x08FB3,
	},
	{
		commonTestData,
		3,
		[]byte{0x46, 0x95, 0x88, 0xEA},
		0x2AE4, 0x0A56, 0x01CD, 0x08FB3,
	},
	/*
		// TODO(Andoryuuta): This case fails. Debug the client and figure out if this is valid expected data.
		{
			commonTestData,
			995117,
			[]byte{0x46, 0x28, 0xFF, 0xAA},
			0x2A22, 0x09D4, 0x014C, 0x08FB3,
		},
	*/
	{
		commonTestData,
		0x7FFFFFFF,
		[]byte{0x46, 0x53, 0x28, 0x5E},
		0x2976, 0x06ea, 0x0215, 0x08FB3,
	},
	{
		commonTestData,
		0x80000000,
		[]byte{0x46, 0x95, 0x88, 0xEA},
		0x2AE4, 0x0A56, 0x01CD, 0x08FB3,
	},
	{
		commonTestData,
		0xFFFFFFFF,
		[]byte{0x46, 0xB5, 0xDC, 0xB2},
		0x2ADD, 0x09A6, 0x021E, 0x08FB3,
	},
}

func TestEncrypt(t *testing.T) {
	for k, tt := range tests {
		testname := fmt.Sprintf("encrypt_test_%d", k)
		t.Run(testname, func(t *testing.T) {
			out, cc, c0, c1, c2 := Encrypt(tt.decryptedData, tt.key, nil)
			if cc != tt.ecc {
				t.Errorf("got cc 0x%X, want 0x%X", cc, tt.ecc)
			} else if c0 != tt.ec0 {
				t.Errorf("got c0 0x%X, want 0x%X", c0, tt.ec0)
			} else if c1 != tt.ec1 {
				t.Errorf("got c1 0x%X, want 0x%X", c1, tt.ec1)
			} else if c2 != tt.ec2 {
				t.Errorf("got c2 0x%X, want 0x%X", c2, tt.ec2)
			} else if !bytes.Equal(out, tt.encryptedData) {
				t.Errorf("got out\n\t%s\nwant\n\t%s", hex.Dump(out), hex.Dump(tt.encryptedData))
			}
		})
	}

}

func TestDecrypt(t *testing.T) {
	for k, tt := range tests {
		testname := fmt.Sprintf("decrypt_test_%d", k)
		t.Run(testname, func(t *testing.T) {
			out, cc, c0, c1, c2 := Decrypt(tt.encryptedData, tt.key, nil)
			if cc != tt.ecc {
				t.Errorf("got cc 0x%X, want 0x%X", cc, tt.ecc)
			} else if c0 != tt.ec0 {
				t.Errorf("got c0 0x%X, want 0x%X", c0, tt.ec0)
			} else if c1 != tt.ec1 {
				t.Errorf("got c1 0x%X, want 0x%X", c1, tt.ec1)
			} else if c2 != tt.ec2 {
				t.Errorf("got c2 0x%X, want 0x%X", c2, tt.ec2)
			} else if !bytes.Equal(out, tt.decryptedData) {
				t.Errorf("got out\n\t%s\nwant\n\t%s", hex.Dump(out), hex.Dump(tt.decryptedData))
			}
		})
	}

}
