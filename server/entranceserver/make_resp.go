package entranceserver

import (
	"encoding/binary"
	"encoding/hex"
	_config "erupe-ce/config"
	"erupe-ce/network/crypto/bin8"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/stringsupport"
	"fmt"
	"net"
)

func encodeServerInfo(config *_config.Config, server *Server, local bool) []byte {
	serverInfos := config.Entrance.Entries
	bf := byteframe.NewByteFrame()

	for serverIdx, si := range serverInfos {
		// Prevent MezFes Worlds displaying on Z1
		if config.ClientID <= _config.Z1 {
			if si.Type == 6 {
				continue
			}
		}
		if config.ClientID <= _config.G6 {
			if si.Type == 5 {
				continue
			}
		}

		if si.IP == "" {
			si.IP = config.Host
		}
		if local {
			bf.WriteUint32(0x0100007F) // 127.0.0.1
		} else {
			bf.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(si.IP).To4()))
		}
		bf.WriteUint16(uint16(serverIdx | 16))
		bf.WriteUint16(0)
		bf.WriteUint16(uint16(len(si.Channels)))
		bf.WriteUint8(si.Type)
		bf.WriteUint8(uint8(((gametime.TimeAdjusted().Unix() / 86400) + int64(serverIdx)) % 3))
		if server.erupeConfig.ClientID >= _config.G1 {
			bf.WriteUint8(si.Recommended)
		}

		fullName := append(append(stringsupport.UTF8ToSJIS(si.Name), []byte{0x00}...), stringsupport.UTF8ToSJIS(si.Description)...)
		if server.erupeConfig.ClientID >= _config.G1 && server.erupeConfig.ClientID <= _config.G5 {
			bf.WriteUint8(uint8(len(fullName)))
			bf.WriteBytes(fullName)
		} else {
			if server.erupeConfig.ClientID >= _config.G51 {
				bf.WriteUint8(0) // Ignored
			}
			bf.WriteBytes(stringsupport.PaddedString(string(fullName), 65, false))
		}

		if server.erupeConfig.ClientID >= _config.GG {
			bf.WriteUint32(si.AllowedClientFlags)
		}

		for channelIdx, ci := range si.Channels {
			sid := (serverIdx<<8 | 4096) + (channelIdx | 16)
			if _config.ErupeConfig.DebugOptions.ProxyPort != 0 {
				bf.WriteUint16(_config.ErupeConfig.DebugOptions.ProxyPort)
			} else {
				bf.WriteUint16(ci.Port)
			}
			bf.WriteUint16(uint16(channelIdx | 16))
			bf.WriteUint16(ci.MaxPlayers)
			var currentPlayers uint16
			server.db.QueryRow("SELECT current_players FROM servers WHERE server_id=$1", sid).Scan(&currentPlayers)
			bf.WriteUint16(currentPlayers)
			bf.WriteUint16(0)
			bf.WriteUint16(0)
			bf.WriteUint16(0)
			bf.WriteUint16(0)
			bf.WriteUint16(0)
			bf.WriteUint16(0)
			bf.WriteUint16(319)                  // Unk
			bf.WriteUint16(254 - currentPlayers) // Unk
			bf.WriteUint16(255 - currentPlayers) // Unk
			bf.WriteUint16(12345)
		}
	}
	bf.WriteUint32(uint32(gametime.TimeAdjusted().Unix()))
	bf.WriteUint32(uint32(server.erupeConfig.GameplayOptions.ClanMemberLimits[len(server.erupeConfig.GameplayOptions.ClanMemberLimits)-1][1]))
	return bf.Data()
}

func makeHeader(data []byte, respType string, entryCount uint16, key byte) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteBytes([]byte(respType))
	bf.WriteUint16(entryCount)
	bf.WriteUint16(uint16(len(data)))
	if len(data) > 0 {
		bf.WriteUint32(bin8.CalcSum32(data))
		bf.WriteBytes(data)
	}

	dataToEncrypt := bf.Data()

	bf = byteframe.NewByteFrame()
	bf.WriteUint8(key)
	bf.WriteBytes(bin8.EncryptBin8(dataToEncrypt, key))
	return bf.Data()
}

func makeSv2Resp(config *_config.Config, server *Server, local bool) []byte {
	serverInfos := config.Entrance.Entries
	// Decrease by the number of MezFes Worlds
	var mf int
	if config.ClientID <= _config.Z1 {
		for _, si := range serverInfos {
			if si.Type == 6 {
				mf++
			}
		}
	}
	// and Return Worlds
	var ret int
	if config.ClientID <= _config.G6 {
		for _, si := range serverInfos {
			if si.Type == 5 {
				ret++
			}
		}
	}
	rawServerData := encodeServerInfo(config, server, local)

	if server.erupeConfig.DebugOptions.LogOutboundMessages {
		fmt.Printf("[Server] -> [Client]\nData [%d bytes]:\n%s\n", len(rawServerData), hex.Dump(rawServerData))
	}

	respType := "SV2"
	if config.ClientID <= _config.G32 {
		respType = "SVR"
	}

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(makeHeader(rawServerData, respType, uint16(len(serverInfos)-(mf+ret)), 0x00))
	return bf.Data()
}

func makeUsrResp(pkt []byte, server *Server) []byte {
	bf := byteframe.NewByteFrameFromBytes(pkt)
	_ = bf.ReadUint32() // ALL+
	_ = bf.ReadUint8()  // 0x00
	userEntries := bf.ReadUint16()
	resp := byteframe.NewByteFrame()
	for i := 0; i < int(userEntries); i++ {
		cid := bf.ReadUint32()
		var sid uint16
		err := server.db.QueryRow("SELECT(SELECT server_id FROM sign_sessions WHERE char_id=$1) AS _", cid).Scan(&sid)
		if err != nil {
			resp.WriteUint16(0)
		} else {
			resp.WriteUint16(sid)
		}
		resp.WriteUint16(0)
	}

	if server.erupeConfig.DebugOptions.LogOutboundMessages {
		fmt.Printf("[Server] -> [Client]\nData [%d bytes]:\n%s\n", len(resp.Data()), hex.Dump(resp.Data()))
	}

	return makeHeader(resp.Data(), "USR", userEntries, 0x00)
}
