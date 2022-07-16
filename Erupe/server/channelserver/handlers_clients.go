package channelserver

import (
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/common/byteframe"
	"go.uber.org/zap"
)

func handleMsgSysEnumerateClient(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateClient)

	// Read-lock the stages map.
	s.server.stagesLock.RLock()

	stage, ok := s.server.stages[pkt.StageID]
	if !ok {
		s.logger.Fatal("Can't enumerate clients for stage that doesn't exist!", zap.String("stageID", pkt.StageID))
	}

	// Unlock the stages map.
	s.server.stagesLock.RUnlock()

	// Read-lock the stage and make the response with all of the charID's in the stage.
	resp := byteframe.NewByteFrame()
	stage.RLock()

	// TODO(Andoryuuta): Is only the reservations needed? Do clients send this packet for mezeporta as well?

	// Make a map to deduplicate the charIDs between the unreserved clients and the reservations.
	deduped := make(map[uint32]interface{})

	// Add the charIDs
	for session := range stage.clients {
		deduped[session.charID] = nil
	}

	for charid := range stage.reservedClientSlots {
		deduped[charid] = nil
	}

	// Write the deduplicated response
	resp.WriteUint16(uint16(len(deduped))) // Client count
	for charid := range deduped {
		resp.WriteUint32(charid)
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
