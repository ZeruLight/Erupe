package channelserver

import (
	"erupe-ce/config"
	"erupe-ce/internal/model"
	"erupe-ce/internal/service"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"erupe-ce/utils/mhfitem"
	ps "erupe-ce/utils/pascalstring"
	"erupe-ce/utils/stringsupport"
	"erupe-ce/utils/token"
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
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

func handleMsgMhfUpdateInterior(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateInterior)

	db.Exec(`UPDATE user_binary SET house_furniture=$1 WHERE id=$2`, pkt.InteriorData, s.CharID)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateHouse(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateHouse)

	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	var houses []model.HouseData
	houseQuery := `SELECT c.id, hr, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
		FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE c.id=$1`
	switch pkt.Method {
	case 1:
		var friendsList string
		db.QueryRow("SELECT friends FROM characters WHERE id=$1", s.CharID).Scan(&friendsList)
		cids := stringsupport.CSVElems(friendsList)
		for _, cid := range cids {
			house := model.HouseData{}
			row := db.QueryRowx(houseQuery, cid)
			err := row.StructScan(&house)
			if err == nil {
				houses = append(houses, house)
			}
		}
	case 2:
		guild, err := service.GetGuildInfoByCharacterId(s.CharID)
		if err != nil || guild == nil {
			break
		}
		guildMembers, err := service.GetGuildMembers(guild.ID, false)
		if err != nil {
			break
		}
		for _, member := range guildMembers {
			house := model.HouseData{}
			row := db.QueryRowx(houseQuery, member.CharID)
			err = row.StructScan(&house)
			if err == nil {
				houses = append(houses, house)
			}
		}
	case 3:
		houseQuery = `SELECT c.id, hr, gr, name, COALESCE(ub.house_state, 2) as house_state, COALESCE(ub.house_password, '') as house_password
			FROM characters c LEFT JOIN user_binary ub ON ub.id = c.id WHERE name ILIKE $1`
		house := model.HouseData{}
		rows, _ := db.Queryx(houseQuery, fmt.Sprintf(`%%%s%%`, pkt.Name))
		for rows.Next() {
			err := rows.StructScan(&house)
			if err == nil {
				houses = append(houses, house)
			}
		}
	case 4:
		house := model.HouseData{}
		row := db.QueryRowx(houseQuery, pkt.CharID)
		err := row.StructScan(&house)
		if err == nil {
			houses = append(houses, house)
		}
	case 5: // Recent visitors
		break
	}
	for _, house := range houses {
		bf.WriteUint32(house.CharID)
		bf.WriteUint8(house.HouseState)
		if len(house.HousePassword) > 0 {
			bf.WriteUint8(3)
		} else {
			bf.WriteUint8(0)
		}
		bf.WriteUint16(house.HR)
		if config.GetConfig().ClientID >= config.G10 {
			bf.WriteUint16(house.GR)
		}
		ps.Uint8(bf, house.Name, true)
	}
	bf.Seek(0, 0)
	bf.WriteUint16(uint16(len(houses)))
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateHouse(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateHouse)

	// 01 = closed
	// 02 = open anyone
	// 03 = open friends
	// 04 = open guild
	// 05 = open friends+guild
	db.Exec(`UPDATE user_binary SET house_state=$1, house_password=$2 WHERE id=$3`, pkt.State, pkt.Password, s.CharID)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadHouse(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHouse)

	bf := byteframe.NewByteFrame()

	var state uint8
	var password string
	db.QueryRow(`SELECT COALESCE(house_state, 2) as house_state, COALESCE(house_password, '') as house_password FROM user_binary WHERE id=$1
	`, pkt.CharID).Scan(&state, &password)

	if pkt.Destination != 9 && len(pkt.Password) > 0 && pkt.CheckPass {
		if pkt.Password != password {
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	if pkt.Destination != 9 && state > 2 {
		allowed := false

		// Friends list verification
		if state == 3 || state == 5 {
			var friendsList string
			db.QueryRow(`SELECT friends FROM characters WHERE id=$1`, pkt.CharID).Scan(&friendsList)
			cids := stringsupport.CSVElems(friendsList)
			for _, cid := range cids {
				if uint32(cid) == s.CharID {
					allowed = true
					break
				}
			}
		}

		// Guild verification
		if state > 3 {
			ownGuild, err := service.GetGuildInfoByCharacterId(s.CharID)
			isApplicant, _ := ownGuild.HasApplicationForCharID(s.CharID)
			if err == nil && ownGuild != nil {
				othersGuild, err := service.GetGuildInfoByCharacterId(pkt.CharID)
				if err == nil && othersGuild != nil {
					if othersGuild.ID == ownGuild.ID && !isApplicant {
						allowed = true
					}
				}
			}
		}

		if !allowed {
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	var houseTier, houseData, houseFurniture, bookshelf, gallery, tore, garden []byte
	db.QueryRow(`SELECT house_tier, house_data, house_furniture, bookshelf, gallery, tore, garden FROM user_binary WHERE id=$1
	`, pkt.CharID).Scan(&houseTier, &houseData, &houseFurniture, &bookshelf, &gallery, &tore, &garden)
	if houseFurniture == nil {
		houseFurniture = make([]byte, 20)
	}

	switch pkt.Destination {
	case 3: // Others house
		bf.WriteBytes(houseTier)
		bf.WriteBytes(houseData)
		bf.WriteBytes(make([]byte, 19)) // Padding?
		bf.WriteBytes(houseFurniture)
	case 4: // Bookshelf
		bf.WriteBytes(bookshelf)
	case 5: // Gallery
		bf.WriteBytes(gallery)
	case 8: // Tore
		bf.WriteBytes(tore)
	case 9: // Own house
		bf.WriteBytes(houseFurniture)
	case 10: // Garden
		bf.WriteBytes(garden)
		goocoos := getGoocooData(s, pkt.CharID)
		bf.WriteUint16(uint16(len(goocoos)))
		bf.WriteUint16(0)
		for _, goocoo := range goocoos {
			bf.WriteBytes(goocoo)
		}
	}
	if len(bf.Data()) == 0 {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
	} else {
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	}
}

func handleMsgMhfGetMyhouseInfo(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetMyhouseInfo)

	var data []byte
	db.QueryRow(`SELECT mission FROM user_binary WHERE id=$1`, s.CharID).Scan(&data)
	if len(data) > 0 {
		s.DoAckBufSucceed(pkt.AckHandle, data)
	} else {
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 9))
	}
}

func handleMsgMhfUpdateMyhouseInfo(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateMyhouseInfo)

	db.Exec("UPDATE user_binary SET mission=$1 WHERE id=$2", pkt.Data, s.CharID)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadDecoMyset(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadDecoMyset)

	var data []byte
	err := db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.CharID).Scan(&data)
	if err != nil {
		s.Logger.Error("Failed to load decomyset", zap.Error(err))
	}
	if len(data) == 0 {
		data = []byte{0x01, 0x00}
		if config.GetConfig().ClientID < config.G10 {
			data = []byte{0x00, 0x00}
		}
	}
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfSaveDecoMyset(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveDecoMyset)

	var temp []byte
	err := db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.CharID).Scan(&temp)
	if err != nil {
		s.Logger.Error("Failed to load decomyset", zap.Error(err))
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		return
	}

	// Version handling
	bf := byteframe.NewByteFrame()
	var size uint
	if config.GetConfig().ClientID >= config.G10 {
		size = 76
		bf.WriteUint8(1)
	} else {
		size = 68
		bf.WriteUint8(0)
	}

	// Handle nil data
	if len(temp) == 0 {
		temp = append(bf.Data(), uint8(0))
	}

	// Build a map of set data
	sets := make(map[uint16][]byte)
	oldSets := byteframe.NewByteFrameFromBytes(temp[2:])
	for i := uint8(0); i < temp[1]; i++ {
		index := oldSets.ReadUint16()
		sets[index] = oldSets.ReadBytes(size)
	}

	// Overwrite existing sets
	newSets := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload[2:])
	for i := uint8(0); i < pkt.RawDataPayload[1]; i++ {
		index := newSets.ReadUint16()
		sets[index] = newSets.ReadBytes(size)
	}

	// Serialise the set data
	bf.WriteUint8(uint8(len(sets)))
	for u, b := range sets {
		bf.WriteUint16(u)
		bf.WriteBytes(b)
	}

	dumpSaveData(s, bf.Data(), "decomyset")
	db.Exec("UPDATE characters SET decomyset=$1 WHERE id=$2", bf.Data(), s.CharID)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateTitle(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateTitle)

	var count uint16
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	bf.WriteUint16(0) // Unk
	rows, err := db.Queryx("SELECT id, unlocked_at, updated_at FROM titles WHERE char_id=$1", s.CharID)
	if err != nil {
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
		return
	}
	for rows.Next() {
		title := &model.Title{}
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
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireTitle(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireTitle)

	for _, title := range pkt.TitleIDs {
		var exists int
		err := db.QueryRow(`SELECT count(*) FROM titles WHERE id=$1 AND char_id=$2`, title, s.CharID).Scan(&exists)
		if err != nil || exists == 0 {
			db.Exec(`INSERT INTO titles VALUES ($1, $2, now(), now())`, title, s.CharID)
		} else {
			db.Exec(`UPDATE titles SET updated_at=now() WHERE id=$1 AND char_id=$2`, title, s.CharID)
		}
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfResetTitle(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func initializeWarehouse(s *Session) {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var t int
	err = db.QueryRow("SELECT character_id FROM warehouse WHERE character_id=$1", s.CharID).Scan(&t)
	if err != nil {
		db.Exec("INSERT INTO warehouse (character_id) VALUES ($1)", s.CharID)
	}
}

func handleMsgMhfOperateWarehouse(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateWarehouse)

	initializeWarehouse(s)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(pkt.Operation)
	switch pkt.Operation {
	case 0:
		var count uint8
		itemNames := make([]string, 10)
		equipNames := make([]string, 10)
		db.QueryRow(fmt.Sprintf("%s WHERE character_id=$1", warehouseNamesQuery), s.CharID).Scan(&itemNames[0],
			&itemNames[1], &itemNames[2], &itemNames[3], &itemNames[4], &itemNames[5], &itemNames[6], &itemNames[7], &itemNames[8], &itemNames[9], &equipNames[0],
			&equipNames[1], &equipNames[2], &equipNames[3], &equipNames[4], &equipNames[5], &equipNames[6], &equipNames[7], &equipNames[8], &equipNames[9])
		bf.WriteUint32(0)
		bf.WriteUint16(10000) // Usages
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
		switch pkt.BoxType {
		case 0:
			db.Exec(fmt.Sprintf("UPDATE warehouse SET item%dname=$1 WHERE character_id=$2", pkt.BoxIndex), pkt.Name, s.CharID)
		case 1:
			db.Exec(fmt.Sprintf("UPDATE warehouse SET equip%dname=$1 WHERE character_id=$2", pkt.BoxIndex), pkt.Name, s.CharID)
		}
	case 3:
		bf.WriteUint32(0)     // Usage renewal time, >1 = disabled
		bf.WriteUint16(10000) // Usages
	case 4:
		bf.WriteUint32(0)
		bf.WriteUint16(10000) // Usages
		bf.WriteUint8(0)
	}
	// Opcodes
	// 0 = Get box names
	// 1 = Commit usage
	// 2 = Rename
	// 3 = Get usage limit
	// 4 = Get gift box names (doesn't do anything?)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func addWarehouseItem(s *Session, item mhfitem.MHFItemStack) {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	giftBox := warehouseGetItems(s, 10)
	item.WarehouseID = token.RNG.Uint32()
	giftBox = append(giftBox, item)
	db.Exec("UPDATE warehouse SET item10=$1 WHERE character_id=$2", mhfitem.SerializeWarehouseItems(giftBox), s.CharID)
}

func addWarehouseEquipment(s *Session, equipment mhfitem.MHFEquipment) {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	giftBox := warehouseGetEquipment(s, 10)
	equipment.WarehouseID = token.RNG.Uint32()
	giftBox = append(giftBox, equipment)
	db.Exec("UPDATE warehouse SET equip10=$1 WHERE character_id=$2", mhfitem.SerializeWarehouseEquipment(giftBox), s.CharID)
}

func warehouseGetItems(s *Session, index uint8) []mhfitem.MHFItemStack {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	initializeWarehouse(s)
	var data []byte
	var items []mhfitem.MHFItemStack
	db.QueryRow(fmt.Sprintf(`SELECT item%d FROM warehouse WHERE character_id=$1`, index), s.CharID).Scan(&data)
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			items = append(items, mhfitem.ReadWarehouseItem(box))
		}
	}
	return items
}

