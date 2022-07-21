package channelserver

import (
	"fmt"
	"time"
	"strings"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/common/byteframe"
	"go.uber.org/zap"
)

func handleMsgSysCreateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateStage)
	s.server.Lock()
	defer s.server.Unlock()
	if _, exists := s.server.stages[pkt.StageID]; exists {
    doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	} else {
		stage := NewStage(pkt.StageID)
		stage.maxPlayers = uint16(pkt.PlayerCount)
		s.server.stages[stage.id] = stage
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgSysStageDestruct(s *Session, p mhfpacket.MHFPacket) {}

func doStageTransfer(s *Session, ackHandle uint32, stageID string) {
	// Remove this session from old stage clients list and put myself in the new one.
	s.server.stagesLock.Lock()
	newStage, gotNewStage := s.server.stages[stageID]
	s.server.stagesLock.Unlock()

	if s.stage != nil {
		removeSessionFromStage(s)
	}

	// Add the new stage.
	if gotNewStage {
		newStage.Lock()
		newStage.clients[s] = s.charID
		newStage.Unlock()
	} else {
		// Fix stages
		s.logger.Info("Fix Map Appliqued")
		s.server.stagesLock.Lock()
		s.server.stages[stageID] = NewStage(stageID)
		newStage = s.server.stages[stageID]
		s.server.stagesLock.Unlock()
		newStage.Lock()
		newStage.clients[s] = s.charID
		newStage.Unlock()
	}

	// Save our new stage ID and pointer to the new stage itself.
	s.Lock()
	s.stageID = string(stageID)
	s.stage = newStage
	s.Unlock()

	// Tell the client to cleanup its current stage objects.
	s.QueueSendMHF(&mhfpacket.MsgSysCleanupObject{})

	// Confirm the stage entry.
	doAckSimpleSucceed(s, ackHandle, []byte{0x00, 0x00, 0x00, 0x00})

	// Notify existing stage clients that this new client has entered.
	if s.stage != nil { // avoids lock up when using bed for dream quests
		var pkt mhfpacket.MHFPacket

		// Notify the entree client about all of the existing clients in the stage.
		s.logger.Info("Notifying entree about existing stage clients")

		clientNotif := byteframe.NewByteFrame()
		s.server.Lock()
		s.server.BroadcastMHF(&mhfpacket.MsgSysDeleteUser{
			CharID: s.charID,
		}, s)
		s.server.BroadcastMHF(&mhfpacket.MsgSysInsertUser {
			CharID: s.charID,
		}, s)

		for session := range s.server.sessions {
			session := s.server.sessions[session]

			// Send existing players back to the client
			pkt = &mhfpacket.MsgSysInsertUser{
				CharID: session.charID,
			}
			clientNotif.WriteUint16(uint16(pkt.Opcode()))
			pkt.Build(clientNotif, session.clientContext)
			for i := 1; i <= 3; i++ {
				pkt = &mhfpacket.MsgSysNotifyUserBinary{
					CharID:     session.charID,
					BinaryType: uint8(i),
				}
				clientNotif.WriteUint16(uint16(pkt.Opcode()))
				pkt.Build(clientNotif, session.clientContext)
			}
		}
		s.server.Unlock()
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
				cur.Build(clientDupObjNotif, s.clientContext)
		}
		s.stage.RUnlock()
		clientDupObjNotif.WriteUint16(0x0010) // End it.
		s.QueueSend(clientDupObjNotif.Data())
	}
}

func removeEmptyStages(s *Session) {
	s.server.Lock()
	for sid, stage := range s.server.stages {
		if len(stage.reservedClientSlots) == 0 && len(stage.clients) == 0 {
			if strings.HasPrefix(sid, "sl1Qs") || strings.HasPrefix(sid, "sl2Qs") || strings.HasPrefix(sid, "sl3Qs") {
				delete(s.server.stages, sid)
			}
		}
	}
	s.server.Unlock()
}

func removeSessionFromStage(s *Session) {
	s.stage.Lock()
	defer s.stage.Unlock()

	// Remove client from old stage.
	delete(s.stage.clients, s)
	delete(s.stage.reservedClientSlots, s.charID)

	// Remove client from all reservations
	s.server.Lock()
	for _, stage := range s.server.stages {
		if _, exists := stage.reservedClientSlots[s.charID]; exists {
			delete(stage.reservedClientSlots, s.charID)
		}
	}
	s.server.Unlock()

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
	for objListID, stageObjectList := range s.stage.objectList {
		if stageObjectList.charid == s.charID {
			//Added to prevent duplicates from flooding ObjectMap and causing server hangs
			s.stage.objectList[objListID].status=false
			s.stage.objectList[objListID].charid=0
		}
	}

	removeEmptyStages(s)
}


func handleMsgSysEnterStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnterStage)
	fmt.Printf("The Stage is %s\n",pkt.StageID)
	// Push our current stage ID to the movement stack before entering another one.
	s.Lock()
	s.stageMoveStack.Push(s.stageID)
	s.Unlock()

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysBackStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysBackStage)

	if s.stage != nil {
		removeSessionFromStage(s)
	}

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

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysLeaveStage(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLockStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLockStage)
	// TODO(Andoryuuta): What does this packet _actually_ do?
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysUnlockStage(s *Session, p mhfpacket.MHFPacket) {
	s.reservationStage.RLock()
	defer s.reservationStage.RUnlock()

	destructMessage := &mhfpacket.MsgSysStageDestruct{}

	for charID := range s.reservationStage.reservedClientSlots {
		session := s.server.FindSessionByCharID(charID)
		session.QueueSendMHF(destructMessage)
	}

	s.server.Lock()
	defer s.server.Unlock()

	delete(s.server.stages, s.reservationStage.id)
}

func handleMsgSysReserveStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserveStage)
	if stage, exists := s.server.stages[pkt.StageID]; exists {
		stage.Lock()
		defer stage.Unlock()
		if _, exists := stage.reservedClientSlots[s.charID]; exists {
			switch pkt.Ready {
			case 1: // 0x01
				stage.reservedClientSlots[s.charID] = false
			case 17: // 0x11
				stage.reservedClientSlots[s.charID] = true
			}
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else if uint16(len(stage.reservedClientSlots)) < stage.maxPlayers {
			if len(stage.password) > 0 {
				// s.logger.Debug("", zap.String("stgpw", stage.password), zap.String("usrpw", s.stagePass))
				if stage.password == s.stagePass {
					stage.reservedClientSlots[s.charID] = false
					s.Lock()
					s.reservationStage = stage
					s.Unlock()
					doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
					return
				}
				doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			} else {
				stage.reservedClientSlots[s.charID] = false
				// Save the reservation stage in the session for later use in MsgSysUnreserveStage.
				s.Lock()
				s.reservationStage = stage
				s.Unlock()
				doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			}
		} else {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		}
	} else {
		s.logger.Error("Failed to get stage", zap.String("StageID", pkt.StageID))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgSysUnreserveStage(s *Session, p mhfpacket.MHFPacket) {
	s.Lock()
	stage := s.reservationStage
	s.reservationStage = nil
	s.Unlock()
	if stage != nil {
		stage.Lock()
		if _, exists := stage.reservedClientSlots[s.charID]; exists {
			delete(stage.reservedClientSlots, s.charID)
		}
		stage.Unlock()
	}
}

func handleMsgSysSetStagePass(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStagePass)
	s.Lock()
	stage := s.reservationStage
	s.Unlock()
	if stage != nil {
		stage.Lock()
		if _, exists := stage.reservedClientSlots[s.charID]; exists {
			stage.password = pkt.Password
		}
		stage.Unlock()
	} else {
		// Store for use on next ReserveStage
		s.Lock()
		s.stagePass = pkt.Password
		s.Unlock()
	}
}

func handleMsgSysSetStageBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStageBinary)

	// Try to get the stage
	stageID := pkt.StageID
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
	stageID := pkt.StageID
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
		doAckBufSucceed(s, pkt.AckHandle, stageBinary)

	} else if pkt.BinaryType1 == 4 {
		// This particular type seems to be expecting data that isn't set
		// is it required before the party joining can be completed
		//s.QueueAck(pkt.AckHandle, []byte{0x01, 0x00, 0x00, 0x00, 0x10})

		// TODO(Andoryuuta): This doesn't fit a normal ack packet? where is this from?
		// This would be a buffered(0x01), non-error(0x00), with no data payload (size 0x00, 0x00) packet.
		// but for some reason has a 0x10 on the end that the client shouldn't parse?

		doAckBufSucceed(s, pkt.AckHandle, []byte{}) // Without the previous 0x10 suffix
	} else {
		s.logger.Warn("Failed to get stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
		s.logger.Warn("Sending blank stage binary")
		doAckBufSucceed(s, pkt.AckHandle, []byte{})
	}

	s.logger.Debug("MsgSysGetStageBinary Done!")
}

func handleMsgSysEnumerateStage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateStage)

	// Read-lock the server stage map.
	s.server.stagesLock.RLock()
	defer s.server.stagesLock.RUnlock()

	// Build the response
	resp := byteframe.NewByteFrame()
	bf := byteframe.NewByteFrame()
	var joinable int
	for sid, stage := range s.server.stages {
		stage.RLock()
		defer stage.RUnlock()

		if len(stage.reservedClientSlots) == 0 && len(stage.clients) == 0 {
			continue
		}

		// Check for valid stage type
		if sid[3:5] != "Qs" && sid[3:5] != "Ms" {
			continue
		}

		joinable++

		resp.WriteUint16(uint16(len(stage.reservedClientSlots))) // Reserved players.
		resp.WriteUint16(0)                    // Unknown value

		var hasDeparted uint16
		if stage.hasDeparted {
			hasDeparted = 1
		}
		resp.WriteUint16(hasDeparted)           // HasDeparted.
		resp.WriteUint16(stage.maxPlayers)      // Max players.
		if len(stage.password) > 0 {
			// This byte has also been seen as 1
			// The quest is also recognised as locked when this is 2
			resp.WriteUint8(3)
		} else {
			resp.WriteUint8(0)
		}
		ps.Uint8(resp, sid, false)
	}
	bf.WriteUint16(uint16(joinable))
	bf.WriteBytes(resp.Data())

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
