package channelserver

import (
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/common/byteframe"
	"go.uber.org/zap"
)

func handleMsgSysEnumerateClient(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateClient)

	s.server.stagesLock.RLock()
	stage, ok := s.server.stages[pkt.StageID]
	if !ok {
		s.logger.Fatal("Can't enumerate clients for stage that doesn't exist!", zap.String("stageID", pkt.StageID))
	}
	s.server.stagesLock.RUnlock()

	// Read-lock the stage and make the response with all of the charID's in the stage.
	resp := byteframe.NewByteFrame()
	stage.RLock()
	var clients []uint32
	switch pkt.Unk1 {
	case 1: // Not ready
		for cid, ready := range stage.reservedClientSlots {
			if !ready {
				clients = append(clients, cid)
			}
		}
	case 2: // Ready
	for cid, ready := range stage.reservedClientSlots {
		if ready {
			clients = append(clients, cid)
		}
	}
	}
	resp.WriteUint16(uint16(len(clients)))
	for _, cid := range clients {
		resp.WriteUint32(cid)
	}
	stage.RUnlock()

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	s.logger.Debug("MsgSysEnumerateClient Done!")
}

func handleMsgMhfListMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfListMember)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Members count. (Unsure of what kind of members these actually are, guild, party, COG subscribers, etc.)

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfOprMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOprMember)
	var csv string
	if pkt.Blacklist {
		if pkt.Operation {
			// remove from blacklist
		} else {
			// add to blacklist
		}
	} else { // Friendlist
		err := s.server.db.QueryRow("SELECT friends FROM characters WHERE id=$1", s.charID).Scan(&csv)
		if err != nil {
			panic(err)
		}
		if pkt.Operation {
			csv = stringsupport.CSVRemove(csv, int(pkt.CharID))
		} else {
			csv = stringsupport.CSVAdd(csv, int(pkt.CharID))
		}
		_, _ = s.server.db.Exec("UPDATE characters SET friends=$1 WHERE id=$2", csv, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfShutClient(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysHideClient(s *Session, p mhfpacket.MHFPacket) {}
