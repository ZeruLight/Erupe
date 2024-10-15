package channelserver

import (
	"fmt"

	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"

	"github.com/jmoiron/sqlx"
)

func handleMsgSysCreateObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCreateObject)

	s.stage.Lock()
	newObj := &Object{
		id:          s.NextObjectID(),
		ownerCharID: s.CharID,
		x:           pkt.X,
		y:           pkt.Y,
		z:           pkt.Z,
	}
	s.stage.objects[s.CharID] = newObj
	s.stage.Unlock()

	// Response to our requesting client.
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(newObj.id) // New local obj handle.
	s.DoAckSimpleSucceed(pkt.AckHandle, resp.Data())
	// Duplicate the object creation to all sessions in the same stage.
	dupObjUpdate := &mhfpacket.MsgSysDuplicateObject{
		ObjID:       newObj.id,
		X:           newObj.x,
		Y:           newObj.y,
		Z:           newObj.z,
		OwnerCharID: newObj.ownerCharID,
	}

	s.Logger.Info(fmt.Sprintf("Broadcasting new object: %s (%d)", s.Name, newObj.id))
	s.stage.BroadcastMHF(dupObjUpdate, s)
}

func handleMsgSysDeleteObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysPositionObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysPositionObject)
	if config.GetConfig().DebugOptions.LogInboundMessages {
		fmt.Printf("[%s] with objectID [%d] move to (%f,%f,%f)\n\n", s.Name, pkt.ObjID, pkt.X, pkt.Y, pkt.Z)
	}
	s.stage.Lock()
	object, ok := s.stage.objects[s.CharID]
	if ok {
		object.x = pkt.X
		object.y = pkt.Y
		object.z = pkt.Z
	}
	s.stage.Unlock()
	// One of the few packets we can just re-broadcast directly.
	s.stage.BroadcastMHF(pkt, s)
}

func handleMsgSysRotateObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysDuplicateObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysSetObjectBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	_ = p.(*mhfpacket.MsgSysSetObjectBinary)
	/* This causes issues with PS3 as this actually sends with endiness!
	for _, session := range s.Server.sessions {
		if session.CharID == s.CharID {
			s.Server.userBinaryPartsLock.Lock()
			s.Server.userBinaryParts[userBinaryPartID{charID: s.CharID, index: 3}] = pkt.RawDataPayload
			s.Server.userBinaryPartsLock.Unlock()
			msg := &mhfpacket.MsgSysNotifyUserBinary{
				CharID:     s.CharID,
				BinaryType: 3,
			}
			s.Server.BroadcastMHF(msg, s)
		}
	}
	*/
}

func handleMsgSysGetObjectBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysGetObjectOwner(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysUpdateObjectBinary(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysCleanupObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysAddObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysDelObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysDispObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgSysHideObject(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}
