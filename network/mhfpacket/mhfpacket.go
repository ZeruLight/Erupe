package mhfpacket

import (
	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// Parser is the interface that wraps the Parse method.
type Parser interface {
	Parse(bf *byteframe.ByteFrame) error
}

// Builder is the interface that wraps the Build method.
type Builder interface {
	Build(bf *byteframe.ByteFrame) error
}

// Opcoder is the interface that wraps the Opcode method.
type Opcoder interface {
	Opcode() network.PacketID
}

// MHFPacket is the interface that groups the Parse, Build, and Opcode methods.
type MHFPacket interface {
	Parser
	Builder
	Opcoder
}
