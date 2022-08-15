package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"go.uber.org/zap"
	"io"
	"time"
)

const warehouseNamesQuery = `
SELECT
COALESCE(item0name, ''),
COALESCE(item1name, ''),
COALESCE(item2name, ''),
COALESCE(item3name, ''),
COALESCE(item4name, ''),
COALESCE(item5name, ''),
COALESCE(item6name, ''),
COALESCE(item7name, ''),
COALESCE(item8name, ''),
COALESCE(item9name, ''),
COALESCE(equip0name, ''),
COALESCE(equip1name, ''),
COALESCE(equip2name, ''),
COALESCE(equip3name, ''),
COALESCE(equip4name, ''),
COALESCE(equip5name, ''),
COALESCE(equip6name, ''),
COALESCE(equip7name, ''),
COALESCE(equip8name, ''),
COALESCE(equip9name, '')
FROM warehouse
`

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
				bf.WriteUint8(session.myseries.state)
				if len(session.myseries.password) > 0 {
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
	s.myseries.state = pkt.State
	s.myseries.password = pkt.Password
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHouse)
	bf := byteframe.NewByteFrame()
	if pkt.Destination != 9 && len(pkt.Password) > 0 && pkt.CheckPass {
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID && pkt.Password != session.myseries.password {
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

	switch pkt.Destination {
	case 3: // Others house
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID {
				bf.WriteBytes(session.myseries.houseTier)
				bf.WriteBytes(session.myseries.houseData)
				bf.WriteBytes(make([]byte, 19)) // Padding?
				bf.WriteBytes(furniture)
			}
		}
	case 4: // Bookshelf
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID {
				bf.WriteBytes(session.myseries.bookshelfData)
			}
		}
	case 5: // Gallery
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID {
				bf.WriteBytes(session.myseries.galleryData)
			}
		}
	case 8: // Tore
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID {
				bf.WriteBytes(session.myseries.toreData)
			}
		}
	case 9: // Own house
		bf.WriteBytes(furniture)
	case 10: // Garden
		for _, session := range s.server.sessions {
			if session.charID == pkt.CharID {
				bf.WriteBytes(session.myseries.gardenData)
				c, d := getGookData(s, pkt.CharID)
				bf.WriteUint16(c)
				bf.WriteUint16(0)
				bf.WriteBytes(d)
			}
		}
	}
	if len(bf.Data()) == 0 {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	} else {
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	}
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

type Title struct {
	ID       uint16    `db:"id"`
	Acquired time.Time `db:"unlocked_at"`
	Updated  time.Time `db:"updated_at"`
}

func handleMsgMhfEnumerateTitle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateTitle)
	var count uint16
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	bf.WriteUint16(0) // Unk
	rows, err := s.server.db.Queryx("SELECT id, unlocked_at, updated_at FROM titles WHERE char_id=$1", s.charID)
	if err != nil {
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}
	for rows.Next() {
		title := &Title{}
		err = rows.StructScan(&title)
		if err != nil {
			continue
		}
		count++
		bf.WriteUint16(title.ID)
		bf.WriteUint16(0) // Unk
		bf.WriteUint32(uint32(title.Acquired.Unix()))
		bf.WriteUint32(uint32(title.Updated.Unix()))
	}
	bf.Seek(0, io.SeekStart)
	bf.WriteUint16(count)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireTitle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireTitle)
	var exists int
	err := s.server.db.QueryRow("SELECT count(*) FROM titles WHERE id=$1 AND char_id=$2", pkt.TitleID, s.charID).Scan(&exists)
	if err != nil || exists == 0 {
		s.server.db.Exec("INSERT INTO titles VALUES ($1, $2, now(), now())", pkt.TitleID, s.charID)
	} else {
		s.server.db.Exec("UPDATE titles SET updated_at=now()")
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfResetTitle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateWarehouse)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(pkt.Operation)
	switch pkt.Operation {
	case 0:
		var count uint8
		itemNames := make([]string, 10)
		equipNames := make([]string, 10)
		s.server.db.QueryRow(fmt.Sprintf("%s WHERE character_id=$1", warehouseNamesQuery), s.charID).Scan(&itemNames[0],
			&itemNames[1], &itemNames[2], &itemNames[3], &itemNames[4], &itemNames[5], &itemNames[6], &itemNames[7], &itemNames[8], &itemNames[9], &equipNames[0],
			&equipNames[1], &equipNames[2], &equipNames[3], &equipNames[4], &equipNames[5], &equipNames[6], &equipNames[7], &equipNames[8], &equipNames[9])
		bf.WriteUint32(0)
		bf.WriteUint16(1000) // Usages
		temp := byteframe.NewByteFrame()
		for i, name := range itemNames {
			if len(name) > 0 {
				count++
				temp.WriteUint8(0)
				temp.WriteUint8(uint8(i))
				ps.Uint8(temp, name, true)
			}
		}
		for i, name := range equipNames {
			if len(name) > 0 {
				count++
				temp.WriteUint8(1)
				temp.WriteUint8(uint8(i))
				ps.Uint8(temp, name, true)
			}
		}
		bf.WriteUint8(count)
		bf.WriteBytes(temp.Data())
	case 1:
		bf.WriteUint8(0)
	case 2:
		s.server.db.Exec(fmt.Sprintf("UPDATE warehouse SET %s%dname=$1 WHERE character_id=$2", pkt.BoxType, pkt.BoxIndex), pkt.Name, s.charID)
	case 3:
		var t int
		err := s.server.db.QueryRow("SELECT character_id FROM warehouse WHERE character_id=$1", s.charID).Scan(&t)
		if err != nil {
			s.server.db.Exec("INSERT INTO warehouse (character_id) VALUES ($1)", s.charID)
		}
		bf.WriteUint32(0)
		bf.WriteUint16(1000) // Usages
	case 4:
		bf.WriteUint32(0)
		bf.WriteUint16(1000) // Usages
		bf.WriteUint8(0)
	}
	// Opcodes
	// 0 = Get box names
	// 1 = Commit usage
	// 2 = Rename
	// 3 = Get usage limit
	// 4 = Get gift box names (doesn't do anything?)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnumerateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateWarehouse)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // numStacks
	bf.WriteUint16(0) // Unk
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateWarehouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateWarehouse)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
