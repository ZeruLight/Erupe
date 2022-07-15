package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgSysAcquireSemaphore represents the MSG_SYS_ACQUIRE_SEMAPHORE
type MsgSysAcquireSemaphore struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysAcquireSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_ACQUIRE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysAcquireSemaphore) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysAcquireSemaphore) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
