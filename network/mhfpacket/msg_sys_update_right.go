package mhfpacket

import (
	"errors"
	ps "erupe-ce/common/pascalstring"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

/*
00 58 // Opcode

00 00 00 00
00 00 00 4e

00 04 // Count
00 00 // Skipped(padding?)

00 01  00 00  00 00 00 00
00 02  00 00  5d fa 14 c0
00 03  00 00  5d fa 14 c0
00 06  00 00  5d e7 05 10

00 00 // Count of some buf up to 0x800 bytes following it.

00 10 // Trailer
*/

// ClientRight represents a right that the client has.
type ClientRight struct {
	ID        uint16
	Unk0      uint16
	Timestamp uint32
}

type Course struct {
	Name  string
	ID    uint16
	Value uint32
}

// MsgSysUpdateRight represents the MSG_SYS_UPDATE_RIGHT
type MsgSysUpdateRight struct {
	ClientRespAckHandle uint32 // If non-0, requests the client to send back a MSG_SYS_ACK packet with this value.
	Bitfield            uint32
	Rights              []ClientRight
	UnkSize             uint16 // Count of some buf up to 0x800 bytes following it.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysUpdateRight) Opcode() network.PacketID {
	return network.MSG_SYS_UPDATE_RIGHT
}

// Parse parses the packet from binary
func (m *MsgSysUpdateRight) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysUpdateRight) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.ClientRespAckHandle)
	bf.WriteUint32(m.Bitfield)
	bf.WriteUint16(uint16(len(m.Rights)))
	bf.WriteUint16(0)
	for _, v := range m.Rights {
		bf.WriteUint16(v.ID)
		bf.WriteUint16(v.Unk0)
		bf.WriteUint32(v.Timestamp)
	}
	ps.Uint16(bf, "", false) // update client login token / password in the game's launcherstate struct
	return nil
}

// GetCourseStruct returns a slice of Course(s) from a rights integer
func GetCourseStruct(rights uint32) []Course {
	var courses = []Course{
		{"Trial", 1, 0x00000002},
		{"HunterLife", 2, 0x00000004},
		{"ExtraA", 3, 0x00000008},
		{"ExtraB", 4, 0x00000010},
		{"Mobile", 5, 0x00000020},
		{"Premium", 6, 0x00000040},
		{"Pallone", 7, 0x00000080},
		{"Assist", 8, 0x00000100}, // Legend
		{"Netcafe", 9, 0x00000200},
		{"Hiden", 10, 0x00000400},         // Secret
		{"HunterSupport", 11, 0x00000800}, // Royal
		{"NetcafeBoost", 12, 0x00001000},
	}
	var resp []Course
	for _, course := range courses {
		if rights-course.Value < 0x80000000 {
			resp = append(resp, course)
			rights -= course.Value
		}
	}
	return resp
}
