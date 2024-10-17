package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/stringsupport"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func handleMsgSysEnumerateClient(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateClient)

	s.Server.stagesLock.RLock()
	stage, ok := s.Server.stages[pkt.StageID]
	if !ok {
		s.Server.stagesLock.RUnlock()
		s.Logger.Warn("Can't enumerate clients for stage that doesn't exist!", zap.String("stageID", pkt.StageID))
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}
	s.Server.stagesLock.RUnlock()

	// Read-lock the stage and make the response with all of the charID's in the stage.
	resp := byteframe.NewByteFrame()
	stage.RLock()
	var clients []uint32
	switch pkt.Get {
	case 0: // All
		for _, cid := range stage.Clients {
			clients = append(clients, cid)
		}
		for cid := range stage.ReservedClientSlots {
			clients = append(clients, cid)
		}
	case 1: // Not ready
		for cid, ready := range stage.ReservedClientSlots {
			if !ready {
				clients = append(clients, cid)
			}
		}
	case 2: // Ready
		for cid, ready := range stage.ReservedClientSlots {
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

	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
	s.Logger.Debug("MsgSysEnumerateClient Done!")
}

func handleMsgMhfListMember(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfListMember)

	var csv string
	var count uint32
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Blacklist count
	err := db.QueryRow("SELECT blocked FROM characters WHERE id=$1", s.CharID).Scan(&csv)
	if err == nil {
		cids := stringsupport.CSVElems(csv)
		for _, cid := range cids {
			var name string
			err = db.QueryRow("SELECT name FROM characters WHERE id=$1", cid).Scan(&name)
			if err != nil {
				continue
			}
			count++
			resp.WriteUint32(uint32(cid))
			resp.WriteUint32(16)
			resp.WriteBytes(stringsupport.PaddedString(name, 16, true))
		}
	}
	resp.Seek(0, 0)
	resp.WriteUint32(count)
	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
}

func handleMsgMhfOprMember(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOprMember)

	var csv string
	for _, cid := range pkt.CharIDs {
		if pkt.Blacklist {
			err := db.QueryRow("SELECT blocked FROM characters WHERE id=$1", s.CharID).Scan(&csv)
			if err == nil {
				if pkt.Operation {
					csv = stringsupport.CSVRemove(csv, int(cid))
				} else {
					csv = stringsupport.CSVAdd(csv, int(cid))
				}
				db.Exec("UPDATE characters SET blocked=$1 WHERE id=$2", csv, s.CharID)
			}
		} else { // Friendlist
			err := db.QueryRow("SELECT friends FROM characters WHERE id=$1", s.CharID).Scan(&csv)
			if err == nil {
				if pkt.Operation {
					csv = stringsupport.CSVRemove(csv, int(cid))
				} else {
					csv = stringsupport.CSVAdd(csv, int(cid))
				}
				db.Exec("UPDATE characters SET friends=$1 WHERE id=$2", csv, s.CharID)
			}
		}
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfShutClient(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysHideClient(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}
