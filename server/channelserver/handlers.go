package channelserver

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

// process a datadiff save for platebox and platedata
func saveDataDiff(b []byte, save []byte) []byte {
	// there are a bunch of extra variations on this method in use which this does not handle yet
	// specifically this is for diffs with seek amounts trailed by 02 followed by bytes to be written
	var seekBytes []byte
	seekOperation := 0
	write := byte(0)
	for len(b) > 2 {
		if bytes.IndexRune(b, 2) != 0 {
			seekBytes = b[:bytes.IndexRune(b, 2)+1]
		} else {
			seekBytes = b[:bytes.IndexRune(b[1:], 2)+2]
		}
		if len(seekBytes) == 1 {
			seekBytes = b[:bytes.IndexRune(b, 2)+2]
			//fmt.Printf("Seek: %d SeekBytes: %X Write: %X\n", seekBytes[0], seekBytes, b[len(seekBytes)] )
			seekOperation += int(seekBytes[0])
			write = b[len(seekBytes)]
			b = b[3:]
		} else {
			seek := int32(0)
			for _, b := range seekBytes[:len(seekBytes)-1] {
				seek = (seek << 8) | int32(b)
			}
			//fmt.Printf("Seek: %d SeekBytes: %X Write: %X\n", seek, seekBytes, b[len(seekBytes)] )
			seekOperation += int(seek)
			write = b[len(seekBytes)]
			b = b[len(seekBytes)+1:]
		}
		save[seekOperation-1] = write
	}

	return save
}

// decompress save data
func saveDecompress(compData []byte) ([]byte, error) {
	r := bytes.NewReader(compData)

	header := make([]byte, 16)
	n, err := r.Read(header)
	if err != nil {
		return nil, err
	} else if n != len(header) {
		return nil, err
	}

	if !bytes.Equal(header, []byte("cmp\x2020110113\x20\x20\x20\x00")) {
		return nil, err
	}

	var output []byte
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if b == 0 {
			// If it's a null byte, then the next byte is how many nulls to add.
			nullCount, err := r.ReadByte()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			output = append(output, make([]byte, int(nullCount))...)
		} else {
			output = append(output, b)
		}
	}

	return output, nil
}

// Null compresses a save
func saveCompress(rawData []byte) ([]byte, error) {
	r := bytes.NewReader(rawData)
	var output []byte
	output = append(output, []byte("cmp\x2020110113\x20\x20\x20\x00")...)
	for {
		b, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if b == 0 {
			output = append(output, []byte{0x00}...)
			// read to get null count
			nullCount := 1
			for {
				i, err := r.ReadByte()
				if err == io.EOF {
					output = append(output, []byte{byte(nullCount)}...)
					break
				} else if i != 0 {
					r.UnreadByte()
					output = append(output, []byte{byte(nullCount)}...)
					break
				} else if err != nil {
					return nil, err
				}
				nullCount++

				if nullCount == 255 {
					output = append(output, []byte{0xFF, 0x00}...)
					nullCount = 0
				}
			}
			//output = append(output, []byte{byte(nullCount)}...)
		} else {
			output = append(output, b)
		}
	}
	return output, nil
}

func updateRights(s *Session) {
	update := &mhfpacket.MsgSysUpdateRight{
		Unk0: 0,
		Unk1: 0x4E,
		// 01 = Character can take quests at allows
		// 02 = Hunter Life, normal quests core sub
		// 03 = Extra Course, extra quests, town boxes, QOL course, core sub
		// 06 = Premium Course, standard 'premium' which makes ranking etc. faster
		// some connection to unk1 above for these maybe?
		// 06 0A 0B = Boost Course, just actually 3 subs combined
		// 08 09 1E = N Course, gives you the benefits of being in a netcafe (extra quests, N Points, daily freebies etc.) minimal and pointless
		// no timestamp after 08 or 1E while active
		// 0C = N Boost course, ultra luxury course that ruins the game if in use but also gives a
		Rights: []mhfpacket.ClientRight{
			{
				ID:        1,
				Timestamp: 0,
			},
			{
				ID:        2,
				Timestamp: 0x5FEA1781,
			},
			{
				ID:        3,
				Timestamp: 0x5FEA1781,
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

	// Debug print the request.
	fmt.Printf("%+v\n", pkt)
	if pkt.IsScenario {
		fmt.Printf("%+v\n", pkt.ScenarioIdentifer)
	}

	if !pkt.IsScenario {
		if _, err := os.Stat(filepath.Join(s.server.erupeConfig.BinPath, "quest_override.bin")); err == nil {
			data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "quest_override.bin"))
			if err != nil {
				panic(err)
			}
			doSizedAckResp(s, pkt.AckHandle, data)
		} else {
			// Get quest file.
			data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("quests/%s.bin", stripNullTerminator(pkt.Filename))))
			if err != nil {
				panic(err)
			}
			doSizedAckResp(s, pkt.AckHandle, data)
		}
	} else {

		/*
			// mhf-fake-client format
			filename := fmt.Sprintf(
				"%d_%d_%d_%d",
				pkt.ScenarioIdentifer.CategoryID,
				pkt.ScenarioIdentifer.MainID,
				pkt.ScenarioIdentifer.ChapterID,
				pkt.ScenarioIdentifer.Flags,
			)
		*/

		// Fist's format:
		filename := fmt.Sprintf(
			"%d_0_0_0_S%d_T%d_C%d",
			pkt.ScenarioIdentifer.CategoryID,
			pkt.ScenarioIdentifer.MainID,
			pkt.ScenarioIdentifer.Flags, // Fist had as "type" and is the "T%d"
			pkt.ScenarioIdentifer.ChapterID,
		)

		// Read the scenario file.
		data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("scenarios/%s.bin", filename)))
		if err != nil {
			panic(err)
		}

		doSizedAckResp(s, pkt.AckHandle, data)
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

func handleMsgSysRecordLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysRecordLog)
	resp := make([]byte, 8) // Unk resp.
	s.QueueAck(pkt.AckHandle, resp)
}

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

func doStageTransfer(s *Session, ackHandle uint32, stageID string) {
	// Remove this session from old stage clients list and put myself in the new one.
	s.server.stagesLock.Lock()
	newStage, gotNewStage := s.server.stages[stripNullTerminator(stageID)]
	s.server.stagesLock.Unlock()

	if s.stage != nil {
		s.stage.Lock()

		// Remove client from old stage.
		delete(s.stage.clients, s)

		// Delete old stage objects owned by the client.
		s.logger.Info("Sending MsgSysDeleteObject to old stage clients")
		for objID, stageObject := range s.stage.objects {
			if stageObject.ownerCharID == s.charID {
				// Broadcast the deletion to clients in the stage.
				s.stage.BroadcastMHF(&mhfpacket.MsgSysDeleteObject{
					ObjID: stageObject.id,
				}, s)
				// TODO(Andoryuuta): Should this be sent to the owner's client as well? it currently isn't.

				// Actually delete it form the objects map.
				delete(s.stage.objects, objID)
			}
		}

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
	s.stageID = string(stripNullTerminator(stageID))
	s.stage = newStage
	s.Unlock()

	// Tell the client to cleanup its current stage objects.
	s.QueueSendMHF(&mhfpacket.MsgSysCleanupObject{})

	// Confirm the stage entry.
	s.QueueAck(ackHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

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

func handleMsgSysEnterStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnterStage)

	// Push our current stage ID to the movement stack before entering another one.
	s.Lock()
	s.stageMoveStack.Push(s.stageID)
	s.Unlock()

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysBackStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysBackStage)

	// Transfer back to the saved stage ID before the previous move or enter.
	s.Lock()
	backStage, err := s.stageMoveStack.Pop()
	s.Unlock()

	if err != nil {
		panic(err)
	}

	doStageTransfer(s, pkt.AckHandle, backStage)

}

func handleMsgSysMoveStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysMoveStage)

	// Push our current stage ID to the movement stack before entering another one.
	s.Lock()
	s.stageMoveStack.Push(s.stageID)
	s.Unlock()

	// just make everything use the town hub stage to get into zones for now
	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.FixedStageID {
		doStageTransfer(s, pkt.AckHandle, "sl1Ns200p0a0u0")
	} else {
		doStageTransfer(s, pkt.AckHandle, pkt.StageID)
	}
}

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
		resp.WriteUint16(1) // Current players.
		resp.WriteUint16(0) // Unknown value
		resp.WriteUint16(0) // HasDeparted.
		resp.WriteUint16(4) // Max players.
		resp.WriteUint8(2)  // Password protected.
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
	if pkt.SaveType == 2 {
		_, err = s.server.db.Exec("UPDATE characters SET is_new_character=false, savedata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)

		// Temporary server launcher response stuff
		// 0x1F715	Weapon Class
		// 0x1FDF6 HR (small_gr_level)
		// 0x88 Character Name
		saveFile, _ := saveDecompress(pkt.RawDataPayload)
		_, err = s.server.db.Exec("UPDATE characters SET weapon=$1 WHERE id=$2", uint16(saveFile[128789]), s.charID)
		x := uint16(saveFile[130550])<<8 | uint16(saveFile[130551])
		_, err = s.server.db.Exec("UPDATE characters SET small_gr_level=$1 WHERE id=$2", uint16(x), s.charID)
		_, err = s.server.db.Exec("UPDATE characters SET name=$1 WHERE id=$2", strings.SplitN(string(saveFile[88:100]), "\x00", 2)[0], s.charID)

		if err != nil {
			s.logger.Fatal("Failed to update savedata in db", zap.Error(err))
		}
	} else {
		fmt.Printf("Got savedata packet of type 1, no handling implemented. Not saving.")
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
	// Fist: Using a no favourites placeholder to avoid an in game error message
	// being sent every time you use a counter when it fails to load
	doSizedAckResp(s, pkt.AckHandle, []byte{0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

}

func handleMsgMhfSaveFavoriteQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveFavoriteQuest)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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

func handleMsgMhfTransferItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfTransferItem)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfMercenaryHuntdata(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEntryRookieGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateQuest)

	// questlists seem to be returned based on their internal values, intended order has these as:
	// 0 > 42 > 84 > 126 > 168, what actually makes it stop requesting quests I do not know yet
	// INSERT INTO questlists (ind, questlist) VALUES ('0', pg_read_binary_file('c:\save\quest_0_0.bin'));
	// INSERT INTO questlists (ind, questlist) VALUES ('42', pg_read_binary_file('c:\save\quest_42_2A.bin'));
	// INSERT INTO questlists (ind, questlist) VALUES ('84', pg_read_binary_file('c:\save\quest_84_54.bin'));
	// INSERT INTO questlists (ind, questlist) VALUES ('126', pg_read_binary_file('c:\save\quest_126_7E.bin'));
	// INSERT INTO questlists (ind, questlist) VALUES ('168', pg_read_binary_file('c:\save\quest_168_A8.bin'));
	var data []byte
	err := s.server.db.QueryRow("SELECT questlist FROM questlists WHERE ind = $1", int(pkt.QuestList)).Scan(&data)
	if err != nil {
		fmt.Println("Couldn't find quest list.")
		stubEnumerateNoResults(s, pkt.AckHandle)
	} else {
		if len(data) > 0 {
			doSizedAckResp(s, pkt.AckHandle, data)
		} else {
			stubEnumerateNoResults(s, pkt.AckHandle)
		}
	}
	// Update the client's rights as well:
	updateRights(s)
}

func handleMsgMhfEnumerateEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateEvent)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfEnumeratePrice(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumeratePrice)
	//resp := byteframe.NewByteFrame()
	//resp.WriteUint16(0) // Entry type 1 count
	//resp.WriteUint16(0) // Entry type 2 count
	// directly lifted for now because lacking it crashes the counter on having actual events present
	data, _ := hex.DecodeString("0000000066000003E800000000007300640100000320000000000006006401000003200000000000300064010000044C00000000007200640100000384000000000034006401000003840000000000140064010000051400000000006E006401000003E8000000000016006401000003E8000000000001006401000003200000000000430064010000057800000000006F006401000003840000000000330064010000044C00000000000B006401000003E800000000000F006401000006400000000000700064010000044C0000000000110064010000057800000000004C006401000003E8000000000059006401000006A400000000006D006401000005DC00000000004B006401000005DC000000000050006401000006400000000000350064010000070800000000006C0064010000044C000000000028006401000005DC00000000005300640100000640000000000060006401000005DC00000000005E0064010000051400000000007B006401000003E80000000000740064010000070800000000006B0064010000025800000000001B0064010000025800000000001C006401000002BC00000000001F006401000006A400000000007900640100000320000000000008006401000003E80000000000150064010000070800000000007A0064010000044C00000000000E00640100000640000000000055006401000007D0000000000002006401000005DC00000000002F0064010000064000000000002A0064010000076C00000000007E006401000002BC0000000000440064010000038400000000005C0064010000064000000000005B006401000006A400000000007D0064010000076C00000000007F006401000005DC0000000000540064010000064000000000002900640100000960000000000024006401000007D0000000000081006401000008340000000000800064010000038400000000001A006401000003E800000000002D0064010000038400000000004A006401000006A400000000005A00640100000384000000000027006401000007080000000000830064010000076C000000000040006401000006400000000000690064010000044C000000000025006401000004B000000000003100640100000708000000000082006401000003E800000000006500640100000640000000000051006401000007D000000000008C0064010000070800000000004D0064010000038400000000004E0064010000089800000000008B006401000004B000000000002E006401000009600000000000920064010000076C00000000008E00640100000514000000000068006401000004B000000000002B006401000003E800000000002C00640100000BB8000000000093006401000008FC00000000009000640100000AF0000000000094006401000006A400000000008D0064010000044C000000000052006401000005DC00000000004F006401000008980000000000970064010000070800000000006A0064010000064000000000005F00640100000384000000000026006401000008FC000000000096006401000007D00000000000980064010000076C000000000041006401000006A400000000003B006401000007080000000000360064010000083400000000009F00640100000A2800000000009A0064010000076C000000000021006401000007D000000000006300640100000A8C0000000000990064010000089800000000009E006401000007080000000000A100640100000C1C0000000000A200640100000C800000000000A400640100000DAC0000000000A600640100000C800000000000A50064010010")
	doSizedAckResp(s, pkt.AckHandle, data)
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

	if pkt.ShopType == 1 {
		stubEnumerateNoResults(s, pkt.AckHandle)
	} else if pkt.ShopType == 10 {
		// shops might be better off being in a persistent database as they changed semi-frequently
		// TODO (Fist): Identify what shops are actually on which IDs (GCP, N Point, Road, Festival, Diva Defense GCP, Zeny, Netcafe, Gacha Coins)
		if pkt.ShopID == 4 {
			data, _ := hex.DecodeString("000800082C93447100003B80000003E80001000000000001000000000000000100000C7BE2C200003A66000003E80001000000000001000000000000000100000B51B75F00003AF5000003E80001000000000001000000000000000B00003693367000003DF7000003E80001000000000001000000000000000B000019B22C8D00003DF8000003E80001000000000001000000000000000B00001B3F4BBE00003DF9000003E80001000000000001000000000000000B00001A91C77B00003DFA000003E80001000000000001000000000000000B00000A5FDB5C00003DFB000003E80001000000000001000000000000000B0000")
			doSizedAckResp(s, pkt.AckHandle, data)
		} else if pkt.ShopID == 6 {
			data, _ := hex.DecodeString("0008000807749AA900004027000000C800010000000000C8000000000000000B000031950D2A00004028000000DC00010000000000C8000000000000001400003697CA6700004029000000FA00010000000000C8000000000000001E0000005788980000402A0000012C00010000000000C8000000000000002800000C727D150000406E000000C800010000000000C8000000000000000B000024D1B6460000406F000000DC00010000000000C80000000000000014000020B429A300004070000000FA00010000000000C8000000000000001E00003FFDDD84000040710000012C00010000000000C800000000000000280000")
			doSizedAckResp(s, pkt.AckHandle, data)
		} else if pkt.ShopID == 7 {
			// hunter's road, 30 bytes chunks, item ID at 0x10, cost at 0x14, floor req at 0x04
			data, _ := hex.DecodeString("002C002C2C934441000030AC000001F400010000000000010000008C0000000100000A51B75F000030AD000001F400010000000000010000008C00000001000019B22C9D000030AE000003E80001000000000001000000410000000100001B91C74B000030AF000003E800010000000000010000004100000001000007749AB9000030B0000005DC00010000000000010000000F0000000100003797CA67000030B1000005DC0001000000000001000000230000000100000C707D05000030B2000005DC00010000000000010000003200000001000020B629B3000030B3000005DC00010000000000010000000F0000000100002D934471000030B4000005DC0001000000000001000000140000000100000B53B74F000030B5000005DC00010000000000010000003200000001000018B02CBD0000386D000003E80001000000000001000000010000000100001B91C76B0000386B00000BB800010000000000010000000500000001000006769AA90000386C000013880001000000000001000000050000000100003697CA5700003997000001F400010000000000010000002E0000000A00000C707D2500003A05000001F400010000000000010000002E0000000A000020B6298300003AAC000001F400010000000000010000002E0000000A00002D91445100003B08000001F400010000000000010000002E0000000A00000B53B75F00003FD2000001F400010000000000010000001F0000000A000019B02C8D00003FD3000001F400010000000000010000001F0000000A00001B91C77B00003835000001F400010000000000010000002E0000000A000007769AA900003AAD000001F400010000000000010000002E0000000A00003695CA4700003B07000001F400010000000000010000002E0000000A00000C727D3500003D9D000001F400010000000000010000002E0000000A000021B629B300003FD4000001F400010000000000010000002E0000003200002C91445100003FD5000001F400010000000000010000002E0000003200000A51B77F00003858000003E80001000000000001000000050005001E000019B02CAD00003AB8000003E80001000000000001000000050000002800001A93C76B00003B0D000003E800010000000000010000000500000032000006769A9900003DFC000003E80001000000000001000000050000003C00003695CA670000407C000003E80001000000000001000000050000004600000C727D050000407D000003E800010000000000010000000500000050000021B4299300003A62000003E80001000000000001000000050000006400002D93444100002C16000001F40001000000000001000000140000000100000B51B74F00003C58000003E800010000000000C8000000020002000B000019B02CBD0000392F000003E800010000000000C8000000050000000B00001B93C76B00003860000003E80001000000000001000000050000005A000006749A8900003861000003E80001000000000001000000050000006400003697CA7700003862000003E80001000000000001000000050000006E00000D727D3500003863000003E800010000000000010000000500000078000020B4299300003864000003E80001000000000001000000050000008200002D934451000036C5000003E80001000000000001000000050000008C00000B51B76F0000399600000FA000010000000000010000000A00000014000018B22CAD00003996000007D000010000000000010000000A0000002800001A93C75B000037DA000003E800010000000000C8000000050000000A0000")
			doSizedAckResp(s, pkt.AckHandle, data)
		} else if pkt.ShopID == 8 {
			data, _ := hex.DecodeString("000B000B00758A5800003836000000C800010000000000010000000A00000000000130F3B54600003837000000C800010000000000010000000A0000000000012BFDDC840000383C000001900001000000000001000000020000000000010C79E3420000383E00000190000100000000000100000002000000000001269337B000003839000001F40001000000000001000000020000000000011B1D4B7E0000383A000002BC0001000000000001000000020000000000020A7FDA9C0000383D0000032000010000000000010000000200000000000235950C2A0000383B000003E8000100000000000100000002000000000002007588580000383F0000044C00010000000000010000000200000000000224F3B5C600003838000007D00001000000000001000000030000000000033FFDDE840000386E000003E8000100000000000100000005000000000004")
			doSizedAckResp(s, pkt.AckHandle, data)
		} else {
			// generic respose for no items
			doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		}

	}

}

func handleMsgMhfGetExtraInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateInterior(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateHouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUpdateHouse(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadHouse(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHouse)
	// Seems to generate same response regardless of upgrade tier
	data, _ := hex.DecodeString("0000000000000000000000000000000000000000")
	doSizedAckResp(s, pkt.AckHandle, data)
}

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

func handleMsgMhfUpdateCafepoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateCafepoint)
	resp := byteframe.NewByteFrame()
	resp.WriteBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

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
	resp.WriteUint16(0x0000) // 0x0000 stops the vaguely annoying log in pop up
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

func handleMsgMhfEnumerateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuacot)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfUpdateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuacot)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfInfoTournament(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEntryTournament(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnterTournamentQuest(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireTournament(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfResetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPaymentAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfDisplayedAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfInfoScenarioCounter(s *Session, p mhfpacket.MHFPacket) {

	pkt := p.(*mhfpacket.MsgMhfInfoScenarioCounter)

	scenarioCounter := []struct {
		Unk0 uint32 // Main ID?
		Unk1 uint8
		Unk2 uint8
	}{
		{
			Unk0: 0x00000000,
			Unk1: 1,
			Unk2: 4,
		},
		{
			Unk0: 0x00000001,
			Unk1: 1,
			Unk2: 4,
		},
		{
			Unk0: 0x00000002,
			Unk1: 1,
			Unk2: 4,
		},
		{
			Unk0: 0x00000003,
			Unk1: 1,
			Unk2: 4,
		},
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(scenarioCounter))) // Entry count
	for _, entry := range scenarioCounter {
		resp.WriteUint32(entry.Unk0)
		resp.WriteUint8(entry.Unk1)
		resp.WriteUint8(entry.Unk2)
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())

	// DEBUG, DELETE ME!
	/*
		data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "debug/info_scenario_counter_resp.bin"))
		if err != nil {
			panic(err)
		}

		doSizedAckResp(s, pkt.AckHandle, data)
	*/

}

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

func handleMsgMhfGetMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetMyhouseInfo)
	// another save potentially since it can be updated?
	// set first byte to 1 to avoid pop up every time without save
	body := make([]byte, 0x16A)
	// parity with the only packet capture available
	//body[0] = 10;
	//body[21] = 10;
	doSizedAckResp(s, pkt.AckHandle, body)
}

func handleMsgMhfUpdateMyhouseInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateMyhouseInfo)
	// looks to be the sized datachunk from above without the size bytes, quite possibly intended to be persistent
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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

func handleMsgMhfStampcardStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStampcardStamp)
	// TODO: Work out where it gets existing stamp count from, its format and then
	// update the actual sent values to be correct
	doSizedAckResp(s, pkt.AckHandle, []byte{0x03, 0xe7, 0x03, 0xe7, 0x02, 0x99, 0x02, 0x9c, 0x00, 0x00, 0x00, 0x00, 0x14, 0xf8, 0x69, 0x54})
}

func handleMsgMhfStampcardPrize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUnreserveSrg(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadPlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateData)
	var data []byte
	err := s.server.db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get plate data savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		doSizedAckResp(s, pkt.AckHandle, []byte{})
	}
}

func handleMsgMhfSavePlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateData)
	err := ioutil.WriteFile(fmt.Sprintf("savedata\\%d_platedata.bin", time.Now().Unix()), pkt.RawDataPayload, 0644)
	if err != nil {
		s.logger.Fatal("Error dumping platedata", zap.Error(err))
	}
	if pkt.IsDataDiff {
		// https://gist.github.com/Andoryuuta/9c524da7285e4b5ca7e52e0fc1ca1daf
		var data []byte
		//load existing save
		err := s.server.db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Fatal("Failed to get platedata savedata from db", zap.Error(err))
		}
		//decompress
		fmt.Println("Decompressing...")
		data, err = saveDecompress(data)
		if err != nil {
			s.logger.Fatal("Failed to decompress platedata from db", zap.Error(err))
		}
		// perform diff and compress it to write back to db
		fmt.Println("Diffing...")
		saveOutput, err := saveCompress(saveDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.logger.Fatal("Failed to diff and compress platedata savedata", zap.Error(err))
		}
		_, err = s.server.db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platedata savedata in db", zap.Error(err))
		}
	} else {
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platedata savedata in db", zap.Error(err))
		}
	}
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateBox)
	var data []byte
	err := s.server.db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get sigil box savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		doSizedAckResp(s, pkt.AckHandle, []byte{})
	}
}

func handleMsgMhfSavePlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateBox)
	err := ioutil.WriteFile(fmt.Sprintf("savedata\\%d_platebox.bin", time.Now().Unix()), pkt.RawDataPayload, 0644)
	if err != nil {
		s.logger.Fatal("Error dumping hunter platebox savedata", zap.Error(err))
	}
	if pkt.IsDataDiff {
		var data []byte
		//load existing save
		err := s.server.db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Fatal("Failed to get sigil box savedata from db", zap.Error(err))
		}
		//decompress
		fmt.Println("Decompressing...")
		data, err = saveDecompress(data)
		if err != nil {
			s.logger.Fatal("Failed to decompress savedata from db", zap.Error(err))
		}
		// perform diff and compress it to write back to db
		fmt.Println("Diffing...")
		saveOutput, err := saveCompress(saveDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.logger.Fatal("Failed to diff and compress savedata", zap.Error(err))
		}
		_, err = s.server.db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platebox savedata in db", zap.Error(err))
		} else {
			fmt.Println("Wrote recompressed save back to DB.")
		}

	} else {
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platedata savedata in db", zap.Error(err))
		}
	}
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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

func handleMsgMhfGetAdditionalBeatReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAdditionalBeatReward)
	// Actual response in packet captures are all just giant batches of null bytes
	// I'm assuming this is because it used to be tied to an actual event and
	// they never bothered killing off the packet when they made it static
	doSizedAckResp(s, pkt.AckHandle, make([]byte, 0x104))
}

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
	// load partner from database
	var data []byte
	err := s.server.db.QueryRow("SELECT partner FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get partner savedata from db", zap.Error(err))
	}
	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
	// TODO(Andoryuuta): Figure out unusual double ack. One sized, one not.
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfSavePartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePartner)
	err := ioutil.WriteFile(fmt.Sprintf("savedata\\%d_partner.bin", time.Now().Unix()), pkt.RawDataPayload, 0644)
	if err != nil {
		s.logger.Fatal("Error dumping partner savedata", zap.Error(err))
	}

	_, err = s.server.db.Exec("UPDATE characters SET partner=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update partner savedata in db", zap.Error(err))
	}
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetGuildMissionList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildMissionRecord(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddGuildMissionCount(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCancelGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadOtomoAirou)
	// load partnyaa from database
	var data []byte
	err := s.server.db.QueryRow("SELECT otomoairou FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get partnyaa savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgMhfSaveOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveOtomoAirou)
	err := ioutil.WriteFile(fmt.Sprintf("savedata\\%d_otomoairou.bin", time.Now().Unix()), pkt.RawDataPayload, 0644)
	if err != nil {
		s.logger.Fatal("Error dumping partnyaa savedata", zap.Error(err))
	}

	_, err = s.server.db.Exec("UPDATE characters SET otomoairou=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update partnyaa savedata in db", zap.Error(err))
	}
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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
	var data []byte
	err := s.server.db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get preset decorations savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
		//doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		// set first byte to 1 to avoid pop up every time without save
		body := make([]byte, 0x226)
		body[0] = 1
		doSizedAckResp(s, pkt.AckHandle, body)
	}
}

func handleMsgMhfSaveDecoMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveDecoMyset)
	// https://gist.github.com/Andoryuuta/9c524da7285e4b5ca7e52e0fc1ca1daf
	var loadData []byte
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload[1:]) // skip first unk byte
	err := s.server.db.QueryRow("SELECT decomyset FROM characters WHERE id = $1", s.charID).Scan(&loadData)
	if err != nil {
		s.logger.Fatal("Failed to get preset decorations savedata from db", zap.Error(err))
	} else {
		numSets := bf.ReadUint8() // sets being written
		// empty save
		if len(loadData) == 0 {
			loadData = []byte{0x01, 0x00}
		}

		savedSets := loadData[1] // existing saved sets
		// no sets, new slice with just first 2 bytes for appends later
		if savedSets == 0 {
			loadData = []byte{0x01, 0x00}
		}
		for i := 0; i < int(numSets); i++ {
			writeSet := bf.ReadUint16()
			dataChunk := bf.ReadBytes(76)
			setBytes := append([]byte{uint8(writeSet >> 8), uint8(writeSet & 0xff)}, dataChunk...)
			for x := 0; true; x++ {
				if x == int(savedSets) {
					// appending set
					if loadData[len(loadData)-1] == 0x10 {
						// sanity check for if there was a messy manual import
						loadData = append(loadData[:len(loadData)-2], setBytes...)
					} else {
						loadData = append(loadData, setBytes...)
					}
					savedSets++
					break
				}
				currentSet := loadData[3+(x*78)]
				if int(currentSet) == int(writeSet) {
					// replacing a set
					loadData = append(loadData[:2+(x*78)], append(setBytes, loadData[2+((x+1)*78):]...)...)
					break
				} else if int(currentSet) > int(writeSet) {
					// inserting before current set
					loadData = append(loadData[:2+((x)*78)], append(setBytes, loadData[2+((x)*78):]...)...)
					savedSets++
					break
				}
			}
			loadData[1] = savedSets // update set count
		}
		_, err := s.server.db.Exec("UPDATE characters SET decomyset=$1 WHERE id=$2", loadData, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update decomyset savedata in db", zap.Error(err))
		}
	}
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

}

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
	var data []byte
	err := s.server.db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get hunter navigation savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
		//doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		// set first byte to 1 to avoid pop up every time without save
		body := make([]byte, 0x226)
		body[0] = 1
		doSizedAckResp(s, pkt.AckHandle, body)
	}
}

func handleMsgMhfSaveHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveHunterNavi)
	err := ioutil.WriteFile(fmt.Sprintf("savedata\\%d_hunternavi.bin", time.Now().Unix()), pkt.RawDataPayload, 0644)
	if err != nil {
		s.logger.Fatal("Error dumping hunter navigation savedata", zap.Error(err))
	}

	if pkt.IsDataDiff {
		// https://gist.github.com/Andoryuuta/9c524da7285e4b5ca7e52e0fc1ca1daf
		// doesn't seem fully consistent with platedata?
	} else {
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET hunternavi=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update hunternavi savedata in db", zap.Error(err))
		}
	}
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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

func handleMsgMhfPostTowerInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTowerInfo)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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
	// if the game gets bad responses for this it breaks the ability to save
	pkt := p.(*mhfpacket.MsgMhfGetPaperData)
	var data []byte
	var err error
	if pkt.Unk2 == 4 {
		data, err = hex.DecodeString("0A218EAD000000000000000000000000")
	} else if pkt.Unk2 == 5 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000003403E900010000000000000000000003E900020000000000000000000003EB00010064006400C80064000003EB00020096006400F00064000003EC000A270F002800000000000003ED000A01F4000000000000000003EF00010000000000000000000003F000C801900BB801900BB8000003F200010FA0000000000000000003F200020FA0000000000000000003F3000117703A984E2061A8753003F3000217703A984E2061A8753003F400011F40445C57E46B6C791803F400021F40445C57E46B6C791803F700010010001000100000000003F7000200100010001000000000044D000107E001F4000000000000044D000207E001F4000000000000044F0001000000000BB800000BB8044F0002000000000BB800000BB804500001000A270F00280000000004500002000A270F00280000000004510001000A01F400000000000004510002000A01F400000000000007D100010011003A0000000602BC07D100010014003A0000000300C807D100010016003A0000000700FA07D10001001B003A00000001006407D100010035003A0000000803E807D100010043003A0000000901F407D100010044003A00000002009607D10001004A003A0000000400C807D10001004B003A0000000501F407D10001004C003A0000000A032007D100010050003A0000000B038407D100010059003A0000000C025807D100020011003C0000000602BC07D100020014003C0000000300C807D100020016003C00000007015E07D10002001B003C00000001006407D100020027003C0000000D00C807D100020028003C0000000F025807D100020035003C0000000803E807D100020043003C0000000201F407D100020044003C00000009009607D10002004A003C0000000400C807D10002004B003C0000000501F407D10002004C003C0000000A032007D100020050003C0000000B038407D100020051003C0000000E038407D100020059003C0000000C025807D10002005E003C0000001003E8")
	} else if pkt.Unk2 == 6 {
		data, err = hex.DecodeString("0A218EAD0000000000000000000001A503EA00640000000000000000000003EE00012710271000000000000003EE000227104E2000000000000003F100140000000000000000000003F5000100010001006400C8012C03F5000100010002006400C8012C03F5000100020001012C006400C803F5000100020002012C006400C803F500010003000100C8012C006403F500010003000200C8012C006403F5000200010001012C006400C803F5000200010002012C006400C803F500020002000100C8012C006403F500020002000200C8012C006403F5000200030001006400C8012C03F5000200030002006400C8012C03F500030001000100C8012C006403F500030001000200C8012C006403F5000300020001006400C8012C03F5000300020002006400C8012C03F5000300030001012C006400C803F5000300030002012C006400C803F800010001005000000000000003F800010002005000000000000003F800010003005000000000000003F800020001005000000000000003F800020002005000000000000003F800020003005000000000000004B10001003C003200000000000004B10002003C003200000000000004B200010000000500320000000004B2000100060014003C0000000004B200010015002800460000000004B200010029007800500000000004B20001007900A0005A0000000004B2000100A100FA00640000000004B2000100FB01F400640000000004B2000101F5270F00640000000004B200020000006400640000000004B20002006500C800640000000004B2000200C901F400960000000004B2000201F5270F00960000000004B3000100000005000A0000000004B300010006000A00140000000004B30001000B001E001E0000000004B30001001F003C00280000000004B30001003D007800320000000004B3000100790082003C0000000004B300010083008C00460000000004B30001008D009600500000000004B30001009700A000550000000004B3000100A100C800640000000004B3000100C901F400640000000004B3000101F5270F00640000000004B300020000007800460000000004B30002007901F400780000000004B3000201F5270F00780000000004B4000100000005000F0000000004B400010006000A00140000000004B40001000B000F00190000000004B4000100100014001B0000000004B4000100150019001E0000000004B40001001A001E00200000000004B40001001F002800230000000004B400010029003200250000000004B400010033003C00280000000004B40001003D0046002B0000000004B4000100470050002D0000000004B400010051005A002F0000000004B40001005B006400320000000004B400010065006E003C0000000004B40001006F007800460000000004B4000100790082004B0000000004B400010083008C00520000000004B40001008D00A000550000000004B4000100A100C800640000000004B4000100C901F400640000000004B4000101F5270F00640000000004B400020000007800460000000004B40002007901F400780000000004B4000201F5270F0078000000000FA10001000000000000000000000FA10002000029AB0005000000010FA10002000029AB0005000000010FA10002000029AB0005000000010FA10002000029AB0005000000010FA10002000029AC0002000000010FA10002000029AC0002000000010FA10002000029AC0002000000010FA10002000029AC0002000000010FA10002000029AD0001000000010FA10002000029AD0001000000010FA10002000029AD0001000000010FA10002000029AD0001000000010FA10002000029AF0003000000010FA10002000029AF0003000000010FA10002000029AF0003000000010FA10002000029AF0003000000010FA10002000028900001000000010FA10002000028900001000000010FA10002000029AE0002000000010FA10002000029AE0002000000010FA10002000029BA0002000000010FA10002000029BB0002000000010FA10002000029B60001000000010FA10002000029B60001000000010FA5000100002B970001138800010FA5000100002B9800010D1600010FA5000100002B99000105DC00010FA5000100002B9A0001006400010FA5000100002B9B0001003200010FA5000200002B970002070800010FA5000200002B98000204B000010FA5000200002B99000201F400010FA5000200002B9A0001003200010FA5000200002B1D0001009600010FA5000200002B1E0001009600010FA5000200002B240001009600010FA5000200002B310001009600010FA5000200002B330001009600010FA5000200002B470001009600010FA5000200002B5A0001009600010FA5000200002B600001009600010FA5000200002B6D0001009600010FA5000200002B780001009600010FA5000200002B7D0001009600010FA5000200002B810001009600010FA5000200002B870001009600010FA5000200002B7C0001009600010FA5000200002B1F0001009600010FA5000200002B200001009600010FA5000200002B290001009600010FA5000200002B350001009600010FA5000200002B370001009600010FA5000200002B450001009600010FA5000200002B5B0001009600010FA5000200002B610001009600010FA5000200002B790001009600010FA5000200002B7A0001009600010FA5000200002B7B0001009600010FA5000200002B830001009600010FA5000200002B890001009600010FA5000200002B580001009600010FA5000200002B210001009600010FA5000200002B270001009600010FA5000200002B2E0001009600010FA5000200002B390001009600010FA5000200002B3C0001009600010FA5000200002B430001009600010FA5000200002B5C0001009600010FA5000200002B620001009600010FA5000200002B6F0001009600010FA5000200002B7F0001009600010FA5000200002B800001009600010FA5000200002B820001009600010FA5000200002B500001009600010FA50002000028820001009600010FA50002000028800001009600010FA6000100002B970001138800010FA6000100002B9800010D1600010FA6000100002B99000105DC00010FA6000100002B9A0001006400010FA6000100002B9B0001003200010FA6000200002B970002070800010FA6000200002B98000204B000010FA6000200002B99000201F400010FA6000200002B9A0001003200010FA6000200002B1D0001009600010FA6000200002B1E0001009600010FA6000200002B240001009600010FA6000200002B310001009600010FA6000200002B330001009600010FA6000200002B470001009600010FA6000200002B5A0001009600010FA6000200002B600001009600010FA6000200002B6D0001009600010FA6000200002B780001009600010FA6000200002B7D0001009600010FA6000200002B810001009600010FA6000200002B870001009600010FA6000200002B7C0001009600010FA6000200002B1F0001009600010FA6000200002B200001009600010FA6000200002B290001009600010FA6000200002B350001009600010FA6000200002B370001009600010FA6000200002B450001009600010FA6000200002B5B0001009600010FA6000200002B610001009600010FA6000200002B790001009600010FA6000200002B7A0001009600010FA6000200002B7B0001009600010FA6000200002B830001009600010FA6000200002B890001009600010FA6000200002B580001009600010FA6000200002B210001009600010FA6000200002B270001009600010FA6000200002B2E0001009600010FA6000200002B390001009600010FA6000200002B3C0001009600010FA6000200002B430001009600010FA6000200002B5C0001009600010FA6000200002B620001009600010FA6000200002B6F0001009600010FA6000200002B7F0001009600010FA6000200002B800001009600010FA6000200002B820001009600010FA6000200002B500001009600010FA60002000028820001009600010FA60002000028800001009600010FA7000100002B320001004600010FA7000100002B340001004600010FA7000100002B360001004600010FA7000100002B380001004600010FA7000100002B3A0001004600010FA7000100002B6E0001004600010FA7000100002B700001004600010FA7000100002B660001004600010FA7000100002B680001004600010FA7000100002B6A0001004600010FA7000100002B220001004600010FA7000100002B230001004600010FA7000100002B420001004600010FA7000100002B840001004600010FA7000100002B3B0001004600010FA7000100002B280001004600010FA7000100002B260001004600010FA7000100002B5F0001004600010FA7000100002B630001004600010FA7000100002B640001004600010FA7000100002B710001004600010FA7000100002B7E0001004600010FA7000100002B4C0001004600010FA7000100002B4D0001004600010FA7000100002B4E0001004600010FA7000100002B4F0001004600010FA7000100002B560001004600010FA7000100002B570001004600010FA70001000028860001004600010FA70001000028870001004600010FA70001000028880001004600010FA70001000028890001004600010FA700010000288A0001004600010FA7000100002B3D0001002D00010FA7000100002B3F0001002D00010FA7000100002B410001002D00010FA7000100002B440001002D00010FA7000100002B460001002D00010FA7000100002B6C0001002D00010FA7000100002B730001002D00010FA7000100002B770001002D00010FA7000100002B860001002D00010FA7000100002B300001002D00010FA7000100002B520001002D00010FA7000100002B590001002D00010FA700010000287F0001002D00010FA70001000028830001002D00010FA70001000028850001002D00010FA7000100002B480001000F00010FA7000100002B490001000F00010FA7000100002B4B0001000F00010FA7000100002B750001000F00010FA7000100002B550001000E00010FA7000100002B2D0001000A00010FA7000100002B8B0001000A00010FA70001000028840001000500010FA70001000028810001000100010FA7000100002B9B0001009600010FA7000100002CC90001003200010FA7000100002CCA0001001900010FA7000100002CCB000100C800010FA7000100002CCC0001019000010FA7000100002CCD0001009600010FA7000100002B1D0001005C00010FA7000100002B1E0001005C00010FA7000100002B240001005C00010FA7000100002B310001005C00010FA7000100002B330001005C00010FA7000100002B470001005C00010FA7000100002B5A0001005C00010FA7000100002B600001005C00010FA7000100002B6D0001005C00010FA7000100002B7D0001005C00010FA7000100002B810001005C00010FA7000100002B870001005C00010FA7000100002B7C0001005C00010FA7000100002B1F0001005C00010FA7000100002B200001005C00010FA7000100002B290001005C00010FA7000100002B350001005C00010FA7000100002B370001005C00010FA7000100002B450001005C00010FA7000100002B5B0001005C00010FA7000100002B610001005C00010FA7000100002B790001005C00010FA7000100002B7A0001005C00010FA7000100002B7B0001005C00010FA7000100002B830001005C00010FA7000100002B890001005B00010FA7000100002B580001005B00010FA7000100002B210001005B00010FA7000100002B270001005B00010FA7000100002B2E0001005B00010FA7000100002B390001005B00010FA7000100002B3C0001005B00010FA7000100002B430001005B00010FA7000100002B5C0001005B00010FA7000100002B620001005B00010FA7000100002B6F0001005B00010FA7000100002B7F0001005B00010FA7000100002B800001005B00010FA7000100002B820001005B00010FA7000100002B500001005B00010FA70001000028820001005B00010FA70001000028800001005B00010FA7000100002B250001005B00010FA7000100002B3E0001005B00010FA7000100002B5D0001005B00010FA7000100002B650001005B00010FA7000100002B720001005B00010FA7000100002B850001005B00010FA7000100002B2B0001005B00010FA7000100002B5E0001005B00010FA7000100002B740001005B00010FA7000100002B400001005B00010FA7000100002B4A0001005B00010FA7000100002B6B0001005B00010FA7000100002B880001005B00010FA7000100002B510001005B00010FA7000100002B530001005B00010FA7000100002B540001005B00010FA7000100002B2A0001005B00010FA7000100002B670001005B00010FA7000100002B690001005B00010FA7000100002B760001005B00010FA7000100002B2F0001005B00010FA7000100002B2C0001005B00010FA7000100002B8A0001005B00010FA7000200002B320001005A00010FA7000200002B340001005A00010FA7000200002B360001005A00010FA7000200002B380001005A00010FA7000200002B3A0001005A00010FA7000200002B6E0001005A00010FA7000200002B700001005A00010FA7000200002B660001005A00010FA7000200002B680001005A00010FA7000200002B6A0001005A00010FA7000200002B220001005A00010FA7000200002B230001005A00010FA7000200002B420001005A00010FA7000200002B840001005A00010FA7000200002B3B0001005A00010FA7000200002B280001005A00010FA7000200002B260001005A00010FA7000200002B5F0001005A00010FA7000200002B630001005A00010FA7000200002B640001005A00010FA7000200002B710001005A00010FA7000200002B7E0001005A00010FA7000200002B4C0001005A00010FA7000200002B4D0001005A00010FA7000200002B4E0001005A00010FA7000200002B4F0001005A00010FA7000200002B560001005A00010FA7000200002B570001005A00010FA70002000028860001005A00010FA70002000028870001005A00010FA70002000028880001005A00010FA70002000028890001005A00010FA700020000288A0001005A00010FA7000200002B3D0001005000010FA7000200002B3F0001005000010FA7000200002B410001005000010FA7000200002B440001005000010FA7000200002B460001005000010FA7000200002B6C0001005000010FA7000200002B730001005000010FA7000200002B770001005000010FA7000200002B860001005000010FA7000200002B300001005000010FA7000200002B520001005000010FA7000200002B590001005000010FA700020000287F0001005000010FA70002000028830001005000010FA70002000028850001005000010FA7000200002B480001001600010FA7000200002B490001001600010FA7000200002B4B0001001600010FA7000200002B750001001600010FA7000200002B550001001600010FA7000200002B2D0001000F00010FA7000200002B8B0001000F00010FA70002000028840001000800010FA70002000028810001000200010FA7000200002B97000304C400010FA7000200002B980003028A00010FA7000200002B99000300A000010FA7000200002D8D0001032000010FA7000200002D8E0001032000010FA7000200002B9B000101F400010FA7000200002B9A0001022600010FA7000200002CC90001003200010FA7000200002CCA0001001900010FA7000200002CCB000100FA00010FA7000200002CCC000101F400010FA7000200002CCD000100AF0001106A000100002B9B000117700001106A000100002CC9000100C80001106A000100002CCA000100640001106A000100002CCB000103E80001106A000100002CCC000107D00001106A000100002CCD000102BC0001106A000200002D8D000103200001106A000200002D8E000103200001106A000200002B9B000101900001106A000200002CC9000101900001106A000200002CCA000100C80001106A000200002CCB000107D00001106A000200002CCC00010FA00001106A000200002CCD000105780001")
	} else if pkt.Unk2 == 6001 {
		data, err = hex.DecodeString("0A218EAD0000000000000000000000052B97010113882B9801010D162B99010105DC2B9A010100642B9B01010032")
	} else if pkt.Unk2 == 6002 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000002F2B97020107082B98020104B02B99020101F42B9A010100322B1D010100962B1E010100962B24010100962B31010100962B33010100962B47010100962B5A010100962B60010100962B6D010100962B78010100962B7D010100962B81010100962B87010100962B7C010100962B1F010100962B20010100962B29010100962B35010100962B37010100962B45010100962B5B010100962B61010100962B79010100962B7A010100962B7B010100962B83010100962B89010100962B58010100962B21010100962B27010100962B2E010100962B39010100962B3C010100962B43010100962B5C010100962B62010100962B6F010100962B7F010100962B80010100962B82010100962B5001010096288201010096288001010096")
	} else if pkt.Unk2 == 6010 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000000B2B9701010E742B9801010B542B99010105142CBD010100FA2CBE010100FA2F17010100FA2F21010100FA2F1A010100FA2F24010100FA2DFE010100C82DFD01010190")
	} else if pkt.Unk2 == 6011 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000000B2B9701010E742B9801010B542B99010105142CBD010100FA2CBE010100FA2F17010100FA2F21010100FA2F1A010100FA2F24010100FA2DFE010100C82DFD01010190")
	} else if pkt.Unk2 == 6012 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000000D2B9702010DAC2B9802010B542B990201051430DC010101902CBD010100C82CBE010100C82F17010100C82F21010100C82F1A010100C82F24010100C82DFF010101902E00010100C82E0101010064")
	} else if pkt.Unk2 == 7001 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000009D2B1D010101222B1E0101010E2B240101010E2B31010101222B33010101222B47010101222B5A010101182B600101012C2B6D010101182B78010101222B7D010101222B810101012C2B87010101222B7C0101010E2B220101002F2B250101002F2B380101002F2B360101002F2B3E010100302B5D0101002F2B640101002F2B650101002F2B700101002F2B720101002F2B7E0101002F2B850101002F2B4C0101002F2B4F0101002F2B560101002F28860101002F28870101002F2B2B010100112B3F010100102B44010100102B5E010100112B74010100112B52010100112B97010104B02B970201028A2B98010103202B980201012C2B99010100642B99020100322B9C010100642B9A010100642B9B010100642B960101012C2CC70101012C2C5C0101012C2CC80101012C2C5D010101F42B1F0102012C2B200102010E2B290102012C2B35010201222B37010201222B45010201222B5B010201182B610102012C2B79010200FA2B7A0102012C2B7B010201182B83010201222B89010201042B580102012C2B260102002F2B3A0102002F2B3B0102002F2B400102002F2B4A0102002F2B5F0102002F2B660102002F2B680102002F2B6A0102002F2B6B0102002F2B710102002F2B88010200302B4D0102002F2B510102002F2B530102002F28880102002F28890102002F2B77010200112B3D010200112B86010200112B46010200112B30010200102B54010200102B97010204B02B970202028A2B98010203202B980202012C2B99010200642B99020200322B9C010200642B9A010200642B9B010200642B960102012C2CC70102012C2C5C0102012C2CC80102012C2C5D010201F42B210103010A2B270103010A2B2E0103010A2B390103010A2B3C0103010A2B430103010A2B5C0103010A2B620103010A2B6F0103010A2B7F0103010C2B800103010C2B820103010C2B500103010C28820103010A28800103010C2B23010300322B28010300322B2A010300322B32010300322B34010300322B42010300322B63010300322B67010300322B69010300322B6E010300322B76010300322B84010300322B4E010300322B57010300322B2F01030032288A010300322B2C0103000F2B410103000F2B8A0103000F2B6C0103000F2B730103000F2B590103000F287F0103000F28830103000F28850103000F2A1A010301772BC9010301772A3D010301772C7D010301772B97010303E82B97020300FA2B98010302BC2B98020300AF2B990103012C2B990203004B2CC9010300352CCA0103001B2CCB0103010A2CCC010302152CCD010300BA")
	} else if pkt.Unk2 == 7002 {
		data, err = hex.DecodeString("0A218EAD0000000000000000000000B92B1D010100642B1E010100642B24010100642B31010100642B33010100642B47010100642B5A010100642B60010100642B6D010100642B78010100642B7D010100642B81010100642B87010100642B7C010100642B220101003C2B250101003C2B380101003C2B360101003C2B3E0101003C2B5D0101003C2B640101003C2B650101003C2B700101003C2B720101003C2B7E0101003C2B850101003C2B4C0101003C2B4F0101003C2B560101003C28860101003C28870101003C2B2B010100142B3F010100142B44010100142B5E010100142B74010100142B52010100142B9C010101902B9A010100C82B9B010100C82CC7010100642CC80101009628730101009630DA010100C830DB0101012C30DC01010384353D0101015E353C010100C82C5C010100642C5D010100962EEE010100FA2EF0010101902EEF0101019A2B97020101F42B97040101F42B97060101F42B98020101902B98040101902B98060101902B99020100642B99040100642B99060100642B1F010200642B20010200642B29010200642B35010200642B37010200642B45010200642B5B010200642B61010200642B79010200642B7A010200642B7B010200642B83010200642B89010200642B58010200642B260102003C2B3A0102003C2B3B0102003C2B400102003C2B4A0102003C2B5F0102003C2B660102003C2B680102003C2B6A0102003C2B6B0102003C2B710102003C2B880102003C2B4D0102003C2B510102003C2B530102003C28880102003C28890102003C2B77010200142B3D010200142B86010200142B46010200142B30010200142B54010200142B9C010201902B9A010200C82B9B010200C82CC7010200FA2CC80102015E30DA0102009630DB010200C830DC0102015E353D010200FA353C010200C82873010201902B96010200642C5C010200642C5D010200642EEE0102012C2EF0010201C22EEF010201CC2B97020201F42B97040201F42B97060201F42B98020201902B98040201902B98060201902B99020200642B99040200642B99060200642B21010300782B27010300782B2E010300782B39010300782B3C010300782B43010300782B5C010300782B62010300782B6F010300782B7F010300782B80010300782B82010300782B50010300782882010300782880010300782B23010300412B28010300412B2A010300412B32010300412B34010300412B42010300412B63010300412B67010300412B69010300412B6E010300412B76010300412B84010300412B4E010300412B57010300412B2F01030041288A010300412B2C0103000F2B410103000F2B8A0103000F2B6C0103000F2B730103000F2B590103000F287F0103000F28830103000F28850103000F2A1A030301EA2BC9030301EA2A3D030301EA2C7D030301EA2F0E030301F430D7030301F42B97020301F42B97040301F42B97060301F42B98020301902B98040301902B98060301902B99020300642B99040300642B99060300642CC9010300352CCA0103001B2CCB0103010A2CCC010302152CCD010300BA")
	} else if pkt.Unk2 == 7011 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000009D2B1D010101222B1E0101010E2B240101010E2B31010101222B33010101222B47010101222B5A010101182B600101012C2B6D010101182B78010101222B7D010101222B810101012C2B87010101222B7C0101010E2B220101002F2B250101002F2B380101002F2B360101002F2B3E010100302B5D0101002F2B640101002F2B650101002F2B700101002F2B720101002F2B7E0101002F2B850101002F2B4C0101002F2B4F0101002F2B560101002F28860101002F28870101002F2B2B010100112B3F010100102B44010100102B5E010100112B74010100112B52010100112B97010104B02B970201028A2B98010103202B980201012C2B99010100642B99020100322B9C010100642B9A010100642B9B010100642B960101012C2CC70101012C2C5C0101012C2CC80101012C2C5D010101F42B1F0102012C2B200102010E2B290102012C2B35010201222B37010201222B45010201222B5B010201182B610102012C2B79010200FA2B7A0102012C2B7B010201182B83010201222B89010201042B580102012C2B260102002F2B3A0102002F2B3B0102002F2B400102002F2B4A0102002F2B5F0102002F2B660102002F2B680102002F2B6A0102002F2B6B0102002F2B710102002F2B88010200302B4D0102002F2B510102002F2B530102002F28880102002F28890102002F2B77010200112B3D010200112B86010200112B46010200112B30010200102B54010200102B97010204B02B970202028A2B98010203202B980202012C2B99010200642B99020200322B9C010200642B9A010200642B9B010200642B960102012C2CC70102012C2C5C0102012C2CC80102012C2C5D010201F42B210103010A2B270103010A2B2E0103010A2B390103010A2B3C0103010A2B430103010A2B5C0103010A2B620103010A2B6F0103010A2B7F0103010C2B800103010C2B820103010C2B500103010C28820103010A28800103010C2B23010300322B28010300322B2A010300322B32010300322B34010300322B42010300322B63010300322B67010300322B69010300322B6E010300322B76010300322B84010300322B4E010300322B57010300322B2F01030032288A010300322B2C0103000F2B410103000F2B8A0103000F2B6C0103000F2B730103000F2B590103000F287F0103000F28830103000F28850103000F2A1A010301772BC9010301772A3D010301772C7D010301772B97010303E82B97020300FA2B98010302BC2B98020300AF2B990103012C2B990203004B2CC9010300352CCA0103001B2CCB0103010A2CCC010302152CCD010300BA")
	} else if pkt.Unk2 == 7012 {
		data, err = hex.DecodeString("0A218EAD00000000000000000000009D2B1D010101222B1E0101010E2B240101010E2B31010101222B33010101222B47010101222B5A010101182B600101012C2B6D010101182B78010101222B7D010101222B810101012C2B87010101222B7C0101010E2B220101002F2B250101002F2B380101002F2B360101002F2B3E010100302B5D0101002F2B640101002F2B650101002F2B700101002F2B720101002F2B7E0101002F2B850101002F2B4C0101002F2B4F0101002F2B560101002F28860101002F28870101002F2B2B010100112B3F010100102B44010100102B5E010100112B74010100112B52010100112B97010104B02B970201028A2B98010103202B980201012C2B99010100642B99020100322B9C010100642B9A010100642B9B010100642B960101012C2CC70101012C2C5C0101012C2CC80101012C2C5D010101F42B1F0102012C2B200102010E2B290102012C2B35010201222B37010201222B45010201222B5B010201182B610102012C2B79010200FA2B7A0102012C2B7B010201182B83010201222B89010201042B580102012C2B260102002F2B3A0102002F2B3B0102002F2B400102002F2B4A0102002F2B5F0102002F2B660102002F2B680102002F2B6A0102002F2B6B0102002F2B710102002F2B88010200302B4D0102002F2B510102002F2B530102002F28880102002F28890102002F2B77010200112B3D010200112B86010200112B46010200112B30010200102B54010200102B97010204B02B970202028A2B98010203202B980202012C2B99010200642B99020200322B9C010200642B9A010200642B9B010200642B960102012C2CC70102012C2C5C0102012C2CC80102012C2C5D010201F42B210103010A2B270103010A2B2E0103010A2B390103010A2B3C0103010A2B430103010A2B5C0103010A2B620103010A2B6F0103010A2B7F0103010C2B800103010C2B820103010C2B500103010C28820103010A28800103010C2B23010300322B28010300322B2A010300322B32010300322B34010300322B42010300322B63010300322B67010300322B69010300322B6E010300322B76010300322B84010300322B4E010300322B57010300322B2F01030032288A010300322B2C0103000F2B410103000F2B8A0103000F2B6C0103000F2B730103000F2B590103000F287F0103000F28830103000F28850103000F2A1A010301772BC9010301772A3D010301772C7D010301772B97010303E82B97020300FA2B98010302BC2B98020300AF2B990103012C2B990203004B2CC9010300352CCA0103001B2CCB0103010A2CCC010302152CCD010300BA")
	} else {
		data = []byte{0x00, 0x00, 0x00, 0x00}
		s.logger.Info("GET_PAPER request for unknown type")
	}
	if err != nil {
		panic(err)
	}
	doSizedAckResp(s, pkt.AckHandle, data)
	//	s.QueueAck(pkt.AckHandle, data)

}

