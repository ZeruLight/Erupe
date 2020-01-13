package entranceserver

import (
	"encoding/binary"
	"net"

	"github.com/Andoryuuta/Erupe/config"
	"github.com/Andoryuuta/byteframe"
)

func paddedString(x string, size uint) []byte {
	out := make([]byte, size)
	copy(out, x)

	// Null terminate it.
	out[len(out)-1] = 0
	return out
}

func encodeServerInfo(serverInfos []config.EntranceServerInfo) []byte {
	bf := byteframe.NewByteFrame()

	for serverIdx, si := range serverInfos {
		bf.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(si.IP).To4()))
		bf.WriteUint16(16 + uint16(serverIdx))
		bf.WriteUint16(si.Unk2)
		bf.WriteUint16(uint16(len(si.Channels)))
		bf.WriteUint8(si.Type)
		bf.WriteUint8(si.Season)
		bf.WriteUint8(si.Unk6)
		bf.WriteBytes(paddedString(si.Name, 66))
		bf.WriteUint32(si.AllowedClientFlags)

		for channelIdx, ci := range si.Channels {
			bf.WriteUint16(ci.Port)
			bf.WriteUint16(16 + uint16(channelIdx))
			bf.WriteUint16(ci.MaxPlayers)
			bf.WriteUint16(ci.CurrentPlayers)
			bf.WriteUint16(ci.Unk4)
			bf.WriteUint16(ci.Unk5)
			bf.WriteUint16(ci.Unk6)
			bf.WriteUint16(ci.Unk7)
			bf.WriteUint16(ci.Unk8)
			bf.WriteUint16(ci.Unk9)
			bf.WriteUint16(ci.Unk10)
			bf.WriteUint16(ci.Unk11)
			bf.WriteUint16(ci.Unk12)
			bf.WriteUint16(ci.Unk13)
		}
	}

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

func makeResp(servers []config.EntranceServerInfo) []byte {
	rawServerData := encodeServerInfo(servers)

	bf := byteframe.NewByteFrame()
	bf.WriteBytes(makeHeader(rawServerData, "SV2", uint16(len(servers)), 0x00))

	// TODO(Andoryuuta): Figure out what this user data is.
	// Is it for the friends list at the world selection screen?
	// If so, how does it work without the entrance server connection being authenticated?
	bf.WriteBytes(makeHeader([]byte{}, "USR", 0, 0x00))

	return bf.Data()

}
