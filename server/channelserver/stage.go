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

// stageBinaryKey is a struct used as a map key for identifying a stage binary part.
type stageBinaryKey struct {
	id0 uint8
	id1 uint8
}

// Stage holds stage-specific information
type Stage struct {
	sync.RWMutex

	// Stage ID string
	id string

	// Total count of objects ever created for this stage. Used for ObjID generation.
	gameObjectCount uint32

	// Map of ObjID -> StageObject
	objects map[uint32]*StageObject

	// Map of session -> charID.
	// These are clients that are CURRENTLY in the stage
	clients map[*Session]uint32

	// Map of charID -> interface{}, only the key is used, value is always nil.
	// These are clients that aren't in the stage, but have reserved a slot (for quests, etc).
	reservedClientSlots map[uint32]interface{}

	// These are raw binary blobs that the stage owner sets,
	// other clients expect the server to echo them back in the exact same format.
	rawBinaryData map[stageBinaryKey][]byte

	maxPlayers  uint16
	hasDeparted bool
	password    string
}

// NewStage creates a new stage with intialized values.
func NewStage(ID string) *Stage {
	s := &Stage{
		id:                  ID,
		objects:             make(map[uint32]*StageObject),
		clients:             make(map[*Session]uint32),
		reservedClientSlots: make(map[uint32]interface{}),
		rawBinaryData:       make(map[stageBinaryKey][]byte),
		maxPlayers:          4,
		gameObjectCount:     1,
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