func handleMsgMhfGetNotice(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostNotice(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTime)
	doSizedAckResp(s, pkt.AckHandle, []byte{})

	// Update the client's rights as well:
	updateRights(s)
}

func handleMsgMhfPostBoostTime(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgMhfPostBoostTime)
}

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
	// temp values from actual char, 4 bytes header, int32s for real gacha, trial gacha, frontier points
	// presumably should be made persistent and into another database entry
	data, _ := hex.DecodeString("0100000C0000000000000312000001E80010")
	s.QueueAck(pkt.AckHandle, data)

	// this sure breaks this horrifically 	doSizedAckResp(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfUseGachaPoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfExchangeFpoint2Item(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfExchangeItem2Fpoint(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetFpointExchangeList(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPlayStepupGacha(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfReceiveGachaItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetStepupStatus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPlayFreeGacha(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTinyBin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTinyBin)
	// requested after conquest quests
	// 00 02 01 req returns 01 00 00 00 so using that as general placeholder
	s.QueueAck(pkt.AckHandle, []byte{0x01, 0x00, 0x00, 0x00})
}

func handleMsgMhfPostTinyBin(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetSenyuDailyCount(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetGuildTargetMemberNum(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetBoostRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostRight)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfStartBoostTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostBoostTimeQuestReturn(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTimeQuestReturn)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPlayBoxGacha(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfResetBoxGachaInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetSeibattle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetSeibattle)
	stubGetNoResults(s, pkt.AckHandle)
}

func handleMsgMhfPostSeibattle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRyoudama(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPostRyoudama(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTenrouirai(s *Session, p mhfpacket.MHFPacket) {
	// if the game gets bad responses for this it breaks the ability to save
	pkt := p.(*mhfpacket.MsgMhfGetTenrouirai)
	var data []byte
	var err error
	if pkt.Unk0 == 1 {
		data, err = hex.DecodeString("0A218EAD000000000000000000000001010000000000060010")
	} else if pkt.Unk2 == 4 {
		data, err = hex.DecodeString("0A218EAD0000000000000000000000210101005000000202010102020104001000000202010102020106003200000202010002020104000C003202020101020201030032000002020101020202059C4000000202010002020105C35000320202010102020201003C00000202010102020203003200000201010001020203002800320201010101020204000C00000201010101020206002800000201010001020101003C00320201020101020105C35000000301020101020106003200000301020001020104001000320301020101020105C350000003010201010202030028000003010200010201030032003203010201010202059C4000000301020101010206002800000301020001010201003C00320301020101010206003200000301020101010204000C000003010200010101010050003203010201010101059C40000003010201010101030032000003010200010101040010003203010001010101060032000003010001010102030028000003010001010101010050003203010000010102059C4000000301000001010206002800000301000001010010")
	} else {
		data = []byte{0x00, 0x00, 0x00, 0x00}
		s.logger.Info("GET_TENROUIRAI request for unknown type")
	}
	if err != nil {
		panic(err)
	}
	doSizedAckResp(s, pkt.AckHandle, data)

}

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

func handleMsgMhfGetUdBonusQuestInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdBonusQuestInfo)

	udBonusQuestInfos := []struct {
		Unk0      uint8
		Unk1      uint8
		StartTime uint32 // Unix timestamp (seconds)
		EndTime   uint32 // Unix timestamp (seconds)
		Unk4      uint32
		Unk5      uint8
		Unk6      uint8
	}{} // Blank stub array.

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(udBonusQuestInfos)))
	for _, q := range udBonusQuestInfos {
		resp.WriteUint8(q.Unk0)
		resp.WriteUint8(q.Unk1)
		resp.WriteUint32(q.StartTime)
		resp.WriteUint32(q.EndTime)
		resp.WriteUint32(q.Unk4)
		resp.WriteUint8(q.Unk5)
		resp.WriteUint8(q.Unk6)
	}

	doSizedAckResp(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetUdSelectedColorInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdMonsterPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMonsterPoint)

	monsterPoints := []struct {
		MID    uint8
		Points uint16
	}{
		{MID: 0x01, Points: 0x3C}, // em1 Rathian
		{MID: 0x02, Points: 0x5A}, // em2 Fatalis
		{MID: 0x06, Points: 0x14}, // em6 Yian Kut-Ku
		{MID: 0x07, Points: 0x50}, // em7 Lao-Shan Lung
		{MID: 0x08, Points: 0x28}, // em8 Cephadrome
		{MID: 0x0B, Points: 0x3C}, // em11 Rathalos
		{MID: 0x0E, Points: 0x3C}, // em14 Diablos
		{MID: 0x0F, Points: 0x46}, // em15 Khezu
		{MID: 0x11, Points: 0x46}, // em17 Gravios
		{MID: 0x14, Points: 0x28}, // em20 Gypceros
		{MID: 0x15, Points: 0x3C}, // em21 Plesioth
		{MID: 0x16, Points: 0x32}, // em22 Basarios
		{MID: 0x1A, Points: 0x32}, // em26 Monoblos
		{MID: 0x1B, Points: 0x0A}, // em27 Velocidrome
		{MID: 0x1C, Points: 0x0A}, // em28 Gendrome
		{MID: 0x1F, Points: 0x0A}, // em31 Iodrome
		{MID: 0x21, Points: 0x50}, // em33 Kirin
		{MID: 0x24, Points: 0x64}, // em36 Crimson Fatalis
		{MID: 0x25, Points: 0x3C}, // em37 Pink Rathian
		{MID: 0x26, Points: 0x1E}, // em38 Blue Yian Kut-Ku
		{MID: 0x27, Points: 0x28}, // em39 Purple Gypceros
		{MID: 0x28, Points: 0x50}, // em40 Yian Garuga
		{MID: 0x29, Points: 0x5A}, // em41 Silver Rathalos
		{MID: 0x2A, Points: 0x50}, // em42 Gold Rathian
		{MID: 0x2B, Points: 0x3C}, // em43 Black Diablos
		{MID: 0x2C, Points: 0x3C}, // em44 White Monoblos
		{MID: 0x2D, Points: 0x46}, // em45 Red Khezu
		{MID: 0x2E, Points: 0x3C}, // em46 Green Plesioth
		{MID: 0x2F, Points: 0x50}, // em47 Black Gravios
		{MID: 0x30, Points: 0x1E}, // em48 Daimyo Hermitaur
		{MID: 0x31, Points: 0x3C}, // em49 Azure Rathalos
		{MID: 0x32, Points: 0x50}, // em50 Ashen Lao-Shan Lung
		{MID: 0x33, Points: 0x3C}, // em51 Blangonga
		{MID: 0x34, Points: 0x28}, // em52 Congalala
		{MID: 0x35, Points: 0x50}, // em53 Rajang
		{MID: 0x36, Points: 0x6E}, // em54 Kushala Daora
		{MID: 0x37, Points: 0x50}, // em55 Shen Gaoren
		{MID: 0x3A, Points: 0x50}, // em58 Yama Tsukami
		{MID: 0x3B, Points: 0x6E}, // em59 Chameleos
		{MID: 0x40, Points: 0x64}, // em64 Lunastra
		{MID: 0x41, Points: 0x6E}, // em65 Teostra
		{MID: 0x43, Points: 0x28}, // em67 Shogun Ceanataur
		{MID: 0x44, Points: 0x0A}, // em68 Bulldrome
		{MID: 0x47, Points: 0x6E}, // em71 White Fatalis
		{MID: 0x4A, Points: 0xFA}, // em74 Hypnocatrice
		{MID: 0x4B, Points: 0xFA}, // em75 Lavasioth
		{MID: 0x4C, Points: 0x46}, // em76 Tigrex
		{MID: 0x4D, Points: 0x64}, // em77 Akantor
		{MID: 0x4E, Points: 0xFA}, // em78 Bright Hypnoc
		{MID: 0x4F, Points: 0xFA}, // em79 Lavasioth Subspecies
		{MID: 0x50, Points: 0xFA}, // em80 Espinas
		{MID: 0x51, Points: 0xFA}, // em81 Orange Espinas
		{MID: 0x52, Points: 0xFA}, // em82 White Hypnoc
		{MID: 0x53, Points: 0xFA}, // em83 Akura Vashimu
		{MID: 0x54, Points: 0xFA}, // em84 Akura Jebia
		{MID: 0x55, Points: 0xFA}, // em85 Berukyurosu
		{MID: 0x59, Points: 0xFA}, // em89 Pariapuria
		{MID: 0x5A, Points: 0xFA}, // em90 White Espinas
		{MID: 0x5B, Points: 0xFA}, // em91 Kamu Orugaron
		{MID: 0x5C, Points: 0xFA}, // em92 Nono Orugaron
		{MID: 0x5E, Points: 0xFA}, // em94 Dyuragaua
		{MID: 0x5F, Points: 0xFA}, // em95 Doragyurosu
		{MID: 0x60, Points: 0xFA}, // em96 Gurenzeburu
		{MID: 0x63, Points: 0xFA}, // em99 Rukodiora
		{MID: 0x65, Points: 0xFA}, // em101 Gogomoa
		{MID: 0x67, Points: 0xFA}, // em103 Taikun Zamuza
		{MID: 0x68, Points: 0xFA}, // em104 Abiorugu
		{MID: 0x69, Points: 0xFA}, // em105 Kuarusepusu
		{MID: 0x6A, Points: 0xFA}, // em106 Odibatorasu
		{MID: 0x6B, Points: 0xFA}, // em107 Disufiroa
		{MID: 0x6C, Points: 0xFA}, // em108 Rebidiora
		{MID: 0x6D, Points: 0xFA}, // em109 Anorupatisu
		{MID: 0x6E, Points: 0xFA}, // em110 Hyujikiki
		{MID: 0x6F, Points: 0xFA}, // em111 Midogaron
		{MID: 0x70, Points: 0xFA}, // em112 Giaorugu
		{MID: 0x72, Points: 0xFA}, // em114 Farunokku
		{MID: 0x73, Points: 0xFA}, // em115 Pokaradon
		{MID: 0x74, Points: 0xFA}, // em116 Shantien
		{MID: 0x77, Points: 0xFA}, // em119 Goruganosu
		{MID: 0x78, Points: 0xFA}, // em120 Aruganosu
		{MID: 0x79, Points: 0xFA}, // em121 Baruragaru
		{MID: 0x7A, Points: 0xFA}, // em122 Zerureusu
		{MID: 0x7B, Points: 0xFA}, // em123 Gougarf
		{MID: 0x7D, Points: 0xFA}, // em125 Forokururu
		{MID: 0x7E, Points: 0xFA}, // em126 Meraginasu
		{MID: 0x7F, Points: 0xFA}, // em127 Diorekkusu
		{MID: 0x80, Points: 0xFA}, // em128 Garuba Daora
		{MID: 0x81, Points: 0xFA}, // em129 Inagami
		{MID: 0x82, Points: 0xFA}, // em130 Varusaburosu
		{MID: 0x83, Points: 0xFA}, // em131 Poborubarumu
		{MID: 0x8B, Points: 0xFA}, // em139 Gureadomosu
		{MID: 0x8C, Points: 0xFA}, // em140 Harudomerugu
		{MID: 0x8D, Points: 0xFA}, // em141 Toridcless
		{MID: 0x8E, Points: 0xFA}, // em142 Gasurabazura
		{MID: 0x90, Points: 0xFA}, // em144 Yama Kurai
		{MID: 0x92, Points: 0x78}, // em146 Zinogre
		{MID: 0x93, Points: 0x78}, // em147 Deviljho
		{MID: 0x94, Points: 0x78}, // em148 Brachydios
		{MID: 0x96, Points: 0xFA}, // em150 Toa Tesukatora
		{MID: 0x97, Points: 0x78}, // em151 Barioth
		{MID: 0x98, Points: 0x78}, // em152 Uragaan
		{MID: 0x99, Points: 0x78}, // em153 Stygian Zinogre
		{MID: 0x9A, Points: 0xFA}, // em154 Guanzorumu
		{MID: 0x9E, Points: 0xFA}, // em158 Voljang
		{MID: 0x9F, Points: 0x78}, // em159 Nargacuga
		{MID: 0xA0, Points: 0xFA}, // em160 Keoaruboru
		{MID: 0xA1, Points: 0xFA}, // em161 Zenaserisu
		{MID: 0xA2, Points: 0x78}, // em162 Gore Magala
		{MID: 0xA4, Points: 0x78}, // em164 Shagaru Magala
		{MID: 0xA5, Points: 0x78}, // em165 Amatsu
		{MID: 0xA6, Points: 0xFA}, // em166 Elzelion
		{MID: 0xA9, Points: 0x78}, // em169 Seregios
		{MID: 0xAA, Points: 0xFA}, // em170 Bogabadorumu
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

func handleMsgMhfGetEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEquipSkinHist)
	// Transmog / reskin system,  bitmask of 3200 bytes length
	// presumably divided by 5 sections for 5120 armour IDs covered
	// +10,000 for actual ID to be unlocked by each bit
	// Returning 3200 bytes of FF just unlocks everything for now
	doSizedAckResp(s, pkt.AckHandle, bytes.Repeat([]byte{0xFF}, 0xC80))
}

func handleMsgMhfUpdateEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateEquipSkinHist)
	// sends a raw armour ID back that needs to be mapped into the persistent bitmask above (-10,000)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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

func handleMsgMhfAddKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddKouryouPoint)
	// Adds pkt.KouryouPoints to the value in get kouryou points, not sure if the actual value is saved for sending in MsgMhfGetKouryouPoint or in SaveData
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

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
	var data []byte
	err := s.server.db.QueryRow("SELECT platemyset FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get presets sigil savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
	} else {
		blankData := make([]byte, 0x780)
		doSizedAckResp(s, pkt.AckHandle, blankData)
	}
}

func handleMsgMhfSavePlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateMyset)
	// looks to always return the full thing, simply update database, no extra processing
	_, err := s.server.db.Exec("UPDATE characters SET platemyset=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update platemyset savedata in db", zap.Error(err))
	}
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysReserve18B(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserve18B)

	// Left as raw bytes because I couldn't easily find the request or resp parser function in the binary.
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x3C})

}

func handleMsgMhfGetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve18E(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve18F(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTrendWeapon(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTrendWeapon)
	// TODO (Fist): Work out actual format limitations, seems to be final upgrade
	// for weapons and it traverses its upgrade tree to recommend base as final
	// 423C correlates with most popular magnet spike in use on JP
	// 2A 00 3C 44 00 3C 76 00 3F EA 01 0F 20 01 0F 50 01 0F F8 02 3C 7E 02 3D
	// F3 02 40 2A 03 3D 65 03 3F 2A 03 40 36 04 3D 59 04 41 E7 04 43 3E 05 0A
	// ED 05 0F 4C 05 0F F2 06 3A FE 06 41 E8 06 41 FA 07 3B 02 07 3F ED 07 40
	// 24 08 3D 37 08 3F 66 08 41 EC 09 3D 38 09 3F 8A 09 41 EE 0A 0E 78 0A 0F
	// AA 0A 0F F9 0B 3E 2E 0B 41 EF 0B 42 FB 0C 41 F0 0C 43 3F 0C 43 EE 0D 41 F1 0D 42 10 0D 42 3C 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
	doSizedAckResp(s, pkt.AckHandle, make([]byte, 0xA9))
}

func handleMsgMhfUpdateUseTrendWeaponLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateUseTrendWeaponLog)
	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysReserve192(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve193(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysReserve194(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSaveRengokuData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveRengokuData)
	_, err := s.server.db.Exec("UPDATE characters SET rengokudata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update rengokudata savedata in db", zap.Error(err))
	}

	s.QueueAck(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadRengokuData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadRengokuData)
	var data []byte
	err := s.server.db.QueryRow("SELECT rengokudata FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get rengokudata savedata from db", zap.Error(err))
	}
	if len(data) > 0 {
		doSizedAckResp(s, pkt.AckHandle, data)
	} else {

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
}

func handleMsgMhfGetRengokuBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRengokuBinary)
	// a (massively out of date) version resides in the game's /dat/ folder or up to date can be pulled from packets
	data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, fmt.Sprintf("rengoku_data.bin")))
	if err != nil {
		panic(err)
	}

	doSizedAckResp(s, pkt.AckHandle, data)

}

func handleMsgMhfEnumerateRengokuRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRengokuRanking)
	doSizedAckResp(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

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
