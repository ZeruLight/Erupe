package channelserver

import (
	"erupe-ce/internal/service"
	"erupe-ce/utils/stringsupport"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"

	"github.com/jmoiron/sqlx"
)

func handleMsgMhfReadMail(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMail)

	mailId := s.mailList[pkt.AccIndex]
	if mailId == 0 {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0})
		return
	}

	mail, err := service.GetMailByID(mailId)
	if err != nil {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0})
		return
	}

	db.Exec(`UPDATE mail SET read = true WHERE id = $1`, mail.ID)
	bf := byteframe.NewByteFrame()
	body := stringsupport.UTF8ToSJIS(mail.Body)
	bf.WriteNullTerminatedBytes(body)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfListMail(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfListMail)

	mail, err := service.GetMailListForCharacter(s.CharID)
	if err != nil {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0})
		return
	}

	if s.mailList == nil {
		s.mailList = make([]int, 256)
	}

	msg := byteframe.NewByteFrame()

	msg.WriteUint32(uint32(len(mail)))

	startIndex := s.mailAccIndex

	for i, m := range mail {
		accIndex := startIndex + uint8(i)
		s.mailList[accIndex] = m.ID
		s.mailAccIndex++

		itemAttached := m.AttachedItemID != 0

		msg.WriteUint32(m.SenderID)
		msg.WriteUint32(uint32(m.CreatedAt.Unix()))

		msg.WriteUint8(accIndex)
		msg.WriteUint8(uint8(i))

		flags := uint8(0x00)

		if m.Read {
			flags |= 0x01
		}

		if m.Locked {
			flags |= 0x02
		}

		if m.IsSystemMessage {
			flags |= 0x04
		}

		if m.AttachedItemReceived {
			flags |= 0x08
		}

		if m.IsGuildInvite {
			flags |= 0x10
		}

		msg.WriteUint8(flags)
		msg.WriteBool(itemAttached)
		msg.WriteUint8(16)
		msg.WriteUint8(21)
		msg.WriteBytes(stringsupport.PaddedString(m.Subject, 16, true))
		msg.WriteBytes(stringsupport.PaddedString(m.SenderName, 21, true))
		if itemAttached {
			msg.WriteUint16(m.AttachedItemAmount)
			msg.WriteUint16(m.AttachedItemID)
		}
	}

	s.DoAckBufSucceed(pkt.AckHandle, msg.Data())
}

func handleMsgMhfOprtMail(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOprtMail)

	mail, err := service.GetMailByID(s.mailList[pkt.AccIndex])
	if err != nil {
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		return
	}

	switch pkt.Operation {
	case mhfpacket.OperateMailDelete:
		db.Exec(`UPDATE mail SET deleted = true WHERE id = $1`, mail.ID)
	case mhfpacket.OperateMailLock:
		db.Exec(`UPDATE mail SET locked = TRUE WHERE id = $1`, mail.ID)
	case mhfpacket.OperateMailUnlock:
		db.Exec(`UPDATE mail SET locked = FALSE WHERE id = $1`, mail.ID)
	case mhfpacket.OperateMailAcquireItem:
		db.Exec(`UPDATE mail SET attached_item_received = TRUE WHERE id = $1`, mail.ID)
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfSendMail(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSendMail)

	query := `
		INSERT INTO mail (sender_id, recipient_id, subject, body, attached_item, attached_item_amount, is_guild_invite)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if pkt.RecipientID == 0 { // Guild mail
		g, err := GetGuildInfoByCharacterId(s, s.CharID)
		if err != nil {
			s.Logger.Error("Failed to get guild info for mail")
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
			return
		}
		gm, err := GetGuildMembers(s, g.ID, false)
		if err != nil {
			s.Logger.Error("Failed to get guild members for mail")
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
			return
		}
		for i := 0; i < len(gm); i++ {
			_, err := db.Exec(query, s.CharID, gm[i].CharID, pkt.Subject, pkt.Body, 0, 0, false)
			if err != nil {
				s.Logger.Error("Failed to send mail")
				s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
				return
			}
		}
	} else {
		_, err := db.Exec(query, s.CharID, pkt.RecipientID, pkt.Subject, pkt.Body, pkt.ItemID, pkt.Quantity, false)
		if err != nil {
			s.Logger.Error("Failed to send mail")
		}
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}
