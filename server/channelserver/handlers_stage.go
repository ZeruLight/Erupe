package channelserver

import (
	"fmt"
	"strings"
	"time"

	"erupe-ce/internal/system"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"

	ps "erupe-ce/utils/pascalstring"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func handleMsgSysCreateStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateStage)
	s.Server.Lock()
	defer s.Server.Unlock()
	if _, exists := s.Server.stages[pkt.StageID]; exists {
		s.DoAckSimpleFail(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	} else {
		stage := system.NewStage(pkt.StageID)
		stage.Host = s
		stage.MaxPlayers = uint16(pkt.PlayerCount)
		s.Server.stages[stage.Id] = stage
		s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgSysStageDestruct(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func doStageTransfer(s *Session, ackHandle uint32, stageID string) {
	s.Server.Lock()
	stage, exists := s.Server.stages[stageID]
	s.Server.Unlock()

	if exists {
		stage.Lock()
		stage.Clients[s] = s.CharID
		stage.Unlock()
	} else { // Create new stage object
		s.Server.Lock()
		s.Server.stages[stageID] = system.NewStage(stageID)
		stage = s.Server.stages[stageID]
		s.Server.Unlock()
		stage.Lock()
		stage.Host = s
		stage.Clients[s] = s.CharID
		stage.Unlock()
	}

	// Ensure this session no longer belongs to reservations.
	if s.stage != nil {
		removeSessionFromStage(s)
	}

	// Save our new stage ID and pointer to the new stage itself.
	s.Lock()
	s.stage = s.Server.stages[stageID]
	s.Unlock()

	// Tell the client to cleanup its current stage objects.
	s.QueueSendMHFLazy(&mhfpacket.MsgSysCleanupObject{})

	// Confirm the stage entry.
	s.DoAckSimpleSucceed(ackHandle, []byte{0x00, 0x00, 0x00, 0x00})

	var temp mhfpacket.MHFPacket

	// Cast existing user data to new user
	if !s.userEnteredStage {
		s.userEnteredStage = true

		for _, session := range s.Server.sessions {
			if s == session {
				continue
			}
			temp = &mhfpacket.MsgSysInsertUser{CharID: session.CharID}
			s.QueueSendMHFLazy(temp)
			for i := 0; i < 3; i++ {
				temp = &mhfpacket.MsgSysNotifyUserBinary{
					CharID:     session.CharID,
					BinaryType: uint8(i + 1),
				}
				s.QueueSendMHFLazy(temp)
			}
		}
	}

	if s.stage != nil { // avoids lock up when using bed for dream quests
		// Notify the client to duplicate the existing objects.
		s.Logger.Info(fmt.Sprintf("Sending existing stage objects to %s", s.Name))
		s.stage.RLock()
		for _, obj := range s.stage.Objects {
			if obj.OwnerCharID == s.CharID {
				continue
			}
			temp = &mhfpacket.MsgSysDuplicateObject{
				ObjID:       obj.Id,
				X:           obj.X,
				Y:           obj.Y,
				Z:           obj.Z,
				Unk0:        0,
				OwnerCharID: obj.OwnerCharID,
			}
			s.QueueSendMHFLazy(temp)
		}
		s.stage.RUnlock()
	}
}

func destructEmptyStages(s *Session) {
	s.Server.Lock()
	defer s.Server.Unlock()
	for _, stage := range s.Server.stages {
		// Destroy empty Quest/My series/Guild stages.
		if stage.Id[3:5] == "Qs" || stage.Id[3:5] == "Ms" || stage.Id[3:5] == "Gs" || stage.Id[3:5] == "Ls" {
			if len(stage.ReservedClientSlots) == 0 && len(stage.Clients) == 0 {
				delete(s.Server.stages, stage.Id)
				s.Logger.Debug("Destructed stage", zap.String("stage.id", stage.Id))
			}
		}
	}
}

func removeSessionFromStage(s *Session) {
	// Remove client from old stage.
	delete(s.stage.Clients, s)

	// Delete old stage objects owned by the client.
	s.Logger.Info("Sending notification to old stage clients")
	for _, object := range s.stage.Objects {
		if object.OwnerCharID == s.CharID {
			s.stage.BroadcastMHF(&mhfpacket.MsgSysDeleteObject{ObjID: object.Id}, s)
			delete(s.stage.Objects, object.OwnerCharID)
		}
	}
	destructEmptyStages(s)
	destructEmptySemaphores(s)
}

func isStageFull(s *Session, StageID string) bool {
	if stage, exists := s.Server.stages[StageID]; exists {
		if _, exists := stage.ReservedClientSlots[s.CharID]; exists {
			return false
		}
		return len(stage.ReservedClientSlots)+len(stage.Clients) >= int(stage.MaxPlayers)
	}
	return false
}

func handleMsgSysEnterStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnterStage)

	if isStageFull(s, pkt.StageID) {
		s.DoAckSimpleFail(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
		return
	}

	// Push our current stage ID to the movement stack before entering another one.
	if s.stage != nil {
		s.stage.Lock()
		s.stage.ReservedClientSlots[s.CharID] = false
		s.stage.Unlock()
		s.stageMoveStack.Push(s.stage.Id)
	}

	if s.reservationStage != nil {
		s.reservationStage = nil
	}

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysBackStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysBackStage)

	// Transfer back to the saved stage ID before the previous move or enter.
	backStage, err := s.stageMoveStack.Pop()
	if backStage == "" || err != nil {
		backStage = "sl1Ns200p0a0u0"
	}

	if isStageFull(s, backStage) {
		s.stageMoveStack.Push(backStage)
		s.DoAckSimpleFail(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
		return
	}

	if _, exists := s.stage.ReservedClientSlots[s.CharID]; exists {
		delete(s.stage.ReservedClientSlots, s.CharID)
	}

	if _, exists := s.Server.stages[backStage].ReservedClientSlots[s.CharID]; exists {
		delete(s.Server.stages[backStage].ReservedClientSlots, s.CharID)
	}

	doStageTransfer(s, pkt.AckHandle, backStage)
}

func handleMsgSysMoveStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysMoveStage)

	if isStageFull(s, pkt.StageID) {
		s.DoAckSimpleFail(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
		return
	}

	doStageTransfer(s, pkt.AckHandle, pkt.StageID)
}

func handleMsgSysLeaveStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysLockStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLockStage)
	if stage, exists := s.Server.stages[pkt.StageID]; exists {
		stage.Lock()
		stage.Locked = true
		stage.Unlock()
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgSysUnlockStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	if s.reservationStage != nil {
		s.reservationStage.RLock()
		defer s.reservationStage.RUnlock()

		for charID := range s.reservationStage.ReservedClientSlots {
			session := s.Server.FindSessionByCharID(charID)
			if session != nil {
				session.QueueSendMHFLazy(&mhfpacket.MsgSysStageDestruct{})
			}
		}

		delete(s.Server.stages, s.reservationStage.Id)
	}

	destructEmptyStages(s)
}

func handleMsgSysReserveStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysReserveStage)
	if stage, exists := s.Server.stages[pkt.StageID]; exists {
		stage.Lock()
		defer stage.Unlock()
		if _, exists := stage.ReservedClientSlots[s.CharID]; exists {
			switch pkt.Ready {
			case 1: // 0x01
				stage.ReservedClientSlots[s.CharID] = false
			case 17: // 0x11
				stage.ReservedClientSlots[s.CharID] = true
			}
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		} else if uint16(len(stage.ReservedClientSlots)) < stage.MaxPlayers {
			if stage.Locked {
				s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
				return
			}
			if len(stage.Password) > 0 {
				if stage.Password != s.stagePass {
					s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
					return
				}
			}
			stage.ReservedClientSlots[s.CharID] = false
			// Save the reservation stage in the session for later use in MsgSysUnreserveStage.
			s.Lock()
			s.reservationStage = stage
			s.Unlock()
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		} else {
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		}
	} else {
		s.Logger.Error("Failed to get stage", zap.String("StageID", pkt.StageID))
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgSysUnreserveStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	s.Lock()
	stage := s.reservationStage
	s.reservationStage = nil
	s.Unlock()
	if stage != nil {
		stage.Lock()
		if _, exists := stage.ReservedClientSlots[s.CharID]; exists {
			delete(stage.ReservedClientSlots, s.CharID)
		}
		stage.Unlock()
	}
}

