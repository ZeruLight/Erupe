package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/deltacomp"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"erupe-ce/utils/db"
	"fmt"

	"go.uber.org/zap"
)

func handleMsgMhfLoadPlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateData)
	var data []byte
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.CharID).Scan(&data)
	if err != nil {
		s.Logger.Error("Failed to load platedata", zap.Error(err))
	}
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfSavePlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateData)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := database.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.CharID).Scan(&data)
		if err != nil {
			s.Logger.Error("Failed to load platedata", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		if len(data) > 0 {
			// Decompress
			s.Logger.Info("Decompressing...")
			data, err = nullcomp.Decompress(data)
			if err != nil {
				s.Logger.Error("Failed to decompress platedata", zap.Error(err))
				s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
				return
			}
		} else {
			// create empty save if absent
			data = make([]byte, 140000)
		}

		// Perform diff and compress it to write back to db
		s.Logger.Info("Diffing...")
		saveOutput, err := nullcomp.Compress(deltacomp.ApplyDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.Logger.Error("Failed to diff and compress platedata", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		_, err = database.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", saveOutput, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platedata", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		s.Logger.Info("Wrote recompressed platedata back to DB")
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "platedata")
		// simply update database, no extra processing
		_, err := database.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", pkt.RawDataPayload, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platedata", zap.Error(err))
		}
	}

	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateBox)
	var data []byte
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.CharID).Scan(&data)
	if err != nil {
		s.Logger.Error("Failed to load platebox", zap.Error(err))
	}
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfSavePlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateBox)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := database.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.CharID).Scan(&data)
		if err != nil {
			s.Logger.Error("Failed to load platebox", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		// Decompress
		if len(data) > 0 {
			// Decompress
			s.Logger.Info("Decompressing...")
			data, err = nullcomp.Decompress(data)
			if err != nil {
				s.Logger.Error("Failed to decompress platebox", zap.Error(err))
				s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
				return
			}
		} else {
			// create empty save if absent
			data = make([]byte, 4800)
		}

		// Perform diff and compress it to write back to db
		s.Logger.Info("Diffing...")
		saveOutput, err := nullcomp.Compress(deltacomp.ApplyDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.Logger.Error("Failed to diff and compress platebox", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		_, err = database.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", saveOutput, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platebox", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		s.Logger.Info("Wrote recompressed platebox back to DB")
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "platebox")
		// simply update database, no extra processing
		_, err := database.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", pkt.RawDataPayload, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platebox", zap.Error(err))
		}
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateMyset)
	var data []byte
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT platemyset FROM characters WHERE id = $1", s.CharID).Scan(&data)
	if len(data) == 0 {
		s.Logger.Error("Failed to load platemyset", zap.Error(err))
		data = make([]byte, 1920)
	}
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfSavePlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateMyset)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	// looks to always return the full thing, simply update database, no extra processing
	dumpSaveData(s, pkt.RawDataPayload, "platemyset")
	_, err = database.Exec("UPDATE characters SET platemyset=$1 WHERE id=$2", pkt.RawDataPayload, s.CharID)
	if err != nil {
		s.Logger.Error("Failed to save platemyset", zap.Error(err))
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
