package channelserver

import (
	"erupe-ce/network/mhfpacket"

	"github.com/jmoiron/sqlx"
)

func handleMsgSysCreateMutex(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysCreateOpenMutex(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteMutex(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysOpenMutex(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysCloseMutex(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}
