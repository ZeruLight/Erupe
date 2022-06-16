package channelserver

import (
	"encoding/hex"
	"github.com/Andoryuuta/byteframe"
	"github.com/Solenataris/Erupe/network/mhfpacket"
)

func handleMsgSysOperateRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysOperateRegister)
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
	s.server.raviente.Lock()
	switch pkt.RegisterID {
	case 786461:
		resp := byteframe.NewByteFrame()
		size := 6
		for i := 0; i < len(bf.Data())-1; i += size {
			op := bf.ReadUint8()
			dest := bf.ReadUint8()
			data := bf.ReadUint32()
			resp.WriteUint8(1)
			resp.WriteUint8(dest)
			switch dest {
				case 0:
					resp.WriteUint32(0)
					resp.WriteUint32(data)
					s.server.raviente.register.nextTime = data
				case 1:
					resp.WriteUint32(0)
					resp.WriteUint32(data)
					s.server.raviente.register.startTime = data
				case 2:
					resp.WriteUint32(0)
					resp.WriteUint32(data)
					s.server.raviente.register.killedTime = data
				case 3:
					resp.WriteUint32(0)
					resp.WriteUint32(data)
					s.server.raviente.register.postTime = data
				case 4:
					ref := &s.server.raviente.register.register[0]
					switch op {
						case 2:
							resp.WriteUint32(*ref)
							resp.WriteUint32(*ref + uint32(data))
							*ref += data
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
							*ref = data
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
					}
				case 5:
					resp.WriteUint32(0)
					resp.WriteUint32(data)
					s.server.raviente.register.carveQuest = data
				case 6:
					ref := &s.server.raviente.register.register[1]
					switch op {
						case 2:
							resp.WriteUint32(*ref)
							resp.WriteUint32(*ref + uint32(data))
							*ref += data
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
							*ref = data
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
					}
				case 7:
					ref := &s.server.raviente.register.register[2]
					switch op {
						case 2:
							resp.WriteUint32(*ref)
							resp.WriteUint32(*ref + uint32(data))
							*ref += data
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
							*ref = data
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
					}
				case 8:
					ref := &s.server.raviente.register.register[3]
					switch op {
						case 2:
							resp.WriteUint32(*ref)
							resp.WriteUint32(*ref + uint32(data))
							*ref += data
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
							*ref = data
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
					}
				case 9:
					resp.WriteUint32(0)
					resp.WriteUint32(data)
					s.server.raviente.register.maxPlayers = data
				case 10:
					resp.WriteUint32(0)
					resp.WriteUint32(data)
					s.server.raviente.register.ravienteType = data
				case 11:
					ref := &s.server.raviente.register.register[4]
					switch op {
						case 2:
							resp.WriteUint32(*ref)
							resp.WriteUint32(*ref + uint32(data))
							*ref += data
						case 13:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
							*ref = data
						case 14:
							resp.WriteUint32(0)
							resp.WriteUint32(data)
					}
				default:
					resp.WriteUint32(0)
					resp.WriteUint32(0)
			}
		}
		resp.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())

	case 917533:
		resp := byteframe.NewByteFrame()
		size := 6
		for i := 0; i < len(bf.Data())-1; i += size {
			op := bf.ReadUint8()
			dest := bf.ReadUint8()
			data := bf.ReadUint32()
			resp.WriteUint8(1)
			resp.WriteUint8(dest)
			ref := &s.server.raviente.state.stateData[dest]
			damageMultiplier := s.server.raviente.state.damageMultiplier
			switch op {
			case 2:
				resp.WriteUint32(*ref)
				if dest == 28 { // Berserk resurrection tracker
					resp.WriteUint32(*ref + data)
					*ref += data
				} else if dest == 17 { // Berserk poison tracker
					if damageMultiplier == 1 {
						resp.WriteUint32(*ref + data)
					} else {
						resp.WriteUint32(*ref + data)
						*ref += data
					}
				} else {
					resp.WriteUint32(*ref + data * damageMultiplier)
					*ref += data * damageMultiplier
				}
			case 13:
				fallthrough
			case 14:
				resp.WriteUint32(0)
				resp.WriteUint32(data)
				*ref = data
			}
		}
		resp.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())

	case 851997:
		resp := byteframe.NewByteFrame()
		size := 6
		for i := 0; i < len(bf.Data())-1; i += size {
			op := bf.ReadUint8()
			dest := bf.ReadUint8()
			data := bf.ReadUint32()
			resp.WriteUint8(1)
			resp.WriteUint8(dest)
			ref := &s.server.raviente.support.supportData[dest]
			switch op {
			case 2:
				resp.WriteUint32(*ref)
				resp.WriteUint32(*ref + data)
				*ref += data
			case 13:
				fallthrough
			case 14:
				resp.WriteUint32(0)
				resp.WriteUint32(data)
				*ref = data
			}
		}
		resp.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
	s.notifyall()
	s.server.raviente.Unlock()
}

func handleMsgSysLoadRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLoadRegister)
	r := pkt.Unk1
	switch r {
		case 12:
			if pkt.RegisterID == 983077 {
				data, _ := hex.DecodeString("000C000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
				doAckBufFail(s, pkt.AckHandle, data)
			} else if pkt.RegisterID == 983069 {
				data, _ := hex.DecodeString("000C000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
				doAckBufFail(s, pkt.AckHandle, data)
			}
			resp := byteframe.NewByteFrame()
			resp.WriteUint8(0)
			resp.WriteUint8(12)
			resp.WriteUint32(s.server.raviente.register.nextTime)
			resp.WriteUint32(s.server.raviente.register.startTime)
			resp.WriteUint32(s.server.raviente.register.killedTime)
			resp.WriteUint32(s.server.raviente.register.postTime)
			resp.WriteUint32(s.server.raviente.register.register[0])
			resp.WriteUint32(s.server.raviente.register.carveQuest)
			resp.WriteUint32(s.server.raviente.register.register[1])
			resp.WriteUint32(s.server.raviente.register.register[2])
			resp.WriteUint32(s.server.raviente.register.register[3])
			resp.WriteUint32(s.server.raviente.register.maxPlayers)
			resp.WriteUint32(s.server.raviente.register.ravienteType)
			resp.WriteUint32(s.server.raviente.register.register[4])
			doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		case 29:
			resp := byteframe.NewByteFrame()
			resp.WriteUint8(0)
			resp.WriteUint8(29)
			for _, v := range s.server.raviente.state.stateData {
				resp.WriteUint32(v)
			}
			doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		case 25:
			resp := byteframe.NewByteFrame()
			resp.WriteUint8(0)
			resp.WriteUint8(25)
			for _, v := range s.server.raviente.support.supportData {
				resp.WriteUint32(v)
			}
			doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

// Unused
func (s *Session) notifyplayer() {
	s.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0E, 0x00, 0x1D})
	s.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0D, 0x00, 0x1D})
	s.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0C, 0x00, 0x1D})
}

func (s *Session) notifyall() {
	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
		for session := range s.server.semaphore["hs_l0u3B51J9k3"].clients {
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0C, 0x00, 0x1D})
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0D, 0x00, 0x1D})
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0E, 0x00, 0x1D})
		}
	} else if _, exists := s.server.semaphore["hs_l0u3B5129k3"]; exists {
		for session := range s.server.semaphore["hs_l0u3B5129k3"].clients {
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0C, 0x00, 0x1D})
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0D, 0x00, 0x1D})
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0E, 0x00, 0x1D})
		}
	} else if _, exists := s.server.semaphore["hs_l0u3B512Ak3"]; exists {
		for session := range s.server.semaphore["hs_l0u3B512Ak3"].clients {
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0C, 0x00, 0x1D})
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0D, 0x00, 0x1D})
			session.QueueSendNonBlocking([]byte{0x00, 0x3F, 0x00, 0x0E, 0x00, 0x1D})
		}
	}
}

func checkRaviSemaphore(s *Session) bool {
	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
		return true
	} else if _, exists := s.server.semaphore["hs_l0u3B5129k3"]; exists {
		return true
	} else if _, exists := s.server.semaphore["hs_l0u3B512Ak3"]; exists {
		return true
	}
	return false
}

func releaseRaviSemaphore(s *Session) {
	s.server.raviente.Lock()
	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
		if len(s.server.semaphore["hs_l0u3B51J9k3"].reservedClientSlots) == 0 {
			resetRavi(s)
		}
	}
	if _, exists := s.server.semaphore["hs_l0u3B5129k3"]; exists {
		if len(s.server.semaphore["hs_l0u3B5129k3"].reservedClientSlots) == 0 {
			resetRavi(s)
		}
	}
	if _, exists := s.server.semaphore["hs_l0u3B512Ak3"]; exists {
		if len(s.server.semaphore["hs_l0u3B512Ak3"].reservedClientSlots) == 0 {
			resetRavi(s)
		}
	}
	s.server.raviente.Unlock()
}

func resetRavi(s *Session) {
	s.server.raviente.register.nextTime = 0
	s.server.raviente.register.startTime = 0
	s.server.raviente.register.killedTime = 0
	s.server.raviente.register.postTime = 0
	s.server.raviente.register.ravienteType = 0
	s.server.raviente.register.maxPlayers = 0
	s.server.raviente.register.carveQuest = 0
	s.server.raviente.state.damageMultiplier = 1
	s.server.raviente.register.register = []uint32{0, 0, 0, 0, 0}
	s.server.raviente.state.stateData = []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	s.server.raviente.support.supportData = []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

// Unused
func (s *Session) notifyticker() {
	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
		s.server.semaphoreLock.Lock()
		getSemaphore := s.server.semaphore["hs_l0u3B51J9k3"]
		s.server.semaphoreLock.Unlock()
		if _, exists := getSemaphore.reservedClientSlots[s.charID]; exists {
			s.notifyall()
		}
	}
}

func handleMsgSysNotifyRegister(s *Session, p mhfpacket.MHFPacket) {}