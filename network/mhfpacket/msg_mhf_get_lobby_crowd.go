package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetLobbyCrowd represents the MSG_MHF_GET_LOBBY_CROWD
type MsgMhfGetLobbyCrowd struct {
	AckHandle uint32
	Server    uint32
	Room      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetLobbyCrowd) Opcode() network.PacketID {
	return network.MSG_MHF_GET_LOBBY_CROWD
}

// Parse parses the packet from binary
func (m *MsgMhfGetLobbyCrowd) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Server = bf.ReadUint32()
	m.Room = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetLobbyCrowd) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
