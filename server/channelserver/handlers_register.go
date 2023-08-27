package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
	"strings"
)

type RaviUpdate struct {
	Op   uint8
	Dest uint8
	Data uint32
}

func handleMsgSysOperateRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysOperateRegister)

	var raviUpdates []RaviUpdate
	var raviUpdate RaviUpdate
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
	for i := len(pkt.RawDataPayload) / 6; i > 0; i-- {
		raviUpdate.Op = bf.ReadUint8()
		raviUpdate.Dest = bf.ReadUint8()
		raviUpdate.Data = bf.ReadUint32()
		s.logger.Debug("RaviOps", zap.Uint8s("Op/Dest", []uint8{raviUpdate.Op, raviUpdate.Dest}), zap.Uint32s("Sema/Data", []uint32{pkt.SemaphoreID, raviUpdate.Data}))
		raviUpdates = append(raviUpdates, raviUpdate)
	}
	bf = byteframe.NewByteFrame()

	var _old, _new uint32
	s.server.raviente.Lock()
	for _, update := range raviUpdates {
		switch update.Op {
		case 2:
			_old, _new = s.server.UpdateRavi(pkt.SemaphoreID, update.Dest, update.Data, true)
		case 13, 14:
			_old, _new = s.server.UpdateRavi(pkt.SemaphoreID, update.Dest, update.Data, false)
		}
		bf.WriteUint8(1)
		bf.WriteUint8(update.Dest)
		bf.WriteUint32(_old)
		bf.WriteUint32(_new)
	}
	s.server.raviente.Unlock()
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())

	if s.server.erupeConfig.GameplayOptions.LowLatencyRaviente {
		s.notifyRavi()
	}
}

func handleMsgSysLoadRegister(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLoadRegister)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	bf.WriteUint8(pkt.Values)
	for i := uint8(0); i < pkt.Values; i++ {
		switch pkt.RegisterID {
		case 4:
			bf.WriteUint32(s.server.raviente.state[i])
		case 5:
			bf.WriteUint32(s.server.raviente.support[i])
		case 6:
			bf.WriteUint32(s.server.raviente.register[i])
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func (s *Session) notifyRavi() {
	sema := getRaviSemaphore(s.server)
	if sema == nil {
		return
	}
	var temp mhfpacket.MHFPacket
	raviNotif := byteframe.NewByteFrame()
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: 4}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: 5}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	temp = &mhfpacket.MsgSysNotifyRegister{RegisterID: 6}
	raviNotif.WriteUint16(uint16(temp.Opcode()))
	temp.Build(raviNotif, s.clientContext)
	raviNotif.WriteUint16(0x0010) // End it.
	if s.server.erupeConfig.GameplayOptions.LowLatencyRaviente {
		for session := range sema.clients {
			session.QueueSend(raviNotif.Data())
		}
	} else {
		for session := range sema.clients {
			if session.charID == s.charID {
				session.QueueSend(raviNotif.Data())
			}
		}
	}
}

func getRaviSemaphore(s *Server) *Semaphore {
	for _, semaphore := range s.semaphore {
		if strings.HasPrefix(semaphore.id_semaphore, "hs_l0u3B5") && strings.HasSuffix(semaphore.id_semaphore, "3") {
			return semaphore
		}
	}
	return nil
}

func resetRavi(s *Session) {
	s.server.raviente = NewRaviente()
}

func handleMsgSysNotifyRegister(s *Session, p mhfpacket.MHFPacket) {}
