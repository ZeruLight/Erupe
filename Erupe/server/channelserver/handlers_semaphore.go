package channelserver

import (
	"fmt"
	"strings"

	"erupe-ce/network/mhfpacket"
)

func removeSessionFromSemaphore(s *Session) {
	s.server.semaphoreLock.Lock()
	for _, semaphore := range s.server.semaphore {
		if _, exists := semaphore.clients[s]; exists {
			delete(semaphore.clients, s)
		}
	}
	releaseRaviSemaphore(s)
	s.server.semaphoreLock.Unlock()
}

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
				} else if s.server.semaphore[id].id_semaphore == "hs_l0u3B5129k3" {
					delete(s.server.semaphore["hs_l0u3B5129k3"].reservedClientSlots, s.charID)
					delete(s.server.semaphore["hs_l0u3B5129k3"].clients, s)
				} else if s.server.semaphore[id].id_semaphore == "hs_l0u3B512Ak3" {
					delete(s.server.semaphore["hs_l0u3B512Ak3"].reservedClientSlots, s.charID)
					delete(s.server.semaphore["hs_l0u3B512Ak3"].clients, s)
				}
			case 851997:
				if s.server.semaphore[id].id_semaphore == "hs_l0u3B51J9k4" {
					delete(s.server.semaphore["hs_l0u3B51J9k4"].reservedClientSlots, s.charID)
				} else if s.server.semaphore[id].id_semaphore == "hs_l0u3B5129k4" {
					delete(s.server.semaphore["hs_l0u3B5129k4"].reservedClientSlots, s.charID)
				} else if s.server.semaphore[id].id_semaphore == "hs_l0u3B512Ak4" {
					delete(s.server.semaphore["hs_l0u3B512Ak4"].reservedClientSlots, s.charID)
				}
			case 786461:
				if s.server.semaphore[id].id_semaphore == "hs_l0u3B51J9k5" {
					delete(s.server.semaphore["hs_l0u3B51J9k5"].reservedClientSlots, s.charID)
				} else if s.server.semaphore[id].id_semaphore == "hs_l0u3B5129k5" {
					delete(s.server.semaphore["hs_l0u3B5129k5"].reservedClientSlots, s.charID)
				} else if s.server.semaphore[id].id_semaphore == "hs_l0u3B512Ak5" {
					delete(s.server.semaphore["hs_l0u3B512Ak5"].reservedClientSlots, s.charID)
				}
			default:
				if len(s.server.semaphore[id].reservedClientSlots) != 0 {
					if s.server.semaphore[id].id_semaphore != "hs_l0u3B51J9k3" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B51J9k4" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B51J9k5" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B5129k3" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B5129k4" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B5129k5" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B512Ak3" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B512Ak4" &&
					s.server.semaphore[id].id_semaphore != "hs_l0u3B512Ak5" {
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
		if strings.HasPrefix(SemaphoreID, "hs_l0u3B51") {
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
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0F, 0x00, 0x1D})
	} else if uint16(len(newSemaphore.reservedClientSlots)) < newSemaphore.maxPlayers {
		switch SemaphoreID {
		case "hs_l0u3B51J9k3", "hs_l0u3B5129k3", "hs_l0u3B512Ak3":
			newSemaphore.reservedClientSlots[s.charID] = nil
			newSemaphore.clients[s] = s.charID
			s.Lock()
			s.semaphore = newSemaphore
			s.Unlock()
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0E, 0x00, 0x1D})
		case "hs_l0u3B51J9k4", "hs_l0u3B5129k4", "hs_l0u3B512Ak4":
			newSemaphore.reservedClientSlots[s.charID] = nil
			s.Lock()
			s.semaphore = newSemaphore
			s.Unlock()
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x0D, 0x00, 0x1D})
		case "hs_l0u3B51J9k5", "hs_l0u3B5129k5", "hs_l0u3B512Ak5":
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
	releaseRaviSemaphore(s)
}

func handleMsgSysCheckSemaphore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCheckSemaphore)
	resp := []byte{0x00, 0x00, 0x00, 0x00}
	s.server.semaphoreLock.Lock()
	if _, exists := s.server.semaphore[pkt.StageID]; exists {
		resp = []byte{0x00, 0x00, 0x00, 0x01}
	}
	s.server.semaphoreLock.Unlock()
	doAckSimpleSucceed(s, pkt.AckHandle, resp)
}