func warehouseGetEquipment(s *Session, index uint8) []mhfitem.MHFEquipment {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var data []byte
	var equipment []mhfitem.MHFEquipment
	db.QueryRow(fmt.Sprintf(`SELECT equip%d FROM warehouse WHERE character_id=$1`, index), s.CharID).Scan(&data)
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			equipment = append(equipment, mhfitem.ReadWarehouseEquipment(box))
		}
	}
	return equipment
}

func handleMsgMhfEnumerateWarehouse(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateWarehouse)
	bf := byteframe.NewByteFrame()
	switch pkt.BoxType {
	case 0:
		items := warehouseGetItems(s, pkt.BoxIndex)
		bf.WriteBytes(mhfitem.SerializeWarehouseItems(items))
	case 1:
		equipment := warehouseGetEquipment(s, pkt.BoxIndex)
		bf.WriteBytes(mhfitem.SerializeWarehouseEquipment(equipment))
	}
	if bf.Index() > 0 {
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	} else {
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfUpdateWarehouse(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateWarehouse)

	switch pkt.BoxType {
	case 0:
		newStacks := mhfitem.DiffItemStacks(warehouseGetItems(s, pkt.BoxIndex), pkt.UpdatedItems)
		db.Exec(fmt.Sprintf(`UPDATE warehouse SET item%d=$1 WHERE character_id=$2`, pkt.BoxIndex), mhfitem.SerializeWarehouseItems(newStacks), s.CharID)
	case 1:
		var fEquip []mhfitem.MHFEquipment
		oEquips := warehouseGetEquipment(s, pkt.BoxIndex)
		for _, uEquip := range pkt.UpdatedEquipment {
			exists := false
			for i := range oEquips {
				if oEquips[i].WarehouseID == uEquip.WarehouseID {
					exists = true
					// Will set removed items to 0
					oEquips[i].ItemID = uEquip.ItemID
					break
				}
			}
			if !exists {
				uEquip.WarehouseID = token.RNG.Uint32()
				fEquip = append(fEquip, uEquip)
			}
		}
		for _, oEquip := range oEquips {
			if oEquip.ItemID > 0 {
				fEquip = append(fEquip, oEquip)
			}
		}
		db.Exec(fmt.Sprintf(`UPDATE warehouse SET equip%d=$1 WHERE character_id=$2`, pkt.BoxIndex), mhfitem.SerializeWarehouseEquipment(fEquip), s.CharID)
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}
