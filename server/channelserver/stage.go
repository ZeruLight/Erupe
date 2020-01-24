package channelserver

import (
	"sync"

	"github.com/Andoryuuta/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
)

// StageObject holds infomation about a specific stage object.
type StageObject struct {
	sync.RWMutex
	id          uint32
	ownerCharID uint32
	x, y, z     float32
}

// Stage holds stage-specific information
type Stage struct {
	sync.RWMutex
	id              string                  // Stage ID string
	gameObjectCount uint32                  // Total count of objects ever created for this stage. Used for ObjID generation.
	objects         map[uint32]*StageObject // Map of ObjID -> StageObject
	clients         map[*Session]uint32     // Map of session -> charID
}

// NewStage creates a new stage with intialized values.
func NewStage(ID string) *Stage {
	s := &Stage{
		objects: make(map[uint32]*StageObject),
		clients: make(map[*Session]uint32),
	}

	return s
}

// BroadcastMHF queues a MHFPacket to be sent to all sessions in the stage.
func (s *Stage) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	// Make the header
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(pkt.Opcode()))

	// Build the packet onto the byteframe.
	pkt.Build(bf)

	// Broadcast the data.
	for session := range s.clients {
		if session == ignoredSession {
			continue
		}
		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}
