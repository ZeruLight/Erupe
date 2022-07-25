package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgMhfUpdateInterior(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateInterior)
	_, err := s.server.db.Exec("UPDATE characters SET house=$1 WHERE id=$2", pkt.InteriorData, s.charID)
	if err != nil {
		panic(err)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

type HouseData struct {
	CharID uint32 `db:"id"`
	HRP    uint16 `db:"hrp"`
	GR     uint16 `db:"gr"`
	Name   string `db:"name"`
}

func handleMsgMhfEnumerateHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateHouse)
	bf := byteframe.NewByteFrame()
	var houses []HouseData
	switch pkt.Method {
	case 1:
		var friendsList string
		s.server.db.QueryRow("SELECT friends FROM characters WHERE id=$1", s.charID).Scan(&friendsList)
		cids := stringsupport.CSVElems(friendsList)
		for _, cid := range cids {
			house := HouseData{}
			row := s.server.db.QueryRowx("SELECT id, hrp, gr, name FROM characters WHERE id=$1", cid)
			err := row.StructScan(&house)
			if err != nil {
				panic(err)
			} else {
				houses = append(houses, house)
			}
		}
	case 2:
		guild, err := GetGuildInfoByCharacterId(s, s.charID)
		if err != nil {
			break
		}
		guildMembers, err := GetGuildMembers(s, guild.ID, false)
		if err != nil {
			break
		}
		for _, member := range guildMembers {
			house := HouseData{}
			row := s.server.db.QueryRowx("SELECT id, hrp, gr, name FROM characters WHERE id=$1", member.CharID)
			err := row.StructScan(&house)
			if err != nil {
				panic(err)
			} else {
				houses = append(houses, house)
			}
		}
	case 3:
		house := HouseData{}
		row := s.server.db.QueryRowx("SELECT id, hrp, gr, name FROM characters WHERE name=$1", pkt.Name)
		err := row.StructScan(&house)
		if err != nil {
			panic(err)
		} else {
			houses = append(houses, house)
		}
	case 4:
		house := HouseData{}
		row := s.server.db.QueryRowx("SELECT id, hrp, gr, name FROM characters WHERE id=$1", pkt.CharID)
		err := row.StructScan(&house)
		if err != nil {
			panic(err)
		} else {
			houses = append(houses, house)
		}
	case 5: // Recent visitors
		break
	}
	var exists int
	for _, house := range houses {
		for _, session := range s.server.sessions {
			if session.charID == house.CharID {
				exists++
				bf.WriteUint32(house.CharID)
				bf.WriteUint8(session.house.state)
				if len(session.house.password) > 0 {
					bf.WriteUint8(3)
				} else {
					bf.WriteUint8(0)
				}
				bf.WriteUint16(house.HRP)
				bf.WriteUint16(house.GR)
				ps.Uint8(bf, house.Name, true)
				break
			}
		}
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint16(uint16(exists))
	resp.WriteBytes(bf.Data())
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateHouse)
	// 01 = closed
	// 02 = open anyone
	// 03 = open friends
	// 04 = open guild
	// 05 = open friends+guild
	s.house.state = pkt.State
	s.house.password = pkt.Password
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHouse)
	bf := byteframe.NewByteFrame()
	if pkt.Destination != 9 && len(pkt.Password) > 0 && pkt.CheckPass {
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID && pkt.Password != session.house.password {
				doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
				return
			}
		}
	}

	var furniture []byte
	err := s.server.db.QueryRow("SELECT house FROM characters WHERE id=$1", pkt.CharID).Scan(&furniture)
	if err != nil {
		panic(err)
	}
	if furniture == nil {
		furniture = make([]byte, 20)
	}

	// TODO: Find where the missing data comes from, savefile offset?
	switch pkt.Destination {
	case 3: // Others house
		houseTier := uint8(2) // Fallback if can't find
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID {
				houseTier = session.house.tier
			}
		}
		bf.WriteBytes(make([]byte, 4))
		bf.WriteUint8(houseTier) // House tier 0x1FB70
		bf.WriteBytes(make([]byte, 80))
		// Item box style bitfield
		// tier 1 = 0x09FE
		// tier 2 = 0x69FF
		// tier 3 = 0xE9FF
		// unused = 0x0001
		// unused = 0x6000
		switch houseTier {
		case 0:
			bf.WriteUint16(0x0000)
		case 1:
			bf.WriteUint16(0x69FF)
		case 2:
			bf.WriteUint16(0xE9FF)
		}
		// Rastae
		// Partner
		bf.WriteBytes(make([]byte, 132))
		bf.WriteBytes(furniture)
	case 4: // Bookshelf
		// Hunting log
		bf.WriteBytes(make([]byte, 5576))
	case 5: // Gallery
		// Furniture placement
		bf.WriteBytes(make([]byte, 1748))
	case 8: // Tore
		// Sister
		// Cat shops
		// Pugis
		bf.WriteBytes(make([]byte, 240))
	case 9: // Own house
		bf.WriteBytes(furniture)
	case 10: // Garden
		// Gardening upgrades
		// Gooks
		bf.WriteBytes(make([]byte, 72))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetMyhouseInfo)

	var data []byte
	err := s.server.db.QueryRow("SELECT trophy FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		panic(err)
	}
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgMhfUpdateMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateMyhouseInfo)

	_, err := s.server.db.Exec("UPDATE characters SET trophy=$1 WHERE id=$2", pkt.Unk0, s.charID)
	if err != nil {
		panic(err)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadDecoMyset)
	var data []byte
	err := s.server.db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get preset decorations savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
		//doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		// set first byte to 1 to avoid pop up every time without save
		body := make([]byte, 0x226)
		body[0] = 1
		doAckBufSucceed(s, pkt.AckHandle, body)
	}
}

func handleMsgMhfSaveDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveDecoMyset)
	// https://gist.github.com/Andoryuuta/9c524da7285e4b5ca7e52e0fc1ca1daf
	var loadData []byte
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload[1:]) // skip first unk byte
	err := s.server.db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.charID).Scan(&loadData)
	if err != nil {
		s.logger.Fatal("Failed to get preset decorations savedata from db", zap.Error(err))
	} else {
		numSets := bf.ReadUint8() // sets being written
		// empty save
		if len(loadData) == 0 {
			loadData = []byte{0x01, 0x00}
		}

		savedSets := loadData[1] // existing saved sets
		// no sets, new slice with just first 2 bytes for appends later
		if savedSets == 0 {
			loadData = []byte{0x01, 0x00}
		}
		for i := 0; i < int(numSets); i++ {
			writeSet := bf.ReadUint16()
			dataChunk := bf.ReadBytes(76)
			setBytes := append([]byte{uint8(writeSet >> 8), uint8(writeSet & 0xff)}, dataChunk...)
			for x := 0; true; x++ {
				if x == int(savedSets) {
					// appending set
					if loadData[len(loadData)-1] == 0x10 {
						// sanity check for if there was a messy manual import
						loadData = append(loadData[:len(loadData)-2], setBytes...)
					} else {
						loadData = append(loadData, setBytes...)
					}
					savedSets++
					break
				}
				currentSet := loadData[3+(x*78)]
				if int(currentSet) == int(writeSet) {
					// replacing a set
					loadData = append(loadData[:2+(x*78)], append(setBytes, loadData[2+((x+1)*78):]...)...)
					break
				} else if int(currentSet) > int(writeSet) {
					// inserting before current set
					loadData = append(loadData[:2+((x)*78)], append(setBytes, loadData[2+((x)*78):]...)...)
					savedSets++
					break
				}
			}
			loadData[1] = savedSets // update set count
		}
		_, err := s.server.db.Exec("UPDATE characters SET decomyset=$1 WHERE id=$2", loadData, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update decomyset savedata in db", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEnumerateTitle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateTitle)
	bf := byteframe.NewByteFrame()
	if pkt.CharID == s.charID {
		titleCount := 114                  // all titles unlocked
		bf.WriteUint16(uint16(titleCount)) // title count
		bf.WriteUint16(0)                  // unk
		for i := 0; i < titleCount; i++ {
			bf.WriteUint16(uint16(i))
			bf.WriteUint16(0) // unk
			bf.WriteUint32(0) // timestamp acquired
			bf.WriteUint32(0) // timestamp updated
		}
	} else {
		bf.WriteUint16(0)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfOperateWarehouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateWarehouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateWarehouse(s *Session, p mhfpacket.MHFPacket) {}
