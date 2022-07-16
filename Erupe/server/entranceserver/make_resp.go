package entranceserver

import (
	"encoding/binary"
	"net"

	"erupe-ce/common/stringsupport"

	"erupe-ce/common/byteframe"
	"erupe-ce/config"
	"erupe-ce/server/channelserver"
)

// Server Entries
var season uint8

// Server Channels
var currentplayers uint16

func encodeServerInfo(serverInfos []config.EntranceServerInfo, s *Server) []byte {
	bf := byteframe.NewByteFrame()

	for serverIdx, si := range serverInfos {
		sid := (4096 + serverIdx * 256) + 16
		err := s.db.QueryRow("SELECT season FROM servers WHERE server_id=$1", sid).Scan(&season)
		if err != nil {
			panic(err)
		}
		bf.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(si.IP).To4()))
		bf.WriteUint16(16 + uint16(serverIdx))
		bf.WriteUint16(0x0000)
		bf.WriteUint16(uint16(len(si.Channels)))
		bf.WriteUint8(si.Type)
		bf.WriteUint8(season)
		bf.WriteUint8(si.Recommended)
		combined := append([]byte{0x00}, stringsupport.UTF8ToSJIS(si.Name)...)
		combined = append(combined, []byte{0x00}...)
		combined = append(combined, stringsupport.UTF8ToSJIS(si.Description)...)
		bf.WriteBytes(stringsupport.PaddedString(string(combined), 66, false))
		bf.WriteUint32(si.AllowedClientFlags)

		for channelIdx, ci := range si.Channels {
			sid = (4096 + serverIdx * 256) + (16 + channelIdx)
			bf.WriteUint16(ci.Port)
			bf.WriteUint16(16 + uint16(channelIdx))
			bf.WriteUint16(ci.MaxPlayers)
			err := s.db.QueryRow("SELECT current_players FROM servers WHERE server_id=$1", sid).Scan(&currentplayers)
			if err != nil {
				panic(err)
			}
			bf.WriteUint16(currentplayers)
			bf.WriteUint32(0)
			bf.WriteUint32(0)
			bf.WriteUint32(0)
			bf.WriteUint16(ci.Unk0)
			bf.WriteUint16(ci.Unk1)
			bf.WriteUint16(ci.Unk2)
			bf.WriteUint16(0x3039)
		}
	}
	bf.WriteUint32(uint32(channelserver.Time_Current_Adjusted().Unix()))
	bf.WriteUint32(0x0000003C)
	return bf.Data()
}

func makeHeader(data []byte, respType string, entryCount uint16, key byte) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteBytes([]byte(respType))
	bf.WriteUint16(entryCount)
	bf.WriteUint16(uint16(len(data)))
	if len(data) > 0 {
		bf.WriteUint32(CalcSum32(data))
		bf.WriteBytes(data)
	}

	dataToEncrypt := bf.Data()

	bf = byteframe.NewByteFrame()
	bf.WriteUint8(key)
	bf.WriteBytes(EncryptBin8(dataToEncrypt, key))
	return bf.Data()
}

func makeSv2Resp(servers []config.EntranceServerInfo, s *Server) []byte {
	rawServerData := encodeServerInfo(servers, s)
	bf := byteframe.NewByteFrame()
	bf.WriteBytes(makeHeader(rawServerData, "SV2", uint16(len(servers)), 0x00))
	return bf.Data()
}

func makeUsrResp(pkt []byte, s *Server) []byte {
	bf := byteframe.NewByteFrameFromBytes(pkt)
	_ = bf.ReadUint32() // ALL+
	_ = bf.ReadUint8()  // 0x00
	userEntries := bf.ReadUint16()
	resp := byteframe.NewByteFrame()
	for i := 0; i < int(userEntries); i++ {
		cid := bf.ReadUint32()
		var sid uint16
		err := s.db.QueryRow("SELECT(SELECT server_id FROM sign_sessions WHERE char_id=$1) AS _", cid).Scan(&sid)
		if err != nil {
			resp.WriteBytes(make([]byte, 4))
			continue
		} else {
			resp.WriteUint16(sid)
			resp.WriteUint16(0)
		}
	}
	return makeHeader(resp.Data(), "USR", userEntries, 0x00)
}
