package channelserver

import (
	"encoding/binary"
	"errors"
	"erupe-ce/common/bfutil"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
)

const (
	pointerGender = 0x51 // +1

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
	pointerKQF         = 0x23D20 // +8

	pointerRPZ            = 0x1A076
	pointerHouseTierZ     = 0x16ECC
	pointerHouseDataZ     = 0x17161
	pointerBookshelfDataZ = 0x195F8
	pointerGalleryDataZ   = 0x19680
	pointerToreDataZ      = 0x17014
	pointerGardenDataZ    = 0x19FB8
	pointerWeaponTypeZ    = 0x16A75
	pointerWeaponIDZ      = 0x1696A
	pointerHRPZ           = 0x17156
	pointerGRPZ           = 0x1715C
	pointerKQFZ           = 0x1B080
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
	KQF           []byte

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
		err = errors.New("no savedata found")
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

	saveData.updateStructWithSaveData()

	return saveData, nil
}

func (save *CharacterSaveData) Save(s *Session) {
	if !s.kqfOverride {
		s.kqf = save.KQF
	} else {
		save.KQF = s.kqf
	}

	save.updateSaveDataWithStruct()

	err := save.Compress()
	if err != nil {
		s.logger.Error("Failed to compress savedata", zap.Error(err))
		return
	}

	_, err = s.server.db.Exec(`UPDATE characters	SET savedata=$1, is_new_character=false, hrp=$2, gr=$3, is_female=$4, weapon_type=$5, weapon_id=$6 WHERE id=$7
	`, save.compSave, save.HRP, save.GR, save.Gender, save.WeaponType, save.WeaponID, save.CharID)
	if err != nil {
		s.logger.Error("Failed to update savedata", zap.Error(err), zap.Uint32("charID", save.CharID))
	}

	s.server.db.Exec(`UPDATE user_binary SET house_tier=$1, house_data=$2, bookshelf=$3, gallery=$4, tore=$5, garden=$6 WHERE id=$7
	`, save.HouseTier, save.HouseData, save.BookshelfData, save.GalleryData, save.ToreData, save.GardenData, s.charID)
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
	if _config.ErupeConfig.RealClientMode == _config.ZZ {
		copy(save.decompSave[pointerRP:pointerRP+2], rpBytes)
		copy(save.decompSave[pointerKQF:pointerKQF+8], save.KQF)
	} else {
		copy(save.decompSave[pointerRPZ:pointerRPZ+2], rpBytes)
		copy(save.decompSave[pointerKQFZ:pointerKQFZ+8], save.KQF)
	}
}

// This will update the save struct with the values stored in the character save
func (save *CharacterSaveData) updateStructWithSaveData() {
	save.Name = stringsupport.SJISToUTF8(bfutil.UpToNull(save.decompSave[88:100]))
	if save.decompSave[pointerGender] == 1 {
		save.Gender = true
	} else {
		save.Gender = false
	}
	if !save.IsNewCharacter {
		if _config.ErupeConfig.RealClientMode == _config.ZZ {
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
			if save.HRP == uint16(999) {
				save.GR = grpToGR(binary.LittleEndian.Uint32(save.decompSave[pointerGRP : pointerGRP+4]))
			}
			save.KQF = save.decompSave[pointerKQF : pointerKQF+8]
		} else {
			save.RP = binary.LittleEndian.Uint16(save.decompSave[pointerRPZ : pointerRPZ+2])
			save.HouseTier = save.decompSave[pointerHouseTierZ : pointerHouseTierZ+5]
			save.HouseData = save.decompSave[pointerHouseDataZ : pointerHouseDataZ+195]
			save.BookshelfData = save.decompSave[pointerBookshelfDataZ : pointerBookshelfDataZ+5576]
			save.GalleryData = save.decompSave[pointerGalleryDataZ : pointerGalleryDataZ+1748]
			save.ToreData = save.decompSave[pointerToreDataZ : pointerToreDataZ+240]
			save.GardenData = save.decompSave[pointerGardenDataZ : pointerGardenDataZ+68]
			save.WeaponType = save.decompSave[pointerWeaponTypeZ]
			save.WeaponID = binary.LittleEndian.Uint16(save.decompSave[pointerWeaponIDZ : pointerWeaponIDZ+2])
			save.HRP = binary.LittleEndian.Uint16(save.decompSave[pointerHRPZ : pointerHRPZ+2])
			if save.HRP == uint16(999) {
				save.GR = grpToGR(binary.LittleEndian.Uint32(save.decompSave[pointerGRPZ : pointerGRPZ+4]))
			}
			save.KQF = save.decompSave[pointerKQFZ : pointerKQFZ+8]
		}
	}
	return
}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSexChanger)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