func handleMsgSysSetStagePass(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStagePass)
	s.Lock()
	stage := s.reservationStage
	s.Unlock()
	if stage != nil {
		stage.Lock()
		// Will only exist if host.
		if _, exists := stage.ReservedClientSlots[s.CharID]; exists {
			stage.Password = pkt.Password
		}
		stage.Unlock()
	} else {
		// Store for use on next ReserveStage.
		s.Lock()
		s.stagePass = pkt.Password
		s.Unlock()
	}
}

func handleMsgSysSetStageBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetStageBinary)
	if stage, exists := s.Server.stages[pkt.StageID]; exists {
		stage.Lock()
		stage.RawBinaryData[system.StageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}] = pkt.RawDataPayload
		stage.Unlock()
	} else {
		s.Logger.Warn("Failed to get stage", zap.String("StageID", pkt.StageID))
	}
}

func handleMsgSysGetStageBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetStageBinary)
	if stage, exists := s.Server.stages[pkt.StageID]; exists {
		stage.Lock()
		if binaryData, exists := stage.RawBinaryData[system.StageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]; exists {
			s.DoAckBufSucceed(pkt.AckHandle, binaryData)
		} else if pkt.BinaryType1 == 4 {
			// Unknown binary type that is supposedly generated server side
			// Temporary response
			s.DoAckBufSucceed(pkt.AckHandle, []byte{})
		} else {
			s.Logger.Warn("Failed to get stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
			s.Logger.Warn("Sending blank stage binary")
			s.DoAckBufSucceed(pkt.AckHandle, []byte{})
		}
		stage.Unlock()
	} else {
		s.Logger.Warn("Failed to get stage", zap.String("StageID", pkt.StageID))
	}
	s.Logger.Debug("MsgSysGetStageBinary Done!")
}

func handleMsgSysWaitStageBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysWaitStageBinary)
	if stage, exists := s.Server.stages[pkt.StageID]; exists {
		if pkt.BinaryType0 == 1 && pkt.BinaryType1 == 12 {
			// This might contain the hunter count, or max player count?
			s.DoAckBufSucceed(pkt.AckHandle, []byte{0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
			return
		}
		for {
			s.Logger.Debug("MsgSysWaitStageBinary before lock and get stage")
			stage.Lock()
			stageBinary, gotBinary := stage.RawBinaryData[system.StageBinaryKey{pkt.BinaryType0, pkt.BinaryType1}]
			stage.Unlock()
			s.Logger.Debug("MsgSysWaitStageBinary after lock and get stage")
			if gotBinary {
				s.DoAckBufSucceed(pkt.AckHandle, stageBinary)
				break
			} else {
				s.Logger.Debug("Waiting stage binary", zap.Uint8("BinaryType0", pkt.BinaryType0), zap.Uint8("pkt.BinaryType1", pkt.BinaryType1))
				time.Sleep(1 * time.Second)
				continue
			}
		}
	} else {
		s.Logger.Warn("Failed to get stage", zap.String("StageID", pkt.StageID))
	}
	s.Logger.Debug("MsgSysWaitStageBinary Done!")
}

func handleMsgSysEnumerateStage(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysEnumerateStage)

	// Read-lock the server stage map.
	s.Server.stagesLock.RLock()
	defer s.Server.stagesLock.RUnlock()

	// Build the response
	bf := byteframe.NewByteFrame()
	var joinable uint16
	bf.WriteUint16(0)
	for sid, stage := range s.Server.stages {
		stage.RLock()

		if len(stage.ReservedClientSlots) == 0 && len(stage.Clients) == 0 {
			stage.RUnlock()
			continue
		}
		if !strings.Contains(stage.Id, pkt.StagePrefix) {
			stage.RUnlock()
			continue
		}
		joinable++

		bf.WriteUint16(uint16(len(stage.ReservedClientSlots)))
		bf.WriteUint16(uint16(len(stage.Clients)))
		if strings.HasPrefix(stage.Id, "sl2Ls") {
			bf.WriteUint16(uint16(len(stage.Clients) + len(stage.ReservedClientSlots)))
		} else {
			bf.WriteUint16(uint16(len(stage.Clients)))
		}
		bf.WriteUint16(stage.MaxPlayers)
		var flags uint8
		if stage.Locked {
			flags |= 1
		}
		if len(stage.Password) > 0 {
			flags |= 2
		}
		bf.WriteUint8(flags)
		ps.Uint8(bf, sid, false)
		stage.RUnlock()
	}
	bf.Seek(0, 0)
	bf.WriteUint16(joinable)

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}
