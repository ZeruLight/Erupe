package channelserver

import (
	"fmt"

	"erupe-ce/network/mhfpacket"

	"github.com/jmoiron/sqlx"
)

func handleMsgSysInsertUser(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteUser(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysSetUserBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetUserBinary)
	s.Server.userBinaryPartsLock.Lock()
	s.Server.userBinaryParts[userBinaryPartID{charID: s.CharID, index: pkt.BinaryType}] = pkt.RawDataPayload
	s.Server.userBinaryPartsLock.Unlock()

	var exists []byte
	err := db.QueryRow("SELECT type2 FROM user_binary WHERE id=$1", s.CharID).Scan(&exists)
	if err != nil {
		db.Exec("INSERT INTO user_binary (id) VALUES ($1)", s.CharID)
	}

	db.Exec(fmt.Sprintf("UPDATE user_binary SET type%d=$1 WHERE id=$2", pkt.BinaryType), pkt.RawDataPayload, s.CharID)

	msg := &mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.CharID,
		BinaryType: pkt.BinaryType,
	}

	s.Server.BroadcastMHF(msg, s)
}

func handleMsgSysGetUserBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetUserBinary)

	// Try to get the data.
	s.Server.userBinaryPartsLock.RLock()
	defer s.Server.userBinaryPartsLock.RUnlock()
	data, ok := s.Server.userBinaryParts[userBinaryPartID{charID: pkt.CharID, index: pkt.BinaryType}]

	// If we can't get the real data, try to get it from the database.
	if !ok {
		err := db.QueryRow(fmt.Sprintf("SELECT type%d FROM user_binary WHERE id=$1", pkt.BinaryType), pkt.CharID).Scan(&data)
		if err != nil {
			s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		} else {
			s.DoAckBufSucceed(pkt.AckHandle, data)
		}
	} else {
		s.DoAckBufSucceed(pkt.AckHandle, data)
	}
}

func handleMsgSysNotifyUserBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}
