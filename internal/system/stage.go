package system

import (
	"sync"

	"erupe-ce/network/mhfpacket"
)

type SessionStage interface {
	QueueSendMHFLazy(packet mhfpacket.MHFPacket)
	GetCharID() uint32
	GetName() string
}

// Object holds infomation about a specific object.
type Object struct {
	sync.RWMutex
	Id          uint32
	OwnerCharID uint32
	X, Y, Z     float32
}

// stageBinaryKey is a struct used as a map key for identifying a stage binary part.
type StageBinaryKey struct {
	Id0 uint8
	Id1 uint8
}

// Stage holds stage-specific information
type Stage struct {
	sync.RWMutex

	// Stage ID string
	Id string

	// Objects
	Objects     map[uint32]*Object
	objectIndex uint8

	// Map of session -> charID.
	// These are clients that are CURRENTLY in the stage
	Clients map[SessionStage]uint32

	// Map of charID -> bool, key represents whether they are ready
	// These are clients that aren't in the stage, but have reserved a slot (for quests, etc).
	ReservedClientSlots map[uint32]bool

	// These are raw binary blobs that the stage owner sets,
	// other clients expect the server to echo them back in the exact same format.
	RawBinaryData map[StageBinaryKey][]byte

	Host       SessionStage
	MaxPlayers uint16
	Password   string
	Locked     bool
}

// NewStage creates a new stage with intialized values.
func NewStage(ID string) *Stage {
	s := &Stage{
		Id:                  ID,
		Clients:             make(map[SessionStage]uint32),
		ReservedClientSlots: make(map[uint32]bool),
		Objects:             make(map[uint32]*Object),
		objectIndex:         0,
		RawBinaryData:       make(map[StageBinaryKey][]byte),
		MaxPlayers:          127,
	}
	return s
}

// BroadcastMHF queues a MHFPacket to be sent to all sessions in the stage.
func (s *Stage) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession SessionStage) {
	s.Lock()
	defer s.Unlock()
	for session := range s.Clients {
		if session == ignoredSession {
			continue
		}
		session.QueueSendMHFLazy(pkt)
	}
}

func (s *Stage) isCharInQuestByID(charID uint32) bool {
	if _, exists := s.ReservedClientSlots[charID]; exists {
		return exists
	}

	return false
}

func (s *Stage) IsQuest() bool {
	return len(s.ReservedClientSlots) > 0
}
