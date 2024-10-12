package mhfitem

import (
	"erupe-ce/config"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/token"
)

type MHFItem struct {
	ItemID uint16
}

type MHFSigilEffect struct {
	ID    uint16
	Level uint16
}

type MHFSigil struct {
	Effects []MHFSigilEffect
	Unk0    uint8
	Unk1    uint8
	Unk2    uint8
	Unk3    uint8
}

type MHFEquipment struct {
	WarehouseID uint32
	ItemType    uint8
	Unk0        uint8
	ItemID      uint16
	Level       uint16
	Decorations []MHFItem
	Sigils      []MHFSigil
	Unk1        uint16
}

type MHFItemStack struct {
	WarehouseID uint32
	Item        MHFItem
	Quantity    uint16
	Unk0        uint32
}

func ReadWarehouseItem(bf *byteframe.ByteFrame) MHFItemStack {
	var item MHFItemStack
	item.WarehouseID = bf.ReadUint32()
	if item.WarehouseID == 0 {
		item.WarehouseID = token.RNG.Uint32()
	}
	item.Item.ItemID = bf.ReadUint16()
	item.Quantity = bf.ReadUint16()
	item.Unk0 = bf.ReadUint32()
	return item
}

func DiffItemStacks(o []MHFItemStack, u []MHFItemStack) []MHFItemStack {
	// o = old, u = update, f = final
	var f []MHFItemStack
	for _, uItem := range u {
		exists := false
		for i := range o {
			if o[i].WarehouseID == uItem.WarehouseID {
				exists = true
				o[i].Quantity = uItem.Quantity
			}
		}
		if !exists {
			uItem.WarehouseID = token.RNG.Uint32()
			f = append(f, uItem)
		}
	}
	for _, oItem := range o {
		if oItem.Quantity > 0 {
			f = append(f, oItem)
		}
	}
	return f
}

func (is MHFItemStack) ToBytes() []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(is.WarehouseID)
	bf.WriteUint16(is.Item.ItemID)
	bf.WriteUint16(is.Quantity)
	bf.WriteUint32(is.Unk0)
	return bf.Data()
}

func SerializeWarehouseItems(i []MHFItemStack) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(i)))
	bf.WriteUint16(0) // Unused
	for _, j := range i {
		bf.WriteBytes(j.ToBytes())
	}
	return bf.Data()
}

func ReadWarehouseEquipment(bf *byteframe.ByteFrame) MHFEquipment {
	var equipment MHFEquipment
	equipment.Decorations = make([]MHFItem, 3)
	equipment.Sigils = make([]MHFSigil, 3)
	for i := 0; i < 3; i++ {
		equipment.Sigils[i].Effects = make([]MHFSigilEffect, 3)
	}
	equipment.WarehouseID = bf.ReadUint32()
	if equipment.WarehouseID == 0 {
		equipment.WarehouseID = token.RNG.Uint32()
	}
	equipment.ItemType = bf.ReadUint8()
	equipment.Unk0 = bf.ReadUint8()
	equipment.ItemID = bf.ReadUint16()
	equipment.Level = bf.ReadUint16()
	for i := 0; i < 3; i++ {
		equipment.Decorations[i].ItemID = bf.ReadUint16()
	}
	if config.GetConfig().ClientID >= config.G1 {
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				equipment.Sigils[i].Effects[j].ID = bf.ReadUint16()
			}
			for j := 0; j < 3; j++ {
				equipment.Sigils[i].Effects[j].Level = bf.ReadUint16()
			}
			equipment.Sigils[i].Unk0 = bf.ReadUint8()
			equipment.Sigils[i].Unk1 = bf.ReadUint8()
			equipment.Sigils[i].Unk2 = bf.ReadUint8()
			equipment.Sigils[i].Unk3 = bf.ReadUint8()
		}
	}
	if config.GetConfig().ClientID >= config.Z1 {
		equipment.Unk1 = bf.ReadUint16()
	}
	return equipment
}

func (e MHFEquipment) ToBytes() []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(e.WarehouseID)
	bf.WriteUint8(e.ItemType)
	bf.WriteUint8(e.Unk0)
	bf.WriteUint16(e.ItemID)
	bf.WriteUint16(e.Level)
	for i := 0; i < 3; i++ {
		bf.WriteUint16(e.Decorations[i].ItemID)
	}
	if config.GetConfig().ClientID >= config.G1 {
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				bf.WriteUint16(e.Sigils[i].Effects[j].ID)
			}
			for j := 0; j < 3; j++ {
				bf.WriteUint16(e.Sigils[i].Effects[j].Level)
			}
			bf.WriteUint8(e.Sigils[i].Unk0)
			bf.WriteUint8(e.Sigils[i].Unk1)
			bf.WriteUint8(e.Sigils[i].Unk2)
			bf.WriteUint8(e.Sigils[i].Unk3)
		}
	}
	if config.GetConfig().ClientID >= config.Z1 {
		bf.WriteUint16(e.Unk1)
	}
	return bf.Data()
}

func SerializeWarehouseEquipment(i []MHFEquipment) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(i)))
	bf.WriteUint16(0) // Unused
	for _, j := range i {
		bf.WriteBytes(j.ToBytes())
	}
	return bf.Data()
}
