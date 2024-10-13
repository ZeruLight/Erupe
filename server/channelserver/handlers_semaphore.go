package channelserver

import (
	"erupe-ce/utils/broadcast"
	"erupe-ce/utils/byteframe"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"erupe-ce/network/mhfpacket"
)

func removeSessionFromSemaphore(s *Session) {
	s.Server.semaphoreLock.Lock()
	for _, semaphore := range s.Server.semaphore {
		if _, exists := semaphore.clients[s]; exists {
			delete(semaphore.clients, s)
		}
	}
	s.Server.semaphoreLock.Unlock()
}

func handleMsgSysCreateSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateSemaphore)
	broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x03, 0x00, 0x0d})
}

func destructEmptySemaphores(s *Session) {
	s.Server.semaphoreLock.Lock()
	for id, sema := range s.Server.semaphore {
		if len(sema.clients) == 0 {
			delete(s.Server.semaphore, id)
			if strings.HasPrefix(id, "hs_l0") {
				s.Server.resetRaviente()
			}
			s.Logger.Debug("Destructed semaphore", zap.String("sema.name", id))
		}
	}
	s.Server.semaphoreLock.Unlock()
}

func handleMsgSysDeleteSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysDeleteSemaphore)
	destructEmptySemaphores(s)
	s.Server.semaphoreLock.Lock()
	for id, sema := range s.Server.semaphore {
		if sema.id == pkt.SemaphoreID {
			for session := range sema.clients {
				if s == session {
					delete(sema.clients, s)
				}
			}
			if len(sema.clients) == 0 {
				delete(s.Server.semaphore, id)
				if strings.HasPrefix(id, "hs_l0") {
					s.Server.resetRaviente()
				}
				s.Logger.Debug("Destructed semaphore", zap.String("sema.name", id))
			}
		}
	}
	s.Server.semaphoreLock.Unlock()
}

func handleMsgSysCreateAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateAcquireSemaphore)
	SemaphoreID := pkt.SemaphoreID

	if s.Server.HasSemaphore(s) {
		s.semaphoreMode = !s.semaphoreMode
	}
	if s.semaphoreMode {
		s.semaphoreID[1]++
	} else {
		s.semaphoreID[0]++
	}

	newSemaphore, exists := s.Server.semaphore[SemaphoreID]
	if !exists {
		s.Server.semaphoreLock.Lock()
		if strings.HasPrefix(SemaphoreID, "hs_l0") {
			suffix, _ := strconv.Atoi(pkt.SemaphoreID[len(pkt.SemaphoreID)-1:])
			s.Server.semaphore[SemaphoreID] = &Semaphore{
				name:       pkt.SemaphoreID,
				id:         uint32((suffix + 1) * 0x10000),
				clients:    make(map[*Session]uint32),
				maxPlayers: 127,
			}
		} else {
			s.Server.semaphore[SemaphoreID] = NewSemaphore(s, SemaphoreID, 1)
		}
		newSemaphore = s.Server.semaphore[SemaphoreID]
		s.Server.semaphoreLock.Unlock()
	}

	newSemaphore.Lock()
	defer newSemaphore.Unlock()
	bf := byteframe.NewByteFrame()
	if _, exists := newSemaphore.clients[s]; exists {
		bf.WriteUint32(newSemaphore.id)
	} else if uint16(len(newSemaphore.clients)) < newSemaphore.maxPlayers {
		newSemaphore.clients[s] = s.CharID
		s.Lock()
		s.semaphore = newSemaphore
		s.Unlock()
		bf.WriteUint32(newSemaphore.id)
	} else {
		bf.WriteUint32(0)
	}
	broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgSysAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysAcquireSemaphore)
	if sema, exists := s.Server.semaphore[pkt.SemaphoreID]; exists {
		sema.host = s
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(sema.id)
		broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		broadcast.DoAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgSysReleaseSemaphore(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysReleaseSemaphore)
}

func handleMsgSysCheckSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCheckSemaphore)
	resp := []byte{0x00, 0x00, 0x00, 0x00}
	s.Server.semaphoreLock.Lock()
	if _, exists := s.Server.semaphore[pkt.SemaphoreID]; exists {
		resp = []byte{0x00, 0x00, 0x00, 0x01}
	}
	s.Server.semaphoreLock.Unlock()
	broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, resp)
}
