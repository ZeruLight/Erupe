package channelserver

import (
	"fmt"

	"github.com/Solenataris/Erupe/network/mhfpacket"
)

func handleMsgSysCreateSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateSemaphore)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x03, 0x00, 0x0d})
}

func handleMsgSysDeleteSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysDeleteSemaphore)
	sem := pkt.AckHandle
	if s.server.semaphore != nil {
	s.server.semaphoreLock.Lock()
	for id := range s.server.semaphore {
		switch sem {
			case 917533:
				if s.server.semaphore[id].id_semaphore == "hs_l0u3B51J9k3" {
				delete(s.server.semaphore["hs_l0u3B51J9k3"].reservedClientSlots, s.charID)
				delete(s.server.semaphore["hs_l0u3B51J9k3"].clients, s)
				}
			case 851997:
				if s.server.semaphore[id].id_semaphore == "hs_l0u3B51J9k4" {
				delete(s.server.semaphore["hs_l0u3B51J9k4"].reservedClientSlots, s.charID)
				}
			case 786461:
				if s.server.semaphore[id].id_semaphore == "hs_l0u3B51J9k5" {
				delete(s.server.semaphore["hs_l0u3B51J9k5"].reservedClientSlots, s.charID)
				}
			default:
				if len(s.server.semaphore[id].reservedClientSlots) != 0 {
				if s.server.semaphore[id].id_semaphore != "hs_l0u3B51J9k3" && s.server.semaphore[id].id_semaphore != "hs_l0u3B51J9k4" && s.server.semaphore[id].id_semaphore != "hs_l0u3B51J9k5" {
					delete(s.server.semaphore[id].reservedClientSlots, s.charID)
					}
				}
			}
		}
	s.server.semaphoreLock.Unlock()
	}
}

func handleMsgSysCreateAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateAcquireSemaphore)
	SemaphoreID := pkt.SemaphoreID

	newSemaphore, gotNewStage := s.server.semaphore[SemaphoreID]

	fmt.Printf("Got reserve stage req, StageID: %v\n\n", SemaphoreID)
	if !gotNewStage {
		s.server.semaphoreLock.Lock()
		if SemaphoreID == "hs_l0u3B51J9k1" ||
			SemaphoreID == "hs_l0u3B51J9k2" ||
			SemaphoreID == "hs_l0u3B51J9k3" ||
			SemaphoreID == "hs_l0u3B51J9k4" ||
			SemaphoreID == "hs_l0u3B51J9k5" {
			s.server.semaphore[SemaphoreID] = NewSemaphore(SemaphoreID, 32)
		} else {
			s.server.semaphore[SemaphoreID] = NewSemaphore(SemaphoreID, 1)
		}
		newSemaphore = s.server.semaphore[SemaphoreID]
		s.server.semaphoreLock.Unlock()
	}

	newSemaphore.Lock()
	defer newSemaphore.Unlock()
	if _, exists := newSemaphore.reservedClientSlots[s.charID]; exists {
		s.logger.Info("IS ALREADY EXIST !")
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0F, 0x00, 0x1D})
	} else if uint16(len(newSemaphore.reservedClientSlots)) < newSemaphore.maxPlayers {
	switch SemaphoreID {
		case "hs_l0u3B51J9k3":
			newSemaphore.reservedClientSlots[s.charID] = nil
			newSemaphore.clients[s] = s.charID
			s.Lock()
			s.semaphore = newSemaphore
			s.Unlock()
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0E, 0x00, 0x1D})
		case "hs_l0u3B51J9k4":
			newSemaphore.reservedClientSlots[s.charID] = nil
			s.Lock()
			s.semaphore = newSemaphore
			s.Unlock()
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0D, 0x00, 0x1D})
		case "hs_l0u3B51J9k5":
			newSemaphore.reservedClientSlots[s.charID] = nil
			s.Lock()
			s.semaphore = newSemaphore
			s.Unlock()
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0C, 0x00, 0x1D})
		default:
			newSemaphore.reservedClientSlots[s.charID] = nil
			s.Lock()
			s.semaphore = newSemaphore
			s.Unlock()
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0F, 0x00, 0x25})
		}
	} else {
		doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgSysAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysAcquireSemaphore)
}

func handleMsgSysReleaseSemaphore(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysReleaseSemaphore)
	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
		reset := len(s.server.semaphore["hs_l0u3B51J9k3"].reservedClientSlots)
		if reset == 0 {
			s.server.db.Exec("CALL ravireset($1)", 0)
		}
	}
}

func removeSessionFromSemaphore(s *Session) {

	s.server.semaphoreLock.Lock()
	for id := range s.server.semaphore {
		delete(s.server.semaphore[id].reservedClientSlots, s.charID)
		if id == "hs_l0u3B51J9k3" {
			delete(s.server.semaphore[id].clients, s)
		} else {
			continue
		}
	}
	s.server.semaphoreLock.Unlock()
}

func handleMsgSysCheckSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCheckSemaphore)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
