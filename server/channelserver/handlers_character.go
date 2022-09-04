package channelserver

import (
	"encoding/binary"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
)

const (
	pointerGender        = 0x81    // +1
	pointerRP            = 0x22D16 // +2
	pointerHouseTier     = 0x1FB6C // +5
	pointerHouseData     = 0x1FE01 // +195
	pointerBookshelfData = 0x22298 // +5576
	// Gallery data also exists at 0x21578, is this the contest submission?
	pointerGalleryData = 0x22320 // +1748
	pointerToreData    = 0x1FCB4 // +240
	pointerGardenData  = 0x22C58 // +68
	pointerWeaponType  = 0x1F715 // +1
	pointerWeaponID    = 0x1F60A // +2
	pointerHRP         = 0x1FDF6 // +2
	pointerGRP         = 0x1FDFC // +4
)

type CharacterSaveData struct {
	CharID         uint32
	Name           string
	IsNewCharacter bool

	Gender        bool
	RP            uint16
	HouseTier     []byte
	HouseData     []byte
	BookshelfData []byte
	GalleryData   []byte
	ToreData      []byte
	GardenData    []byte
	WeaponType    uint8
	WeaponID      uint16
	HRP           uint16
	GR            uint16

	compSave   []byte
	decompSave []byte
}

func GetCharacterSaveData(s *Session, charID uint32) (*CharacterSaveData, error) {
	result, err := s.server.db.Query("SELECT id, savedata, is_new_character, name FROM characters WHERE id = $1", charID)
	if err != nil {
		s.logger.Error("Failed to get savedata", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}
	defer result.Close()
	if !result.Next() {
		s.logger.Error("No savedata found", zap.Uint32("charID", charID))
		return nil, err
	}

	saveData := &CharacterSaveData{}
	err = result.Scan(&saveData.CharID, &saveData.compSave, &saveData.IsNewCharacter, &saveData.Name)
	if err != nil {
		s.logger.Error("Failed to scan savedata", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}

	if saveData.compSave == nil {
		return saveData, nil
	}

	err = saveData.Decompress()
	if err != nil {
		s.logger.Error("Failed to decompress savedata", zap.Error(err))
		return nil, err
	}

	return saveData, nil
}

func (save *CharacterSaveData) Save(s *Session) {
	// We need to update the save data byte array before we save it back to the DB
	save.updateSaveDataWithStruct()
	save.updateStructWithSaveData()

	err := save.Compress()
	if err != nil {
		s.logger.Error("Failed to compress savedata", zap.Error(err))
		return
	}

	_, err = s.server.db.Exec(`UPDATE characters	SET savedata=$1, is_new_character=$2, hrp=$3, gr=$4, is_female=$5, weapon_type=$6, weapon_id=$7 WHERE id=$8
	`, save.compSave, save.IsNewCharacter, save.HRP, save.GR, save.Gender, save.WeaponType, save.WeaponID, save.CharID)
	if err != nil {
		s.logger.Error("Failed to update savedata", zap.Error(err), zap.Uint32("charID", save.CharID))
	}
}

func (save *CharacterSaveData) Compress() error {
	var err error
	save.compSave, err = nullcomp.Compress(save.decompSave)
	if err != nil {
		return err
	}
	return nil
}

func (save *CharacterSaveData) Decompress() error {
	var err error
	save.decompSave, err = nullcomp.Decompress(save.compSave)
	if err != nil {
		return err
	}
	return nil
}

// This will update the character save with the values stored in the save struct
func (save *CharacterSaveData) updateSaveDataWithStruct() {
	rpBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(rpBytes, save.RP)
	copy(save.decompSave[pointerRP:pointerRP+2], rpBytes)
}

// This will update the save struct with the values stored in the character save
func (save *CharacterSaveData) updateStructWithSaveData() {
	if save.decompSave[pointerGender] == 1 {
		save.Gender = true
	} else {
		save.Gender = false
	}
	save.RP = binary.LittleEndian.Uint16(save.decompSave[pointerRP : pointerRP+2])
	save.HouseTier = save.decompSave[pointerHouseTier : pointerHouseTier+5]
	save.HouseData = save.decompSave[pointerHouseData : pointerHouseData+195]
	save.BookshelfData = save.decompSave[pointerBookshelfData : pointerBookshelfData+5576]
	save.GalleryData = save.decompSave[pointerGalleryData : pointerGalleryData+1748]
	save.ToreData = save.decompSave[pointerToreData : pointerToreData+240]
	save.GardenData = save.decompSave[pointerGardenData : pointerGardenData+68]
	save.WeaponType = save.decompSave[pointerWeaponType]
	save.WeaponID = binary.LittleEndian.Uint16(save.decompSave[pointerWeaponID : pointerWeaponID+2])
	save.HRP = binary.LittleEndian.Uint16(save.decompSave[pointerHRP : pointerHRP+2])
	save.GR = grpToGR(binary.LittleEndian.Uint32(save.decompSave[pointerGRP : pointerGRP+4]))
}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSexChanger)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
