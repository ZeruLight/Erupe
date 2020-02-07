package channelserver

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/Andoryuuta/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// Temporary function to just return no results for a MSG_MHF_ENUMERATE* packet
func stubEnumerateNoResults(s *Session, ackHandle uint32) {
	enumBf := byteframe.NewByteFrame()
	enumBf.WriteUint16(0) // Entry count (count for quests, rankings, events, etc.)

	doSizedAckResp(s, ackHandle, enumBf.Data())
}

// Temporary function to just return no results for many MSG_MHF_GET* packets.
func stubGetNoResults(s *Session, ackHandle uint32) {
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0x0A218EAD) // Unk shared ID. Sent in response of MSG_MHF_GET_TOWER_INFO, MSG_MHF_GET_PAPER_DATA etc. (World ID?)
	resp.WriteUint32(0)          // Unk
	resp.WriteUint32(0)          // Unk
	resp.WriteUint32(0)          // Entry count

	doSizedAckResp(s, ackHandle, resp.Data())
}

// Some common ACK response header that a lot (but not all) of the packet responses use.
func doSizedAckResp(s *Session, ackHandle uint32, data []byte) {
	// Wrap the data into another container with the data size.
	bfw := byteframe.NewByteFrame()
	bfw.WriteUint8(1)                  // Unk
	bfw.WriteUint8(0)                  // Unk
	bfw.WriteUint16(uint16(len(data))) // Data size
	if len(data) > 0 {
		bfw.WriteBytes(data)
	}

	s.QueueAck(ackHandle, bfw.Data())
}

func updateRights(s *Session) {
	update := &mhfpacket.MsgSysUpdateRight{
		Unk0: 0,
		Unk1: 0x4E,
		Rights: []mhfpacket.ClientRight{
			{
				ID:        1,
				Timestamp: 0,
			},
			{
				ID:        2,
				Timestamp: 0x5dfa14c0,
			},
			{
				ID:        3,
				Timestamp: 0x5dfa14c0,
			},
			{
				ID:        6,
				Timestamp: 0x5de70510,
			},
		},
		UnkSize: 0,
	}
	s.QueueSendMHF(update)
}

func fixedSizeShiftJIS(text string, size int) []byte {
	r := bytes.NewBuffer([]byte(text))
	encoded, err := ioutil.ReadAll(transform.NewReader(r, japanese.ShiftJIS.NewEncoder()))
	if err != nil {
		panic(err)
	}

	out := make([]byte, size)
	copy(out, encoded)

	// Null terminate it.
	out[len(out)-1] = 0
	return out
}

// TODO(Andoryuuta): Fix/move/remove me!
func stripNullTerminator(x string) string {
	return strings.SplitN(x, "\x00", 2)[0]
}

