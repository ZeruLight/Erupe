package channelserver

import (
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	ps "erupe-ce/utils/pascalstring"

	"go.uber.org/zap"
)

// BroadcastMHF queues a MHFPacket to be sent to all sessions.
func (server *ChannelServer) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	// Broadcast the data.
	server.Lock()
	defer server.Unlock()
	for _, session := range server.sessions {
		if session == ignoredSession {
			continue
		}
		session.QueueSendMHF(pkt)
	}
}

func (server *ChannelServer) WorldcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session, ignoredChannel *ChannelServer) {
	for _, c := range server.Channels {
		if c == ignoredChannel {
			continue
		}
		c.BroadcastMHF(pkt, ignoredSession)
	}
}

// BroadcastChatMessage broadcasts a simple chat message to all the sessions.
func (server *ChannelServer) BroadcastChatMessage(message string) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      0x80,
		Message:    message,
		SenderName: server.name,
	}
	msgBinChat.Build(bf)

	server.BroadcastMHF(&mhfpacket.MsgSysCastedBinary{
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}, nil)
}

func (server *ChannelServer) BroadcastRaviente(ip uint32, port uint16, stage []byte, _type uint8) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint16(0)    // Unk
	bf.WriteUint16(0x43) // Data len
	bf.WriteUint16(3)    // Unk len
	var text string
	switch _type {
	case 2:
		text = server.i18n.raviente.berserk
	case 3:
		text = server.i18n.raviente.extreme
	case 4:
		text = server.i18n.raviente.extremeLimited
	case 5:
		text = server.i18n.raviente.berserkSmall
	default:
		server.logger.Error("Unk raviente type", zap.Uint8("_type", _type))
	}
	ps.Uint16(bf, text, true)
	bf.WriteBytes([]byte{0x5F, 0x53, 0x00})
	bf.WriteUint32(ip)   // IP address
	bf.WriteUint16(port) // Port
	bf.WriteUint16(0)    // Unk
	bf.WriteBytes(stage)
	server.WorldcastMHF(&mhfpacket.MsgSysCastedBinary{
		BroadcastType:  BroadcastTypeServer,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}, nil, server)
}
