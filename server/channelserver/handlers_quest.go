package channelserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func handleMsgSysGetFile(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetFile)

	// Debug print the request.
	if pkt.IsScenario {
		fmt.Printf("%+v\n", pkt.ScenarioIdentifer)
		filename := fmt.Sprintf("%d_0_0_0_S%d_T%d_C%d", pkt.ScenarioIdentifer.CategoryID, pkt.ScenarioIdentifer.MainID, pkt.ScenarioIdentifer.Flags, pkt.ScenarioIdentifer.ChapterID)
		// Read the scenario file.
		data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("scenarios/%s.bin", filename)))
		if err != nil {
			panic(err)
		}
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		if _, err := os.Stat(filepath.Join(s.server.erupeConfig.BinPath, "quest_override.bin")); err == nil {
			data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "quest_override.bin"))
			if err != nil {
				panic(err)
			}
			doAckBufSucceed(s, pkt.AckHandle, data)
		} else {
			// Get quest file.
			data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("quests/%s.bin", pkt.Filename)))
			if err != nil {
				s.logger.Fatal(fmt.Sprintf("Failed to open quest file: quests/%s.bin", pkt.Filename))
			}
			doAckBufSucceed(s, pkt.AckHandle, data)
		}
	}
}

func handleMsgMhfLoadFavoriteQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadFavoriteQuest)
	var data []byte
	err := s.server.db.QueryRow("SELECT savefavoritequest FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err == nil && len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgMhfSaveFavoriteQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveFavoriteQuest)
	dumpSaveData(s, pkt.Data, "favquest")
	s.server.db.Exec("UPDATE characters SET savefavoritequest=$1 WHERE id=$2", pkt.Data, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEnumerateQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateQuest)
	var totalCount, returnedCount uint16
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	err := filepath.Walk(fmt.Sprintf("%s/events/", s.server.erupeConfig.BinPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		} else {
			if len(data) > 850 || len(data) < 400 {
				return nil // Could be more or less strict with size limits
			} else {
				totalCount++
				if totalCount > pkt.Offset && len(bf.Data()) < 64000 {
					returnedCount++
					bf.WriteBytes(data)
					return nil
				}
			}
		}
		return nil
	})
	if err != nil || totalCount == 0 {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 18))
		return
	}
	bf.WriteUint16(0) // Unk
	bf.WriteUint16(0) // Unk
	bf.WriteUint16(0) // Unk
	bf.WriteUint32(0) // Unk
	bf.WriteUint16(0) // Unk
	bf.WriteUint16(totalCount)
	bf.WriteUint16(pkt.Offset)
	bf.Seek(0, io.SeekStart)
	bf.WriteUint16(returnedCount)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnterTournamentQuest(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdBonusQuestInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdBonusQuestInfo)

	udBonusQuestInfos := []struct {
		Unk0      uint8
		Unk1      uint8
		StartTime uint32 // Unix timestamp (seconds)
		EndTime   uint32 // Unix timestamp (seconds)
		Unk4      uint32
		Unk5      uint8
		Unk6      uint8
	}{} // Blank stub array.

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(udBonusQuestInfos)))
	for _, q := range udBonusQuestInfos {
		resp.WriteUint8(q.Unk0)
		resp.WriteUint8(q.Unk1)
		resp.WriteUint32(q.StartTime)
		resp.WriteUint32(q.EndTime)
		resp.WriteUint32(q.Unk4)
		resp.WriteUint8(q.Unk5)
		resp.WriteUint8(q.Unk6)
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}