func handleMsgHead(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve01(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve02(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve03(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve04(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve05(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve06(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve07(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAddObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDelObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDispObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysHideObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve0C(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve0D(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve0E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysExtendThreshold(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysEnd(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysNop(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysAck(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysTerminalLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysTerminalLog)

	resp := byteframe.NewByteFrame()
	/*
		if pkt.LogID == 0{
			fmt.Println("New log session")
		}
	*/
	resp.WriteUint32(0)          // UNK
	resp.WriteUint32(0x98bd51a9) // LogID to use for requests after this.
	s.QueueAck(pkt.AckHandle, resp.Data())
}

func handleMsgSysLogin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLogin)

	s.Lock()
	s.charID = pkt.CharID0
	s.Unlock()

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)                         // Unk
	bf.WriteUint32(uint32(time.Now().Unix())) // Unix timestamp
	s.QueueAck(pkt.AckHandle, bf.Data())
}

func handleMsgSysLogout(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetStatus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysPing(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysPing)

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // Unk
	bf.WriteUint32(0) // Unk
	s.QueueAck(pkt.AckHandle, bf.Data())
}

func handleMsgSysCastBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCastBinary)

	if pkt.Type0 == 3 && pkt.Type1 == 1 {
		fmt.Println("Got chat message!")

		resp := &mhfpacket.MsgSysCastedBinary{
			CharID:         s.charID,
			Type0:          1,
			Type1:          1,
			RawDataPayload: pkt.RawDataPayload,
		}
		s.server.BroadcastMHF(resp, s)

		/*
			// Made the inside of the casted binary
			payload := byteframe.NewByteFrame()
			payload.WriteUint16(uint16(i)) // Chat type

			//Chat type 0 = World
			//Chat type 1 = Local
			//Chat type 2 = Guild
			//Chat type 3 = Alliance
			//Chat type 4 = Party
			//Chat type 5 = Whisper
			//Thanks to @Alice on discord for identifying these.

			payload.WriteUint8(0) // Unknown
			msg := fmt.Sprintf("Chat type %d", i)
			playername := fmt.Sprintf("Ando")
			payload.WriteUint16(uint16(len(playername) + 1))
			payload.WriteUint16(uint16(len(msg) + 1))
			payload.WriteUint8(0) // Is this correct, or do I have the endianess of the prev 2 fields wrong?
			payload.WriteNullTerminatedBytes([]byte(msg))
			payload.WriteNullTerminatedBytes([]byte(playername))
			payloadBytes := payload.Data()

			//Wrap it in a CASTED_BINARY packet to broadcast
			bfw := byteframe.NewByteFrame()
			bfw.WriteUint16(uint16(network.MSG_SYS_CASTED_BINARY))
			bfw.WriteUint32(0x23325A29) // Character ID
			bfw.WriteUint8(1)           // type
			bfw.WriteUint8(1)           // type2
			bfw.WriteUint16(uint16(len(payloadBytes)))
			bfw.WriteBytes(payloadBytes)
		*/
	}
}

func handleMsgSysHideClient(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysHideClient)
}

func handleMsgSysTime(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysTime)

	resp := &mhfpacket.MsgSysTime{
		GetRemoteTime: false,
		Timestamp:     uint32(time.Now().Unix()),
	}
	s.QueueSendMHF(resp)
}

func handleMsgSysCastedBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysGetFile(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetFile)

	if !pkt.IsScenario {
		// Get quest file.
		data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("quests/%s.bin", stripNullTerminator(pkt.Filename))))
		if err != nil {
			panic(err)
		}

		doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		s.logger.Fatal("scenario getfile not implemented.")
	}

}

func handleMsgSysIssueLogkey(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysIssueLogkey)

	// Make a random log key for this session.
	logKey := make([]byte, 16)
	_, err := rand.Read(logKey)
	if err != nil {
		panic(err)
	}

	// TODO(Andoryuuta): In the offical client, the log key index is off by one,
	// cutting off the last byte in _most uses_. Find and document these accordingly.
	s.Lock()
	s.logKey = logKey
	s.Unlock()

	// Issue it.
	resp := byteframe.NewByteFrame()
	resp.WriteBytes(logKey)
	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysRecordLog(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEcho(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCreateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateStage)

	s.server.stagesLock.Lock()
	stage := NewStage(stripNullTerminator(pkt.StageID))
	s.server.stages[stage.id] = stage
	s.server.stagesLock.Unlock()

	resp := make([]byte, 8) // Unk resp.
	s.QueueAck(pkt.AckHandle, resp)
}

func handleMsgSysStageDestruct(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEnterStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnterStage)

	// Remove this session from old stage clients list and put myself in the new one.
	s.server.stagesLock.Lock()
	newStage, gotNewStage := s.server.stages[stripNullTerminator(pkt.StageID)]
	s.server.stagesLock.Unlock()

	// Remove from old stage.
	if s.stage != nil {
		s.stage.Lock()
		delete(s.stage.clients, s)
		s.stage.Unlock()
	}

	// Add the new stage.
	if gotNewStage {
		newStage.Lock()
		newStage.clients[s] = s.charID
		newStage.Unlock()
	}

	// Save our new stage ID and pointer to the new stage itself.
	s.Lock()
	s.stageID = string(stripNullTerminator(pkt.StageID))
	s.stage = newStage
	s.Unlock()

	// Tell the client to cleanup its current stage objects.
	s.QueueSendMHF(&mhfpacket.MsgSysCleanupObject{})

	// Confirm the stage entry.
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// Notify existing stage clients that this new client has entered.
	s.logger.Info("Sending MsgSysInsertUser & MsgSysNotifyUserBinary")
	s.stage.BroadcastMHF(&mhfpacket.MsgSysInsertUser{
		CharID: s.charID,
	}, s)

	s.stage.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.charID,
		BinaryType: 1,
	}, s)
	s.stage.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.charID,
		BinaryType: 2,
	}, s)
	s.stage.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.charID,
		BinaryType: 3,
	}, s)

	// Notify the entree client about all of the existing clients in the stage.
	s.logger.Info("Notifying entree about existing stage clients")
	s.stage.RLock()
	clientNotif := byteframe.NewByteFrame()
	for session := range s.stage.clients {
		var cur mhfpacket.MHFPacket
		cur = &mhfpacket.MsgSysInsertUser{
			CharID: session.charID,
		}
		clientNotif.WriteUint16(uint16(cur.Opcode()))
		cur.Build(clientNotif)

		cur = &mhfpacket.MsgSysNotifyUserBinary{
			CharID:     session.charID,
			BinaryType: 1,
		}
		clientNotif.WriteUint16(uint16(cur.Opcode()))
		cur.Build(clientNotif)

		cur = &mhfpacket.MsgSysNotifyUserBinary{
			CharID:     session.charID,
			BinaryType: 2,
		}
		clientNotif.WriteUint16(uint16(cur.Opcode()))
		cur.Build(clientNotif)

		cur = &mhfpacket.MsgSysNotifyUserBinary{
			CharID:     session.charID,
			BinaryType: 3,
		}
		clientNotif.WriteUint16(uint16(cur.Opcode()))
		cur.Build(clientNotif)
	}
	s.stage.RUnlock()
	clientNotif.WriteUint16(0x0010) // End it.
	s.QueueSend(clientNotif.Data())

	// Notify the client to duplicate the existing objects.
	s.logger.Info("Notifying entree about existing stage objects")
	clientDupObjNotif := byteframe.NewByteFrame()
	s.stage.RLock()
	for _, obj := range s.stage.objects {
		cur := &mhfpacket.MsgSysDuplicateObject{
			ObjID:       obj.id,
			X:           obj.x,
			Y:           obj.y,
			Z:           obj.z,
			Unk0:        0,
			OwnerCharID: obj.ownerCharID,
		}
		clientDupObjNotif.WriteUint16(uint16(cur.Opcode()))
		cur.Build(clientDupObjNotif)
	}
	s.stage.RUnlock()
	clientDupObjNotif.WriteUint16(0x0010) // End it.
	s.QueueSend(clientDupObjNotif.Data())
}

func handleMsgSysBackStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysMoveStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLeaveStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLockStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLockStage)
	// TODO(Andoryuuta): What does this packet _actually_ do?
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysUnlockStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserveStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserveStage)

	fmt.Printf("Got reserve stage req, Unk0:%v, StageID:%q\n", pkt.Unk0, pkt.StageID)

	// TODO(Andoryuuta): Add proper player-slot reservations for stages.

	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysUnreserveStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetStagePass(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysWaitStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysWaitStageBinary)
	defer s.logger.Debug("MsgSysWaitStageBinary Done!")

	// Try to get the stage
	stageID := stripNullTerminator(pkt.StageID)
	s.server.stagesLock.Lock()
	stage, gotStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	// If we got the stage, lock and try to get the data.
	var stageBinary []byte
	var gotBinary bool
	if gotStage {
		for {
			stage.Lock()
			stageBinary, gotBinary = stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]
			stage.Unlock()

			if gotBinary {
				doSizedAckResp(s, pkt.AckHandle, stageBinary)
				break
			} else {
				// Couldn't get binary, sleep for some time and try again.
				time.Sleep(2 * time.Second)
				continue
			}

			// TODO(Andoryuuta): Figure out what the game sends on timeout and implement it!
			/*
				if timeout {
					s.logger.Warn("Failed to get stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
					s.logger.Warn("Sending blank stage binary")
					doSizedAckResp(s, pkt.AckHandle, []byte{})
					return
				}
			*/
		}
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", stageID))
	}
}

func handleMsgSysSetStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStageBinary)

	// Try to get the stage
	stageID := stripNullTerminator(pkt.StageID)
	s.server.stagesLock.Lock()
	stage, gotStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	// If we got the stage, lock and set the data.
	if gotStage {
		stage.Lock()
		stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}] = pkt.RawDataPayload
		stage.Unlock()
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", stageID))
	}
	s.logger.Debug("handleMsgSysSetStageBinary Done!")
}

func handleMsgSysGetStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetStageBinary)

	// Try to get the stage
	stageID := stripNullTerminator(pkt.StageID)
	s.server.stagesLock.Lock()
	stage, gotStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	// If we got the stage, lock and try to get the data.
	var stageBinary []byte
	var gotBinary bool
	if gotStage {
		stage.Lock()
		stageBinary, gotBinary = stage.rawBinaryData[stageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]
		stage.Unlock()
	} else {
		s.logger.Warn("Failed to get stage", zap.String("StageID", stageID))
	}

	if gotBinary {
		doSizedAckResp(s, pkt.AckHandle, stageBinary)
	} else {
		s.logger.Warn("Failed to get stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
		s.logger.Warn("Sending blank stage binary")
		doSizedAckResp(s, pkt.AckHandle, []byte{})
	}

	s.logger.Debug("MsgSysGetStageBinary Done!")
}

func handleMsgSysEnumerateClient(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateClient)

	// Read-lock the stages map.
	s.server.stagesLock.RLock()

	stage, ok := s.server.stages[stripNullTerminator(pkt.StageID)]
	if !ok {
		s.logger.Fatal("Can't enumerate clients for stage that doesn't exist!", zap.String("stageID", pkt.StageID))
	}

	// Unlock the stages map.
	s.server.stagesLock.RUnlock()

	// Read-lock the stage and make the response with all of the charID's in the stage.
	resp := byteframe.NewByteFrame()
	stage.RLock()

	// TODO(Andoryuuta): Add proper player-slot reservations for stages.
	if len(stage.clients) >= 1 {
		resp.WriteUint16(uint16(len(stage.clients))) // Client count
		for session := range stage.clients {
			resp.WriteUint32(session.charID) // Client represented by charID
		}
	} else {
		// Just give our client.
		resp.WriteUint16(1)
		resp.WriteUint32(s.charID)
	}

	stage.RUnlock()

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
	s.logger.Debug("MsgSysEnumerateClient Done!")
}

func handleMsgSysEnumerateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateStage)

	// Read-lock the stages.
	s.server.stagesLock.RLock()
	defer s.server.stagesLock.RUnlock()

	// Build the response
	resp := byteframe.NewByteFrame()
	resp.WriteUint16(uint16(len(s.server.stages)))
	for sid := range s.server.stages {
		// Found parsing code, field sizes are correct, but unknown purposes still.
		//resp.WriteBytes([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x04, 0x00})
		resp.WriteUint16(5)  // Current players.
		resp.WriteUint16(7)  // Unknown value
		resp.WriteUint16(0)  // HasDeparted or IsLocked.
		resp.WriteUint16(20) // Max players.
		resp.WriteUint8(2) // Password protected.
		resp.WriteUint8(uint8(len(sid)))
		resp.WriteBytes([]byte(sid))
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
	s.logger.Debug("handleMsgSysEnumerateStage Done!")
}

func handleMsgSysCreateMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCreateOpenMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysOpenMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCloseMutex(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCreateSemaphore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCreateAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteSemaphore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAcquireSemaphore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReleaseSemaphore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLockGlobalSema(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysUnlockGlobalSema(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCheckSemaphore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysOperateRegister(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLoadRegister(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysNotifyRegister(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCreateObject(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateObject)

	// Make sure we have a stage.
	if s.stage == nil {
		s.logger.Fatal("StageID not in the stages map!", zap.String("stageID", s.stageID))
	}

	// Lock the stage.
	s.stage.Lock()

	// Make a new stage object and insert it into the stage.
	objID := s.stage.gameObjectCount
	s.stage.gameObjectCount++

	newObj := &StageObject{
		id:          objID,
		ownerCharID: s.charID,
		x:           pkt.X,
		y:           pkt.Y,
		z:           pkt.Z,
	}

	s.stage.objects[objID] = newObj

	// Unlock the stage.
	s.stage.Unlock()

	// Response to our requesting client.
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)     // Unk, is this echoed back from pkt.Unk0?
	resp.WriteUint32(objID) // New local obj handle.
	s.QueueAck(pkt.AckHandle, resp.Data())

	// Duplicate the object creation to all sessions in the same stage.
	dupObjUpdate := &mhfpacket.MsgSysDuplicateObject{
		ObjID:       objID,
		X:           pkt.X,
		Y:           pkt.Y,
		Z:           pkt.Z,
		Unk0:        0,
		OwnerCharID: s.charID,
	}
	s.stage.BroadcastMHF(dupObjUpdate, s)
}

func handleMsgSysDeleteObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysPositionObject(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysPositionObject)
	fmt.Printf("Moved object %v to (%f,%f,%f)\n", pkt.ObjID, pkt.X, pkt.Y, pkt.Z)

	// One of the few packets we can just re-broadcast directly.
	s.stage.BroadcastMHF(pkt, s)
}

func handleMsgSysRotateObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDuplicateObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetObjectBinary(s *Session, p mhfpacket.MHFPacket) {

}

func handleMsgSysGetObjectBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysGetObjectOwner(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysUpdateObjectBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCleanupObject(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4A(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4B(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4C(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4D(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve4F(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysInsertUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysDeleteUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetUserBinary)
	s.server.userBinaryPartsLock.Lock()
	s.server.userBinaryParts[userBinaryPartID{charID: s.charID, index: pkt.BinaryType}] = pkt.RawDataPayload
	s.server.userBinaryPartsLock.Unlock()
}

func handleMsgSysGetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetUserBinary)

	// Try to get the data.
	s.server.userBinaryPartsLock.RLock()
	defer s.server.userBinaryPartsLock.RUnlock()
	data, ok := s.server.userBinaryParts[userBinaryPartID{charID: pkt.CharID, index: pkt.BinaryType}]

	resp := byteframe.NewByteFrame()

	// If we can't get the real data, use a placeholder.
	if !ok {
		if pkt.BinaryType == 1 {
			// Stub name response with character ID
			resp.WriteBytes([]byte(fmt.Sprintf("CID%d", s.charID)))
			resp.WriteUint8(0) // NULL terminator.
		} else if pkt.BinaryType == 2 {
			data, err := base64.StdEncoding.DecodeString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABBn8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAwAAAAAAAAAAAAAABAAAAAAAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==")
			if err != nil {
				panic(err)
			}
			resp.WriteBytes(data)
		} else if pkt.BinaryType == 3 {
			data, err := base64.StdEncoding.DecodeString("AQAAA2ea5P8ATgEA/wEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABBn8AAAAAAAAAAAABAKAMAAAAAAAAAAAAACgAAAAAAAAAAAABAsQOAAAAAAAAAAABA6UMAAAAAAAAAAABBKAMAAAAAAAAAAABBToNAAAAAAAAAAAAAAMAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAgACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD/////AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			if err != nil {
				panic(err)
			}
			resp.WriteBytes(data)
		}
	} else {
		resp.WriteBytes(data)
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysNotifyUserBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve55(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve56(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve57(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysUpdateRight(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAuthQuery(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAuthData(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAuthTerminal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve5C(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysRightsReload(s *Session, p mhfpacket.MHFPacket) {

}

func handleMsgSysReserve5E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve5F(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSavedata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavedata)
	err := ioutil.WriteFile(fmt.Sprintf("savedata\\%d.bin", time.Now().Unix()), pkt.RawDataPayload, 0644)
	if err != nil {
		s.logger.Fatal("Error dumping savedata", zap.Error(err))
	}

	_, err = s.server.db.Exec("UPDATE characters SET is_new_character=false, savedata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update savedata in db", zap.Error(err))
	}

	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoaddata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoaddata)

	var data []byte
	err := s.server.db.QueryRow("SELECT savedata FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get savedata from db", zap.Error(err))
	}

	doSizedAckResp(s, pkt.AckHandle, data)
}

func handleMsgMhfListMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfListMember)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Members count. (Unsure of what kind of members these actually are, guild, party, COG subscribers, etc.)

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfOprMember(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateDistItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfApplyDistItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireDistItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetDistDescription(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSendMail(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReadMail(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfListMail(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOprtMail(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadFavoriteQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadFavoriteQuest)
	// TODO(Andoryuuta): Save data from MsgMhfSaveFavoriteQuest and resend it here.
	doSizedAckResp(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfSaveFavoriteQuest(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfRegisterEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReleaseEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfTransitMessage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve71(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve72(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve73(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve74(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve75(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve76(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve77(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve78(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve79(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7A(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7B(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7C(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgCaExchangeItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve7E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPresentBox(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfServerCommand(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfShutClient(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAnnounce(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetLoginwindow(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysTransBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCollectBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysGetState(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSerialize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEnumlobby(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEnumuser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysInfokyserver(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCaUniqueID(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetCaAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCaravanMyScore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCaravanRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCaravanMyRank(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCreateGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateGuildMember(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfInfoGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoGuild)

	// REALLY large/complex format... stubbing it out here for simplicity.
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Count
	resp.WriteUint8(0)  // Unk, read if count == 0.

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfArrangeGuildMember(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateGuildMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildMember)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfEnumerateCampaign(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfStateCampaign(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfApplyCampaign(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfTransferItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfMercenaryHuntdata(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEntryRookieGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateQuest)
	stubEnumerateNoResults(s, pkt.AckHandle)

	// Update the client's rights as well:
	updateRights(s)
}

func handleMsgMhfEnumerateEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateEvent)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfEnumeratePrice(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumeratePrice)
	resp := byteframe.NewByteFrame()
	resp.WriteUint16(0) // Entry type 1 count
	resp.WriteUint16(0) // Entry type 2 count

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRanking)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint8(0)
	resp.WriteUint8(0)  // Some string length following this field.
	resp.WriteUint16(0) // Entry type 1 count
	resp.WriteUint8(0)  // Entry type 2 count

	doSizedAckResp(s, pkt.AckHandle, resp.Data())

	// Update the client's rights as well:
	updateRights(s)
}

func handleMsgMhfEnumerateOrder(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateOrder)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfEnumerateShop(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateShop)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfGetExtraInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateInterior(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateHouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateHouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadHouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateWarehouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateWarehouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateWarehouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireTitle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateTitle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateGuildItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateGuildItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateUnionItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateUnionItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCreateJoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateJoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfInfoJoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateGuildIcon(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfInfoFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoFesta)

	// REALLY large/complex format... stubbing it out here for simplicity.
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEntryFesta(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfChargeFesta(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireFesta(s *Session, p mhfpacket.MHFPacket) {}

// state festa (U)ser
func handleMsgMhfStateFestaU(s *Session, p mhfpacket.MHFPacket) {}

// state festa (G)uild
func handleMsgMhfStateFestaG(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaG)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0xFFFFFFFF)
	resp.WriteUint32(0)
	resp.WriteBytes([]byte{0x00, 0x00, 0x00}) // Not parsed.
	resp.WriteUint8(0)

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateFestaMember(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfVoteFesta(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireCafeItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateCafepoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCheckDailyCafepoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCogInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCheckMonthlyItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireMonthlyItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCheckWeeklyStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckWeeklyStamp)

	resp := byteframe.NewByteFrame()
	resp.WriteUint16(0x0100)
	resp.WriteUint16(0x000E)
	resp.WriteUint16(0x0001)
	resp.WriteUint16(0x0000)
	resp.WriteUint16(0x0001)
	resp.WriteUint32(0)
	resp.WriteUint32(0x5dddcbb3) // Timestamp

	s.QueueAck(pkt.AckHandle, resp.Data())
}

func handleMsgMhfExchangeWeeklyStamp(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCreateMercenary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSaveMercenary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReadMercenaryW(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryW)

	// Unk format:
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadMercenaryM(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfContractMercenary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateMercenaryLog(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateGuacot(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateGuacot(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfInfoTournament(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEntryTournament(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnterTournamentQuest(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireTournament(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfResetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPaymentAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfDisplayedAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfInfoScenarioCounter(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSaveScenarioData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveScenarioData)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40})
}

func handleMsgMhfLoadScenarioData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadScenarioData)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetBbsSnsStatus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfApplyBbsArticle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEtcPoints(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEtcPoints)

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(0x3) // Maybe a count of uint32(s)?
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateEtcPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetWeeklySchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySchedule)

	eventSchedules := []struct {
		StartTime time.Time
		Unk0      uint32 // Event ID?
		Unk1      uint16
	}{
		{
			StartTime: time.Now().Add(time.Duration(-5) * time.Minute), // Event started 5 minutes ago.
			Unk0:      4,
			Unk1:      0,
		},
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(eventSchedules))) // Entry count, client only parses the first 7 or 8.
	resp.WriteUint32(uint32(time.Now().Unix())) // Current server time
	for _, es := range eventSchedules {
		resp.WriteUint32(uint32(es.StartTime.Unix()))
		resp.WriteUint32(es.Unk0)
		resp.WriteUint16(es.Unk1)
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateInvGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperationInvGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfStampcardStamp(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfStampcardPrize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUnreserveSrg(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadPlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateData)

	// TODO(Andoryuuta): Save data from MsgMhfSavePlateData and resend it here.
	doSizedAckResp(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfSavePlateData(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadPlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateBox)
	// TODO(Andoryuuta): Save data from MsgMhfSavePlateBox and resend it here.
	doSizedAckResp(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfSavePlateBox(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReadGuildcard(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadGuildcard)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateGuildcard(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReadBeatLevel(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevel)

	// This response is fixed and will never change on JP,
	// but I've left it dynamic for possible other client differences.
	resp := byteframe.NewByteFrame()
	for i := 0; i < int(pkt.ValidIDCount); i++ {
		resp.WriteUint32(pkt.IDs[i])
		resp.WriteUint32(1)
		resp.WriteUint32(1)
		resp.WriteUint32(1)
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateBeatLevel(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReadBeatLevelAllRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReadBeatLevelMyRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReadLastWeekBeatRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcceptReadReward(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetAdditionalBeatReward(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetFixedSeibatuRankingTable(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBbsUserStatus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfKickExportForce(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBreakSeibatuLevelReward(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetWeeklySeibatuRankingReward(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEarthStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthStatus)

	// TODO(Andoryuuta): Track down format for this data,
	//	it can somehow be parsed as 8*uint32 chunks if the header is right.
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	s.QueueAck(pkt.AckHandle, resp.Data())
}

func handleMsgMhfLoadPartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPartner)

	// TODO(Andoryuuta): Figure out unusual double ack. One sized, one not.

	// TODO(Andoryuuta): Save data from MsgMhfSavePartner and resend it here.
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfSavePartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePartner)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetGuildMissionList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildMissionRecord(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddGuildMissionCount(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCancelGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadOtomoAirou)

	// TODO(Andoryuuta): Save data from MsgMhfSaveOtomoAirou and resend it here.
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfSaveOtomoAirou(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateGuildTresure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateAiroulist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfRegistGuildTresure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireGuildTresure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfOperateGuildTresureReport(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadDecoMyset)

	// TODO(Andoryuuta): Save data from MsgMhfSaveDecoMyset and resend it here.
	doSizedAckResp(s, pkt.AckHandle, []byte{0x01, 0x00})
}

func handleMsgMhfSaveDecoMyset(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReserve010F(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadGuildCooking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfRegistGuildCooking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadGuildAdventure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfRegistGuildAdventure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireGuildAdventure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfChargeGuildAdventure(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadLegendDispatch(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHunterNavi)
	// TODO(Andoryuuta): Save data from MsgMhfSaveHunterNavi and resend it here.
	blankData := make([]byte, 0x228)
	doSizedAckResp(s, pkt.AckHandle, blankData)
}

func handleMsgMhfSaveHunterNavi(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfRegistSpabiTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildWeeklyBonusMaster(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildWeeklyBonusActiveCount(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddGuildWeeklyBonusExceptionalUser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTowerInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTowerInfo)

	/*
		type:
		1 == TOWER_RANK_POINT,
		2 == GET_OWN_TOWER_SKILL
		3 == ?
		4 == TOWER_TOUHA_HISTORY
		5 = ?

		[] = type
		req
		resp

		01 1d 01 fc 00 09 [00 00 00 01] 00 00 00 02 00 00 00 00
		00 12 01 fc 00 09 01 00 00 18 0a 21 8e ad 00 00 00 00 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00 00

		01 1d 01 fc 00 0a [00 00 00 02] 00 00 00 00 00 00 00 00
		00 12 01 fc 00 0a 01 00 00 94 0a 21 8e ad 00 00 00 00 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00

		01 1d 01 ff 00 0f [00 00 00 04] 00 00 00 00 00 00 00 00
		00 12 01 ff 00 0f 01 00 00 24 0a 21 8e ad 00 00 00 00 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00

		01 1d 01 fc 00 0b [00 00 00 05] 00 00 00 00 00 00 00 00
		00 12 01 fc 00 0b 01 00 00 10 0a 21 8e ad 00 00 00 00 00 00 00 00 00 00 00 00
	*/
	/*
		switch pkt.InfoType {
		case mhfpacket.TowerInfoTypeTowerRankPoint:
		case mhfpacket.TowerInfoTypeGetOwnTowerSkill:
		case mhfpacket.TowerInfoTypeUnk3:
			panic("No known response values for TowerInfoTypeUnk3")
		case mhfpacket.TowerInfoTypeTowerTouhaHistory:
		case mhfpacket.TowerInfoTypeUnk5:
		}
	*/

	stubGetNoResults(s, pkt.AckHandle)
}

func handleMsgMhfPostTowerInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGemInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostGemInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEarthValue(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthValue)

	earthValues := []struct {
		Unk0, Unk1, Unk2, Unk3, Unk4, Unk5 uint32
	}{
		{
			Unk0: 0x03E9,
			Unk1: 0x5B,
		},
		{
			Unk0: 0x2329,
			Unk1: 0x03,
		},
		{
			Unk0: 0x232A,
			Unk1: 0x0A,
			Unk2: 0x012C,
		},
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0x0A218EAD)               // Unk shared ID. Sent in response of MSG_MHF_GET_TOWER_INFO, MSG_MHF_GET_PAPER_DATA etc.
	resp.WriteUint32(0)                        // Unk
	resp.WriteUint32(0)                        // Unk
	resp.WriteUint32(uint32(len(earthValues))) // value count
	for _, v := range earthValues {
		resp.WriteUint32(v.Unk0)
		resp.WriteUint32(v.Unk1)
		resp.WriteUint32(v.Unk2)
		resp.WriteUint32(v.Unk3)
		resp.WriteUint32(v.Unk4)
		resp.WriteUint32(v.Unk5)
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfDebugPostValue(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetPaperData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetPaperData)
	stubGetNoResults(s, pkt.AckHandle)
}

func handleMsgMhfGetNotice(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostNotice(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTime)
	doSizedAckResp(s, pkt.AckHandle, []byte{})

	// Update the client's rights as well:
	updateRights(s)
}

func handleMsgMhfPostBoostTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTimeLimit)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfPostBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRandFromTable(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCafeDuration(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCafeDurationBonusInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReceiveCafeDurationBonus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostCafeDurationBonusReceived(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGachaPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGachaPoint)
	doSizedAckResp(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfUseGachaPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfExchangeFpoint2Item(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfExchangeItem2Fpoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetFpointExchangeList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPlayStepupGacha(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReceiveGachaItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetStepupStatus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPlayFreeGacha(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTinyBin(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostTinyBin(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetSenyuDailyCount(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildTargetMemberNum(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBoostRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostRight)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfStartBoostTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostBoostTimeQuestReturn(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPlayBoxGacha(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfResetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetSeibattle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostSeibattle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRyoudama(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostRyoudama(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTenrouirai(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostTenrouirai(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostGuildScout(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCancelGuildScout(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAnswerGuildScout(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildScoutList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildManageRight(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetGuildManageRight(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPlayNormalGacha(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetDailyMissionMaster(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGachaPlayHistory(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRejectGuildScout(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetRejectGuildScout(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetKeepLoginBoostStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKeepLoginBoostStatus)

	unkRespFields := [5]struct {
		U0, U1, U2 uint8
		U3         uint32
	}{
		{
			U0: 1,
			U1: 1,
			U2: 1,
			U3: 0,
		},
		{
			U0: 2,
			U1: 0,
			U2: 1,
			U3: 0,
		},
		{
			U0: 3,
			U1: 0,
			U2: 1,
			U3: 0,
		},
		{
			U0: 4,
			U1: 0,
			U2: 1,
			U3: 0,
		},
		{
			U0: 5,
			U1: 0,
			U2: 1,
			U3: 0,
		},
	}

	resp := byteframe.NewByteFrame()
	for _, v := range unkRespFields {
		resp.WriteUint8(v.U0)
		resp.WriteUint8(v.U1)
		resp.WriteUint8(v.U2)
		resp.WriteUint32(v.U3)
	}
	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUseKeepLoginBoost(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdSchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdSchedule)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0x0b5397df) // Unk
	resp.WriteUint32(0x5ddde6b0) // Timestamp
	resp.WriteUint32(0x5de71320) // Timestamp
	resp.WriteUint32(0x5de7225c) // Timestamp
	resp.WriteUint32(0x5df04da0) // Timestamp
	resp.WriteUint32(0x5df05cdc) // Timestamp
	resp.WriteUint32(0x5dfa30e0) // Timestamp
	resp.WriteUint16(0x19)       // Unk
	resp.WriteUint16(0x2d)       // Unk
	resp.WriteUint16(0x02)       // Unk
	resp.WriteUint16(0x02)       // Unk

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetUdInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdInfo)

	udInfos := []struct {
		Text      string
		StartTime time.Time
		EndTime   time.Time
	}{
		{
			Text:      " ~C17【Erupe】 launch event!\n\n■Features\n~C18 Walk around!\n~C17 Crash your connection by doing nearly anything!",
			StartTime: time.Now().Add(time.Duration(-5) * time.Minute), // Event started 5 minutes ago,
			EndTime:   time.Now().Add(time.Duration(5) * time.Minute),  // Event ends in 5 minutes,
		},
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(udInfos)))
	for _, udInfo := range udInfos {
		resp.WriteBytes(fixedSizeShiftJIS(udInfo.Text, 1024))
		resp.WriteUint32(uint32(udInfo.StartTime.Unix()))
		resp.WriteUint32(uint32(udInfo.EndTime.Unix()))
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetKijuInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetKiju(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddUdPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdMyPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTotalPointInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdBonusQuestInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdSelectedColorInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdMonsterPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMonsterPoint)

	monsterPoints := []struct {
		MID    uint8 // Monster ID ?
		Points uint16
	}{
		{MID: 0x01, Points: 0x3C},
		{MID: 0x02, Points: 0x5A},
		{MID: 0x06, Points: 0x14},
		{MID: 0x07, Points: 0x50},
		{MID: 0x08, Points: 0x28},
		{MID: 0x0B, Points: 0x3C},
		{MID: 0x0E, Points: 0x3C},
		{MID: 0x0F, Points: 0x46},
		{MID: 0x11, Points: 0x46},
		{MID: 0x14, Points: 0x28},
		{MID: 0x15, Points: 0x3C},
		{MID: 0x16, Points: 0x32},
		{MID: 0x1A, Points: 0x32},
		{MID: 0x1B, Points: 0x0A},
		{MID: 0x1C, Points: 0x0A},
		{MID: 0x1F, Points: 0x0A},
		{MID: 0x21, Points: 0x50},
		{MID: 0x24, Points: 0x64},
		{MID: 0x25, Points: 0x3C},
		{MID: 0x26, Points: 0x1E},
		{MID: 0x27, Points: 0x28},
		{MID: 0x28, Points: 0x50},
		{MID: 0x29, Points: 0x5A},
		{MID: 0x2A, Points: 0x50},
		{MID: 0x2B, Points: 0x3C},
		{MID: 0x2C, Points: 0x3C},
		{MID: 0x2D, Points: 0x46},
		{MID: 0x2E, Points: 0x3C},
		{MID: 0x2F, Points: 0x50},
		{MID: 0x30, Points: 0x1E},
		{MID: 0x31, Points: 0x3C},
		{MID: 0x32, Points: 0x50},
		{MID: 0x33, Points: 0x3C},
		{MID: 0x34, Points: 0x28},
		{MID: 0x35, Points: 0x50},
		{MID: 0x36, Points: 0x6E},
		{MID: 0x37, Points: 0x50},
		{MID: 0x3A, Points: 0x50},
		{MID: 0x3B, Points: 0x6E},
		{MID: 0x40, Points: 0x64},
		{MID: 0x41, Points: 0x6E},
		{MID: 0x43, Points: 0x28},
		{MID: 0x44, Points: 0x0A},
		{MID: 0x47, Points: 0x6E},
		{MID: 0x4A, Points: 0xFA},
		{MID: 0x4B, Points: 0xFA},
		{MID: 0x4C, Points: 0x46},
		{MID: 0x4D, Points: 0x64},
		{MID: 0x4E, Points: 0xFA},
		{MID: 0x4F, Points: 0xFA},
		{MID: 0x50, Points: 0xFA},
		{MID: 0x51, Points: 0xFA},
		{MID: 0x52, Points: 0xFA},
		{MID: 0x53, Points: 0xFA},
		{MID: 0x54, Points: 0xFA},
		{MID: 0x55, Points: 0xFA},
		{MID: 0x59, Points: 0xFA},
		{MID: 0x5A, Points: 0xFA},
		{MID: 0x5B, Points: 0xFA},
		{MID: 0x5C, Points: 0xFA},
		{MID: 0x5E, Points: 0xFA},
		{MID: 0x5F, Points: 0xFA},
		{MID: 0x60, Points: 0xFA},
		{MID: 0x63, Points: 0xFA},
		{MID: 0x65, Points: 0xFA},
		{MID: 0x67, Points: 0xFA},
		{MID: 0x68, Points: 0xFA},
		{MID: 0x69, Points: 0xFA},
		{MID: 0x6A, Points: 0xFA},
		{MID: 0x6B, Points: 0xFA},
		{MID: 0x6C, Points: 0xFA},
		{MID: 0x6D, Points: 0xFA},
		{MID: 0x6E, Points: 0xFA},
		{MID: 0x6F, Points: 0xFA},
		{MID: 0x70, Points: 0xFA},
		{MID: 0x72, Points: 0xFA},
		{MID: 0x73, Points: 0xFA},
		{MID: 0x74, Points: 0xFA},
		{MID: 0x77, Points: 0xFA},
		{MID: 0x78, Points: 0xFA},
		{MID: 0x79, Points: 0xFA},
		{MID: 0x7A, Points: 0xFA},
		{MID: 0x7B, Points: 0xFA},
		{MID: 0x7D, Points: 0xFA},
		{MID: 0x7E, Points: 0xFA},
		{MID: 0x7F, Points: 0xFA},
		{MID: 0x80, Points: 0xFA},
		{MID: 0x81, Points: 0xFA},
		{MID: 0x82, Points: 0xFA},
		{MID: 0x83, Points: 0xFA},
		{MID: 0x8B, Points: 0xFA},
		{MID: 0x8C, Points: 0xFA},
		{MID: 0x8D, Points: 0xFA},
		{MID: 0x8E, Points: 0xFA},
		{MID: 0x90, Points: 0xFA},
		{MID: 0x92, Points: 0x78},
		{MID: 0x93, Points: 0x78},
		{MID: 0x94, Points: 0x78},
		{MID: 0x96, Points: 0xFA},
		{MID: 0x97, Points: 0x78},
		{MID: 0x98, Points: 0x78},
		{MID: 0x99, Points: 0x78},
		{MID: 0x9A, Points: 0xFA},
		{MID: 0x9E, Points: 0xFA},
		{MID: 0x9F, Points: 0x78},
		{MID: 0xA0, Points: 0xFA},
		{MID: 0xA1, Points: 0xFA},
		{MID: 0xA2, Points: 0x78},
		{MID: 0xA4, Points: 0x78},
		{MID: 0xA5, Points: 0x78},
		{MID: 0xA6, Points: 0xFA},
		{MID: 0xA9, Points: 0x78},
		{MID: 0xAA, Points: 0xFA},
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(monsterPoints)))
	for _, mp := range monsterPoints {
		resp.WriteUint8(mp.MID)
		resp.WriteUint16(mp.Points)
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetUdDailyPresentList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdNormaPresentList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdRankingRewardList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireUdItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRewardSong(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUseRewardSong(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddRewardSongCount(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdMyRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireMonthlyReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireMonthlyReward)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetUdGuildMapInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGenerateUdGuildMap(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddUdTacticsPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsRewardList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsLog(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsFollower)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfSetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdShopCoin(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUseUdShopCoin(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEnhancedMinidata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEnhancedMinidata)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00})
}

func handleMsgMhfSetEnhancedMinidata(s *Session, p mhfpacket.MHFPacket) {

}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetLobbyCrowd(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve180(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGuildHuntdata(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddKouryouPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKouryouPoint)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfExchangeKouryouPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsBonusQuest(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsFirstQuestBonus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsRemainingPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve188(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserve188)

	// Left as raw bytes because I couldn't easily find the request or resp parser function in the binary.
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateMyset)
	// TODO(Andoryuuta): Save data from MsgMhfSavePlateMyset and resend it here.
	blankData := make([]byte, 0x780)
	doSizedAckResp(s, pkt.AckHandle, blankData)
}

func handleMsgMhfSavePlateMyset(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve18B(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserve18B)

	// Left as raw bytes because I couldn't easily find the request or resp parser function in the binary.
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x3C})

}

func handleMsgMhfGetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve18E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve18F(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTrendWeapon(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateUseTrendWeaponLog(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve192(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve193(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve194(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSaveRengokuData(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadRengokuData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadRengokuData)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint16(0)
	resp.WriteUint32(0)
	resp.WriteUint16(0)
	resp.WriteUint16(0)
	resp.WriteUint32(0)

	resp.WriteUint8(3) // Count of next 3
	resp.WriteUint16(0)
	resp.WriteUint16(0)
	resp.WriteUint16(0)

	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	resp.WriteUint8(3) // Count of next 3
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	resp.WriteUint8(3) // Count of next 3
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	doSizedAckResp(s, pkt.AckHandle, resp.Data())

}

func handleMsgMhfGetRengokuBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateRengokuRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRengokuRankingRank(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRengokuRankingRank)

	resp := byteframe.NewByteFrame()
	resp.WriteBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfAcquireExchangeShop(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve19B(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSaveMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMezfesData)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadMezfesData)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Unk

	resp.WriteUint8(2) // Count of the next 2 uint32s
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	resp.WriteUint32(0) // Unk

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysReserve19E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve19F(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateForceGuildRank(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfResetTitle(s *Session, p mhfpacket.MHFPacket) {}

// "Enumrate_guild_msg_board"
func handleMsgSysReserve202(s *Session, p mhfpacket.MHFPacket) {
}

// "Is_update_guild_msg_board"
func handleMsgSysReserve203(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserve203)
	resp := make([]byte, 8) // Unk resp.
	s.QueueAck(pkt.AckHandle, resp)
}

func handleMsgSysReserve204(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve205(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve206(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve207(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve208(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve209(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve20A(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve20B(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve20C(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve20D(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve20E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve20F(s *Session, p mhfpacket.MHFPacket) {}
