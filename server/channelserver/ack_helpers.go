package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
)

func doAckEarthSucceed(s *Session, ackHandle uint32, data []*byteframe.ByteFrame) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(s.server.erupeConfig.EarthID))
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(uint32(len(data)))
	for i := range data {
		bf.WriteBytes(data[i].Data())
	}
	doAckBufSucceed(s, ackHandle, bf.Data())
}

func doAckBufSucceed(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        0,
		AckData:          data,
	})
}

func doAckBufFail(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        1,
		AckData:          data,
	})
}

func doAckSimpleSucceed(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        0,
		AckData:          data,
	})
}

func doAckSimpleFail(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        1,
		AckData:          data,
	})
}
