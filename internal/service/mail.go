package service

import (
	"database/sql"
	"erupe-ce/internal/constant"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"

	"erupe-ce/utils/logger"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Mail struct {
	ID                   int       `db:"id"`
	SenderID             uint32    `db:"sender_id"`
	RecipientID          uint32    `db:"recipient_id"`
	Subject              string    `db:"subject"`
	Body                 string    `db:"body"`
	Read                 bool      `db:"read"`
	Deleted              bool      `db:"deleted"`
	Locked               bool      `db:"locked"`
	AttachedItemReceived bool      `db:"attached_item_received"`
	AttachedItemID       uint16    `db:"attached_item"`
	AttachedItemAmount   uint16    `db:"attached_item_amount"`
	CreatedAt            time.Time `db:"created_at"`
	IsGuildInvite        bool      `db:"is_guild_invite"`
	IsSystemMessage      bool      `db:"is_sys_message"`
	SenderName           string    `db:"sender_name"`
}

func (m *Mail) Send(transaction *sql.Tx) error {
	db, err := db.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	query := `
		INSERT INTO mail (sender_id, recipient_id, subject, body, attached_item, attached_item_amount, is_guild_invite, is_sys_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	if transaction == nil {
		_, err = db.Exec(query, m.SenderID, m.RecipientID, m.Subject, m.Body, m.AttachedItemID, m.AttachedItemAmount, m.IsGuildInvite, m.IsSystemMessage)
	} else {
		_, err = transaction.Exec(query, m.SenderID, m.RecipientID, m.Subject, m.Body, m.AttachedItemID, m.AttachedItemAmount, m.IsGuildInvite, m.IsSystemMessage)
	}

	if err != nil {
		logger.Error(
			"failed to send mail",
			zap.Error(err),
			zap.Uint32("senderID", m.SenderID),
			zap.Uint32("recipientID", m.RecipientID),
			zap.String("subject", m.Subject),
			zap.String("body", m.Body),
			zap.Uint16("itemID", m.AttachedItemID),
			zap.Uint16("itemAmount", m.AttachedItemAmount),
			zap.Bool("isGuildInvite", m.IsGuildInvite),
			zap.Bool("isSystemMessage", m.IsSystemMessage),
		)
		return err
	}

	return nil
}

func (m *Mail) MarkRead() error {
	db, err := db.GetDB()
	logger := logger.Get()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_, err = db.Exec(`
		UPDATE mail SET read = true WHERE id = $1
	`, m.ID)

	if err != nil {
		logger.Error(
			"failed to mark mail as read",
			zap.Error(err),
			zap.Int("mailID", m.ID),
		)
		return err
	}

	return nil
}

func GetMailListForCharacter(charID uint32) ([]Mail, error) {
	db, err := db.GetDB()
	logger := logger.Get()
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	rows, err := db.Queryx(`
		SELECT
			m.id,
			m.sender_id,
			m.recipient_id,
			m.subject,
			m.read,
			m.attached_item_received,
			m.attached_item,
			m.attached_item_amount,
			m.created_at,
			m.is_guild_invite,
			m.is_sys_message,
			m.deleted,
			m.locked,
			c.name as sender_name
		FROM mail m
			JOIN characters c ON c.id = m.sender_id
		WHERE recipient_id = $1 AND m.deleted = false
		ORDER BY m.created_at DESC, id DESC
		LIMIT 32
	`, charID)

	if err != nil {
		logger.Error("failed to get mail for character", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}

	defer rows.Close()

	allMail := make([]Mail, 0)

	for rows.Next() {
		mail := Mail{}

		err := rows.StructScan(&mail)

		if err != nil {
			return nil, err
		}

		allMail = append(allMail, mail)
	}

	return allMail, nil
}

func GetMailByID(ID int) (*Mail, error) {
	db, err := db.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	row := db.QueryRowx(`
		SELECT
			m.id,
			m.sender_id,
			m.recipient_id,
			m.subject,
			m.read,
			m.body,
			m.attached_item_received,
			m.attached_item,
			m.attached_item_amount,
			m.created_at,
			m.is_guild_invite,
			m.is_sys_message,
			m.deleted,
			m.locked,
			c.name as sender_name
		FROM mail m
			JOIN characters c ON c.id = m.sender_id
		WHERE m.id = $1
		LIMIT 1
	`, ID)

	mail := &Mail{}

	err = row.StructScan(mail)

	if err != nil {
		logger.Error(
			"failed to retrieve mail",
			zap.Error(err),
			zap.Int("mailID", ID),
		)
		return nil, err
	}

	return mail, nil
}

type SessionMail interface {
	QueueSendMHF(packet mhfpacket.MHFPacket)
}

func SendMailNotification(s SessionMail, m *Mail, recipient SessionMail) {
	bf := byteframe.NewByteFrame()

	notification := &binpacket.MsgBinMailNotify{
		SenderName: getCharacterName(m.SenderID),
	}

	notification.Build(bf)

	castedBinary := &mhfpacket.MsgSysCastedBinary{
		CharID:         m.SenderID,
		BroadcastType:  0x00,
		MessageType:    constant.BinaryMessageTypeMailNotify,
		RawDataPayload: bf.Data(),
	}

	castedBinary.Build(bf)

	recipient.QueueSendMHF(castedBinary)
}

func getCharacterName(charID uint32) string {
	db, err := db.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	row := db.QueryRow("SELECT name FROM characters WHERE id = $1", charID)

	charName := ""

	err = row.Scan(&charName)

	if err != nil {
		return ""
	}
	return charName
}
