package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func handleMsgSysOperateRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysOperateRegister)
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
	s.server.raviente.Lock()
	if pkt.SemaphoreID == s.server.raviente.state.semaphoreID {
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
						*ref += data
					} else {
						resp.WriteUint32(*ref + data)
					}
				} else {
					resp.WriteUint32(*ref + data*damageMultiplier)
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
	} else if pkt.SemaphoreID == s.server.raviente.support.semaphoreID {
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
	} else if pkt.SemaphoreID == s.server.raviente.register.semaphoreID {
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
	}
	s.notifyall()
	s.server.raviente.Unlock()
}

func handleMsgSysLoadRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLoadRegister)
	r := pkt.Unk1
	switch r {
	case 12:
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

func (s *Session) notifyall() {
	var temp mhfpacket.MHFPacket
	raviNotif := byteframe.NewByteFrame()
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: s.server.raviente.support.semaphoreID}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: s.server.raviente.state.semaphoreID}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: s.server.raviente.register.semaphoreID}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	raviNotif.WriteUint16(0x0010) // End it.
	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
		for session := range s.server.semaphore["hs_l0u3B51J9k3"].clients {
			session.QueueSend(raviNotif.Data())
		}
	} else if _, exists := s.server.semaphore["hs_l0u3B5129k3"]; exists {
		for session := range s.server.semaphore["hs_l0u3B5129k3"].clients {
			session.QueueSend(raviNotif.Data())
		}
	} else if _, exists := s.server.semaphore["hs_l0u3B512Ak3"]; exists {
		for session := range s.server.semaphore["hs_l0u3B512Ak3"].clients {
			session.QueueSend(raviNotif.Data())
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

//func releaseRaviSemaphore(s *Session) {
//	s.server.raviente.Lock()
//	if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
//		if len(s.server.semaphore["hs_l0u3B51J9k3"].reservedClientSlots) == 0 {
//			resetRavi(s)
//		}
//	}
//	if _, exists := s.server.semaphore["hs_l0u3B5129k3"]; exists {
//		if len(s.server.semaphore["hs_l0u3B5129k3"].reservedClientSlots) == 0 {
//			resetRavi(s)
//		}
//	}
//	if _, exists := s.server.semaphore["hs_l0u3B512Ak3"]; exists {
//		if len(s.server.semaphore["hs_l0u3B512Ak3"].reservedClientSlots) == 0 {
//			resetRavi(s)
//		}
//	}
//	s.server.raviente.Unlock()
//}

func resetRavi(s *Session) {
	s.server.raviente.Lock()
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
	s.server.raviente.Unlock()
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
