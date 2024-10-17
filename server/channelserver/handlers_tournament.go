package channelserver

import (
	"erupe-ce/internal/model"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/gametime"
	ps "erupe-ce/utils/pascalstring"

	"github.com/jmoiron/sqlx"
)

func handleMsgMhfInfoTournament(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoTournament)
	bf := byteframe.NewByteFrame()

	tournamentInfo0 := []model.TournamentInfo0{}
	tournamentInfo21 := []model.TournamentInfo21{}
	tournamentInfo22 := []model.TournamentInfo22{}

	switch pkt.Unk0 {
	case 0:
		bf.WriteUint32(0)
		bf.WriteUint32(uint32(len(tournamentInfo0)))
		for _, tinfo := range tournamentInfo0 {
			bf.WriteUint32(tinfo.ID)
			bf.WriteUint32(tinfo.MaxPlayers)
			bf.WriteUint32(tinfo.CurrentPlayers)
			bf.WriteUint16(tinfo.Unk1)
			bf.WriteUint16(tinfo.TextColor)
			bf.WriteUint32(tinfo.Unk2)
			bf.WriteUint32(uint32(tinfo.Time1.Unix()))
			bf.WriteUint32(uint32(tinfo.Time2.Unix()))
			bf.WriteUint32(uint32(tinfo.Time3.Unix()))
			bf.WriteUint32(uint32(tinfo.Time4.Unix()))
			bf.WriteUint32(uint32(tinfo.Time5.Unix()))
			bf.WriteUint32(uint32(tinfo.Time6.Unix()))
			bf.WriteUint8(tinfo.Unk3)
			bf.WriteUint8(tinfo.Unk4)
			bf.WriteUint32(tinfo.MinHR)
			bf.WriteUint32(tinfo.MaxHR)
			ps.Uint8(bf, tinfo.Unk5, true)
			ps.Uint16(bf, tinfo.Unk6, true)
		}
	case 1:
		bf.WriteUint32(uint32(gametime.TimeAdjusted().Unix()))
		bf.WriteUint32(0) // Registered ID
		bf.WriteUint32(0)
		bf.WriteUint32(0)
		bf.WriteUint8(0)
		bf.WriteUint32(0)
		ps.Uint8(bf, "", true)
	case 2:
		bf.WriteUint32(0)
		bf.WriteUint32(uint32(len(tournamentInfo21)))
		for _, info := range tournamentInfo21 {
			bf.WriteUint32(info.Unk0)
			bf.WriteUint32(info.Unk1)
			bf.WriteUint32(info.Unk2)
			bf.WriteUint8(info.Unk3)
		}
		bf.WriteUint32(uint32(len(tournamentInfo22)))
		for _, info := range tournamentInfo22 {
			bf.WriteUint32(info.Unk0)
			bf.WriteUint32(info.Unk1)
			bf.WriteUint32(info.Unk2)
			bf.WriteUint8(info.Unk3)
			ps.Uint8(bf, info.Unk4, true)
		}
	}

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfEntryTournament(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryTournament)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireTournament(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireTournament)
	rewards := []model.TournamentReward{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(rewards)))
	for _, reward := range rewards {
		bf.WriteUint16(reward.Unk0)
		bf.WriteUint16(reward.Unk1)
		bf.WriteUint16(reward.Unk2)
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}
