package channelserver

import (
	"fmt"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/broadcast"
	"erupe-ce/utils/db"
)

func handleMsgSysInsertUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetUserBinary)
	s.Server.userBinaryPartsLock.Lock()
	s.Server.userBinaryParts[userBinaryPartID{charID: s.CharID, index: pkt.BinaryType}] = pkt.RawDataPayload
	s.Server.userBinaryPartsLock.Unlock()
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var exists []byte
	err = database.QueryRow("SELECT type2 FROM user_binary WHERE id=$1", s.CharID).Scan(&exists)
	if err != nil {
		database.Exec("INSERT INTO user_binary (id) VALUES ($1)", s.CharID)
	}

	database.Exec(fmt.Sprintf("UPDATE user_binary SET type%d=$1 WHERE id=$2", pkt.BinaryType), pkt.RawDataPayload, s.CharID)

	msg := &mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.CharID,
		BinaryType: pkt.BinaryType,
	}

	s.Server.BroadcastMHF(msg, s)
}

func handleMsgSysGetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetUserBinary)

	// Try to get the data.
	s.Server.userBinaryPartsLock.RLock()
	defer s.Server.userBinaryPartsLock.RUnlock()
	data, ok := s.Server.userBinaryParts[userBinaryPartID{charID: pkt.CharID, index: pkt.BinaryType}]
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	// If we can't get the real data, try to get it from the database.
	if !ok {
		err = database.QueryRow(fmt.Sprintf("SELECT type%d FROM user_binary WHERE id=$1", pkt.BinaryType), pkt.CharID).Scan(&data)
		if err != nil {
			broadcast.DoAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		} else {
			broadcast.DoAckBufSucceed(s, pkt.AckHandle, data)
		}
	} else {
		broadcast.DoAckBufSucceed(s, pkt.AckHandle, data)
	}
}

func handleMsgSysNotifyUserBinary(s *Session, p mhfpacket.MHFPacket) {}
