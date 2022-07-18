package channelserver

import (
	"time"
	"encoding/hex"
	"math/rand"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/common/byteframe"
)

func handleMsgMhfSaveMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMezfesData)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadMezfesData)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Unk

	resp.WriteUint8(2) // Count of the next 2 uint32s
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	resp.WriteUint32(0) // Unk

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRanking)
	bf := byteframe.NewByteFrame()
	state := s.server.erupeConfig.DevModeOptions.TournamentEvent
	// Unk
	// Unk
	// Start?
	// End?
	midnight := Time_Current_Midnight()
	switch state {
	case 1:
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(3 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(12 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(21 * 24 * time.Hour).Unix()))
	case 2:
		bf.WriteUint32(uint32(midnight.Add(-3 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(9 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(16 * 24 * time.Hour).Unix()))
	case 3:
		bf.WriteUint32(uint32(midnight.Add(-12 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(-9 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(7 * 24 * time.Hour).Unix()))
	default:
		bf.WriteBytes(make([]byte, 16))
		bf.WriteUint32(uint32(Time_Current_Adjusted().Unix())) // TS Current Time
		bf.WriteUint16(1)
		bf.WriteUint32(0)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}
	bf.WriteUint32(uint32(Time_Current_Adjusted().Unix())) // TS Current Time
	d, _ := hex.DecodeString("031491E631353089F18CF68EAE8EEB97C291E589EF00001200000A54001000000000ED130D949A96B697B393A294B081490000000A55001000010000ED130D949A96B697B393A294B081490000000A56001000020000ED130D949A96B697B393A294B081490000000A57001000030000ED130D949A96B697B393A294B081490000000A58001000040000ED130D949A96B697B393A294B081490000000A59001000050000ED130D949A96B697B393A294B081490000000A5A001000060000ED130D949A96B697B393A294B081490000000A5B001000070000ED130D949A96B697B393A294B081490000000A5C001000080000ED130D949A96B697B393A294B081490000000A5D001000090000ED130D949A96B697B393A294B081490000000A5E0010000A0000ED130D949A96B697B393A294B081490000000A5F0010000B0000ED130D949A96B697B393A294B081490000000A600010000C0000ED130D949A96B697B393A294B081490000000A610010000D0000ED130D949A96B697B393A294B081490000000A620011FFFF0000ED121582DD82F182C882C5949A96B697B393A294B081490000000A63000600EA0000000009834C838C834183570000000A64000600ED000000000B836E838A837D834F838D0000000A65000600EF0000000011834A834E8354839383668381834C83930003000002390006000600000E8CC2906C208B9091E58B9B94740001617E43303581798BA38B5A93E09765817A0A7E433030834E83478358836782C592DE82C182BD8B9B82CC83548343835982F08BA382A40A7E433034817991CE8FDB8B9B817A0A7E433030834C838C8341835781410A836E838A837D834F838D8141834A834E8354839383668381834C83930A7E433037817993FC8FDC8FDC9569817A0A7E4330308B9B947482CC82B582E982B58141835E838B836C835290B68E598C9481410A834F815B834E90B68E598C948141834F815B834E91AB90B68E598C9481410A834F815B834E89F095FA8C94283181603388CA290A2F97C29263837C8343839383672831816031303088CA290A2F8FA08360835083628367817B836E815B8374836083508362836794920A2831816035303088CA290A7E43303381798A4A8DC38AFA8AD4817A0A7E43303032303139944E31318C8E323293FA2031343A303082A982E70A32303139944E31318C8E323593FA2031343A303082DC82C5000000023A0011000700001297C292632082668B89E8E891CA935694740000ED7E43303581798BA38B5A93E09765817A0A7E43303081E182DD82F182C882C5949A96B697B393A294B0814981E282F00A93AF82B697C2926382C98F8A91AE82B782E934906C82DC82C582CC0A97C2926388F582C582A282A982C9918182AD834E838A834182B782E982A90A82F08BA382A40A0A7E433037817993FC8FDC8FDC9569817A0A7E43303091E631343789F18EEB906C8DD582CC8DB02831816032303088CA290A0A7E43303381798A4A8DC38AFA8AD4817A0A7E43303032303139944E31318C8E323293FA2031343A303082A982E70A32303139944E31318C8E323593FA2031343A303082DC82C50A000000023B001000070000128CC2906C2082668B89E8E891CA935694740001497E43303581798BA38B5A93E09765817A0A7E43303081E1949A96B697B393A294B0814981E282F00A82A282A982C9918182AD834E838A834182B782E982A982F08BA382A40A0A7E433037817993FC8FDC8FDC9569817A0A7E43303089A48ED282CC8381835F838B283188CA290A2F8CF68EAE82CC82B582E982B58141835E838B836C835290B68E598C9481410A834F815B834E90B68E598C948141834F815B834E91AB90B68E598C9481410A834F815B834E89F095FA8C94283181603388CA290A2F97C29263837C8343839383672831816031303088CA290A2F8FA08360835083628367817B836E815B8374836083508362836794920A2831816035303088CA290A7E43303381798A4A8DC38AFA8AD4817A0A7E43303032303139944E31318C8E323293FA2031343A303082A982E70A32303139944E31318C8E323593FA2031343A303082DC82C500")
	bf.WriteBytes(d)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfInfoFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoFesta)
	bf := byteframe.NewByteFrame()
	state := s.server.erupeConfig.DevModeOptions.FestaEvent
	bf.WriteUint32(0xdeadbeef) // festaID
	// Registration Week Start
	// Introductory Week Start
	// Totalling Time
	// Reward Festival Start (2.5hrs after totalling)
	// 2 weeks after RewardFes (next fes?)
	midnight := Time_Current_Midnight()
	switch state {
	case 1:
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 7 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 14 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 14 * time.Hour + 150 * time.Minute).Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 28 * time.Hour + 11 * time.Hour).Unix()))
	case 2:
		bf.WriteUint32(uint32(midnight.Add(-24 * 7 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 7 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 7 * time.Hour + 150 * time.Minute).Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 21 * time.Hour).Add(11 * time.Hour).Unix()))
	case 3:
		bf.WriteUint32(uint32(midnight.Add(-24 * 14 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(-24 * 7 * time.Hour + 11 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(150 * time.Minute).Unix()))
		bf.WriteUint32(uint32(midnight.Add(24 * 14 * time.Hour + 11 * time.Hour).Unix()))
	default:
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf.WriteUint32(uint32(Time_Current_Adjusted().Unix())) // TS Current Time
	bf.WriteUint8(4)
	bf.WriteUint8(2)
	bf.WriteBytes([]byte{0x61, 0x00}) // uint8pascal
	bf.WriteUint32(0)
	bf.WriteUint32(0) // Blue souls
	bf.WriteUint32(0) // Red souls

	trials := 0
	bf.WriteUint16(uint16(trials))
	for i := 0; i < trials; i++ {
		bf.WriteUint32(uint32(i+1)) // trialID
		bf.WriteUint8(0xFF) // unk
		bf.WriteUint8(uint8(i)) // objective
		bf.WriteUint32(0x1B) // monID, itemID if deliver
		bf.WriteUint16(1) // huntsRemain?
		bf.WriteUint16(0) // location
		bf.WriteUint16(1) // numSoulsReward
		bf.WriteUint8(0xFF) // unk
		bf.WriteUint8(0xFF) // monopolised
		bf.WriteUint16(0) // unk
	}

	d, _ := hex.DecodeString("0000001901000007015E05F000000000000100000703E81B6300000000010100000C03E8000000000000000100000D0000000000000000000100000100000000000000000002000007015E05F000000000000200000703E81B6300000000010200000C03E8000000000000000200000D0000000000000000000200000400000000000000000003000007015E05F000000000000300000703E81B6300000000010300000C03E8000000000000000300000D0000000000000000000300000100000000000000000004000007015E05F000000000000400000703E81B6300000000010400000C03E8000000000000000400000D0000000000000000000400000400000000000000000005000007015E05F000000000000500000703E81B6300000000010500000C03E8000000000000000500000D000000000000000000050000010000000000000000000001D4C001F4000000000000000100001388000007D0000003E800000064012C00C8009600640032")
	bf.WriteBytes(d)

	bf.WriteUint16(2)
	bf.WriteBytes([]byte{0x61, 0x00}) // uint16pascal

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (U)ser
func handleMsgMhfStateFestaU(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaU)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // souls
	bf.WriteUint32(0) // unk
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (G)uild
func handleMsgMhfStateFestaG(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaG)
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // souls
	resp.WriteUint32(1) // unk
	resp.WriteUint32(1) // unk
	resp.WriteUint32(1) // unk, rank?
	resp.WriteUint32(1) // unk
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateFestaMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaMember)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // numMembers
	// uint16 unk
	// uint32 charID
	// uint32 souls
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfVoteFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryFesta)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEntryFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryFesta)
	bf := byteframe.NewByteFrame()
	rand.Seed(time.Now().UnixNano())
	bf.WriteUint32(uint32(rand.Intn(2)))
	// Update guild table
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfChargeFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfChargeFesta)
	// Update festa state table
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFesta)
	// Mark festa as claimed
	// Update guild table?
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFestaPersonalPrize)
	// Set prize as claimed
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFestaIntermediatePrize)
	// Set prize as claimed
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// uint32 numPrizes
// struct festaPrize
// uint32 prizeID
// uint32 prizeTier (1/2/3, 3 = GR)
// uint32 soulsReq
// uint32 unk (00 00 00 07)
// uint32 itemID
// uint32 numItem
// bool claimed

func handleMsgMhfEnumerateFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaPersonalPrize)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaIntermediatePrize)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}
