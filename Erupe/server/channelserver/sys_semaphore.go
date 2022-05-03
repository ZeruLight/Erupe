package channelserver

import (
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network/mhfpacket"

	"sync"
)

// Stage holds stage-specific information
type Semaphore struct {
	sync.RWMutex

	// Stage ID string
	id_semaphore string

	// Map of session -> charID.
	// These are clients that are CURRENTLY in the stage
	clients map[*Session]uint32

	// Map of charID -> interface{}, only the key is used, value is always nil.
	reservedClientSlots map[uint32]interface{}

	// Max Players for Semaphore
	maxPlayers uint16
}

// NewStage creates a new stage with intialized values.
func NewSemaphore(ID string, MaxPlayers uint16) *Semaphore {
	s := &Semaphore{
		id_semaphore:        ID,
		clients:             make(map[*Session]uint32),
		reservedClientSlots: make(map[uint32]interface{}),
		maxPlayers:          MaxPlayers,
	}
	return s
}

func (s *Semaphore) BroadcastRavi(pkt mhfpacket.MHFPacket) {
	// Broadcast the data.
	for session := range s.clients {


		// Make the header
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(pkt.Opcode()))

		// Build the packet onto the byteframe.
		pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}

// BroadcastMHF queues a MHFPacket to be sent to all sessions in the stage.
func (s *Semaphore) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	// Broadcast the data.
	for session := range s.clients {
		if session == ignoredSession {
			continue
		}

		// Make the header
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(pkt.Opcode()))

		// Build the packet onto the byteframe.
		pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}