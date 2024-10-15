package channelserver

import (
	"erupe-ce/network/mhfpacket"

	"github.com/jmoiron/sqlx"
)

func handleMsgSysReserve188(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserve188)

	// Left as raw bytes because I couldn't easily find the request or resp parser function in the binary.
	s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysReserve18B(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserve18B)

	// Left as raw bytes because I couldn't easily find the request or resp parser function in the binary.
	s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x3C})
}

func handleMsgSysReserve55(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve56(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve57(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve01(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve02(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve03(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve04(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve05(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve06(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve07(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve0C(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve0D(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve0E(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4A(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4B(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4C(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4D(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4E(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4F(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve5C(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve5E(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve5F(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve71(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve72(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve73(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve74(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve75(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve76(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve77(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve78(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve79(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7A(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7B(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7C(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7E(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgMhfReserve10F(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve180(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve18E(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve18F(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve19E(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve19F(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1A4(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1A6(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1A7(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1A8(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1A9(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1AA(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1AB(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1AC(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1AD(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1AE(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve1AF(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve19B(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve192(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve193(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve194(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}
