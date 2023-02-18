package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"
	"time"
)

func handleMsgMhfInfoTournament(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoTournament)
	bf := byteframe.NewByteFrame()

	switch pkt.Unk0 {
	case 0:
		bf.WriteUint32(uint32(Time_Current_Adjusted().Unix()))
		bf.WriteUint32(0) // Tied to schedule ID?
	case 1:

		bf.WriteBytes(make([]byte, 21))
		ps.Uint8(bf, "", false)
		break

		bf.WriteUint32(0xACEDCAFE)

		bf.WriteUint32(5) // Active schedule?

		bf.WriteUint32(1) // Schedule ID?

		bf.WriteUint32(32) // Max players
		bf.WriteUint32(0)  // Registered players

		bf.WriteUint16(0)
		bf.WriteUint16(2) // Color code for schedule item
		bf.WriteUint32(0)

		bf.WriteUint32(uint32(time.Now().Add(time.Hour * -10).Unix()))
		bf.WriteUint32(uint32(time.Now().Add(time.Hour * 10).Unix()))
		bf.WriteUint32(uint32(time.Now().Add(time.Hour * 10).Unix()))
		bf.WriteUint32(uint32(time.Now().Add(time.Hour * 10).Unix()))
		bf.WriteUint32(uint32(time.Now().Add(time.Hour * 10).Unix()))
		bf.WriteUint32(uint32(time.Now().Add(time.Hour * 10).Unix()))

		bf.WriteBool(true)  // Unk
		bf.WriteBool(false) // Cafe-only

		bf.WriteUint32(0) // Min HR
		bf.WriteUint32(0) // Max HR

		ps.Uint8(bf, "Test", false)

		// ...
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEntryTournament(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireTournament(s *Session, p mhfpacket.MHFPacket) {}
