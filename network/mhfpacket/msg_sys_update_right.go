package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"golang.org/x/exp/slices"
	"math"
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
	Aliases []string
	ID      uint16
	Value   uint32
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

func Courses() []Course {
	var courses = []Course{
		{Aliases: []string{"Trial", "TL"}, ID: 1},
		{Aliases: []string{"HunterLife", "HL"}, ID: 2},
		{Aliases: []string{"Extra", "ExtraA", "EX"}, ID: 3},
		{Aliases: []string{"ExtraB"}, ID: 4},
		{Aliases: []string{"Mobile"}, ID: 5},
		{Aliases: []string{"Premium"}, ID: 6},
		{Aliases: []string{"Pallone", "ExtraC"}, ID: 7},
		{Aliases: []string{"Assist", "Legend", "Rasta"}, ID: 8}, // Legend
		{Aliases: []string{"N"}, ID: 9},
		{Aliases: []string{"Hiden", "Secret"}, ID: 10},                                       // Secret
		{Aliases: []string{"HunterSupport", "HunterAid", "Support", "Aid", "Royal"}, ID: 11}, // Royal
		{Aliases: []string{"NBoost", "NetCafeBoost", "Boost"}, ID: 12},
		// 13-19 = (unknown), 20 = DEBUG, 21 = COG_LINK_EXPIRED, 22 = 360_GOLD, 23 = PS3_TROP
		{Aliases: []string{"COG"}, ID: 24},
		{Aliases: []string{"NetCafe", "Cafe", "InternetCafe"}, ID: 25},
		{Aliases: []string{"OfficialCafe", "Official"}, ID: 26},
		{Aliases: []string{"HLRenewing", "HLR", "HLRenewal", "HLRenew"}, ID: 27},
		{Aliases: []string{"EXRenewing", "EXR", "EXRenewal", "EXRenew"}, ID: 28},
		{Aliases: []string{"Free"}, ID: 29},
		// 30 = real netcafe bit
	}
	for i := range courses {
		courses[i].Value = uint32(math.Pow(2, float64(courses[i].ID)))
	}
	return courses
}

// GetCourseStruct returns a slice of Course(s) from a rights integer
func GetCourseStruct(rights uint32) []Course {
	var resp []Course
	s := Courses()
	slices.SortStableFunc(s, func(i, j Course) bool {
		return i.ID > j.ID
	})
	for _, course := range s {
		if rights-course.Value < 0x80000000 {
			resp = append(resp, course)
			rights -= course.Value
		}
	}
	return resp
}
