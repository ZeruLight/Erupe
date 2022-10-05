package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// TerminalLogEntry represents an entry in the MSG_SYS_TERMINAL_LOG packet.
type TerminalLogEntry struct {
	Index uint32
	Type1 uint8
	Type2 uint8
	Data  []int16
}

// MsgSysTerminalLog represents the MSG_SYS_TERMINAL_LOG
type MsgSysTerminalLog struct {
	AckHandle  uint32
	LogID      uint32 // 0 on the first packet, and the server sends back a value to use for subsequent requests.
	EntryCount uint16
	Unk0       uint16 // Hardcoded 0 in the binary
	Entries    []*TerminalLogEntry
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysTerminalLog) Opcode() network.PacketID {
	return network.MSG_SYS_TERMINAL_LOG
}

// Parse parses the packet from binary
func (m *MsgSysTerminalLog) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.LogID = bf.ReadUint32()
	m.EntryCount = bf.ReadUint16()
	m.Unk0 = bf.ReadUint16()

	for i := 0; i < int(m.EntryCount); i++ {
		e := &TerminalLogEntry{}
		e.Index = bf.ReadUint32()
		e.Type1 = bf.ReadUint8()
		e.Type2 = bf.ReadUint8()
		for j := 0; j < 15; j++ {
			e.Data = append(e.Data, bf.ReadInt16())
		}
		m.Entries = append(m.Entries, e)
	}

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysTerminalLog) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
