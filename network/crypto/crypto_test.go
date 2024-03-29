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
	{
		[]byte{0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x03, 0x02, 0x00, 0x6C, 0x6C, 0x00, 0x00, 0x00, 0x12, 0x00, 0xDE, 0x00, 0x03, 0x00, 0x00, 0x00, 0x30, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x00, 0x03, 0x20, 0x18, 0x46, 0x00, 0x00, 0x80, 0x3F, 0xDC, 0xE4, 0x0A, 0x46, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x67, 0xD3, 0x5B, 0x00, 0x77, 0x01, 0x78, 0x00, 0x77, 0x01, 0x4F, 0x01, 0x5B, 0x6F, 0x76, 0xC5, 0x30, 0x00, 0x02, 0x02, 0x00, 0x00, 0x00, 0x00, 0x2A, 0xDD, 0x17, 0x46, 0x00, 0x00, 0x80, 0x3F, 0x9E, 0x11, 0x0C, 0x46, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x37, 0x64, 0x00, 0x2C, 0x01, 0x64, 0x00, 0x2C, 0x01, 0x4F, 0x01, 0x5B, 0x6F, 0x76, 0xC5, 0x00, 0x10},
		2000476,
		[]byte{0x2E, 0x52, 0x24, 0xE3, 0x05, 0x2B, 0xFC, 0x04, 0x0B, 0x26, 0x90, 0xEA, 0x61, 0xDB, 0x8D, 0x27, 0xCB, 0xB1, 0x69, 0xA1, 0x77, 0x80, 0x4A, 0xC2, 0xA0, 0xBD, 0x50, 0x54, 0xF5, 0xC2, 0x94, 0x66, 0xBB, 0xCE, 0x53, 0x29, 0xEE, 0xB4, 0xFA, 0xF6, 0x5F, 0x8D, 0x80, 0x3E, 0x5D, 0x5F, 0xB0, 0x53, 0xE6, 0x92, 0x17, 0x80, 0xE7, 0xED, 0xE7, 0xDC, 0x61, 0xF0, 0xCD, 0xE4, 0x41, 0x82, 0x21, 0xBA, 0x47, 0xAB, 0x58, 0xFF, 0x30, 0x76, 0x80, 0x2D, 0x38, 0xF4, 0xDF, 0x86, 0x8C, 0x6C, 0x8D, 0x33, 0x4C, 0x37, 0xA3, 0xDA, 0x01, 0x3C, 0x98, 0x66, 0x1F, 0xB9, 0xE2, 0xEA, 0xF0, 0x84, 0xE8, 0xAA, 0x00, 0x3D, 0x4A, 0xB6, 0xF2, 0x3D, 0x91, 0x58, 0x4B, 0x0B, 0xE2, 0xD5, 0xC7, 0x39, 0x4D, 0x59, 0xED, 0xC3, 0x61, 0x6F, 0x6E, 0x69, 0x9B, 0x3C},
		0xCFF8, 0x086B, 0x3BAE, 0x4057,
	},
}

func TestEncrypt(t *testing.T) {
	for k, tt := range tests {
		testname := fmt.Sprintf("encrypt_test_%d", k)
		t.Run(testname, func(t *testing.T) {
			out, cc, c0, c1, c2 := Crypto(tt.decryptedData, tt.key, true, nil)
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
			out, cc, c0, c1, c2 := Crypto(tt.decryptedData, tt.key, false, nil)
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
