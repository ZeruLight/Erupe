package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/deltacomp"
	"erupe-ce/server/channelserver/compression/nullcomp"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func handleMsgMhfLoadPlateData(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateData)
	var data []byte

	err := db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.CharID).Scan(&data)
	if err != nil {
		s.Logger.Error("Failed to load platedata", zap.Error(err))
	}
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfSavePlateData(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateData)

	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.CharID).Scan(&data)
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

		_, err = db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", saveOutput, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platedata", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		s.Logger.Info("Wrote recompressed platedata back to DB")
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "platedata")
		// simply update database, no extra processing
		_, err := db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", pkt.RawDataPayload, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platedata", zap.Error(err))
		}
	}

	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateBox(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateBox)
	var data []byte

	err := db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.CharID).Scan(&data)
	if err != nil {
		s.Logger.Error("Failed to load platebox", zap.Error(err))
	}
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfSavePlateBox(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateBox)

	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.CharID).Scan(&data)
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

		_, err = db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", saveOutput, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platebox", zap.Error(err))
			s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		s.Logger.Info("Wrote recompressed platebox back to DB")
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "platebox")
		// simply update database, no extra processing
		_, err := db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", pkt.RawDataPayload, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to save platebox", zap.Error(err))
		}
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateMyset(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateMyset)
	var data []byte

	err := db.QueryRow("SELECT platemyset FROM characters WHERE id = $1", s.CharID).Scan(&data)
	if len(data) == 0 {
		s.Logger.Error("Failed to load platemyset", zap.Error(err))
		data = make([]byte, 1920)
	}
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfSavePlateMyset(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateMyset)

	// looks to always return the full thing, simply update database, no extra processing
	dumpSaveData(s, pkt.RawDataPayload, "platemyset")
	_, err := db.Exec("UPDATE characters SET platemyset=$1 WHERE id=$2", pkt.RawDataPayload, s.CharID)
	if err != nil {
		s.Logger.Error("Failed to save platemyset", zap.Error(err))
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
