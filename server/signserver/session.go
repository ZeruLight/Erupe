package signserver

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"sync"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Client int

const (
	PC100 Client = iota
	VITA
	PS3
)

// Session holds state for the sign server connection.
type Session struct {
	sync.Mutex
	logger    *zap.Logger
	server    *Server
	rawConn   net.Conn
	cryptConn *network.CryptConn
	client    Client
}

func (s *Session) work() {
	pkt, err := s.cryptConn.ReadPacket()

	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.LogInboundMessages {
		fmt.Printf("\n[Client] -> [Server]\nData [%d bytes]:\n%s\n", len(pkt), hex.Dump(pkt))
	}

	if err != nil {
		return
	}
	err = s.handlePacket(pkt)
	if err != nil {
		return
	}
}

func (s *Session) handlePacket(pkt []byte) error {
	bf := byteframe.NewByteFrameFromBytes(pkt)
	reqType := string(bf.ReadNullTerminatedBytes())
	switch reqType {
	case "DLTSKEYSIGN:100", "DSGN:100":
		s.handleDSGN(bf)
	case "PS3SGN:100":
		s.client = PS3
		s.handlePSSGN(bf)
	case "VITASGN:100":
		s.client = VITA
		s.handlePSSGN(bf)
	case "DELETE:100":
		loginTokenString := string(bf.ReadNullTerminatedBytes())
		characterID := int(bf.ReadUint32())
		_ = int(bf.ReadUint32()) // login_token_number
		err := s.server.deleteCharacter(characterID, loginTokenString)
		if err == nil {
			s.logger.Info("Deleted character", zap.Int("CharacterID", characterID))
			s.cryptConn.SendPacket([]byte{0x01}) // DEL_SUCCESS
		}
	default:
		s.logger.Warn("Unknown request", zap.String("reqType", reqType))
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.LogInboundMessages {
			fmt.Printf("\n[Client] -> [Server]\nData [%d bytes]:\n%s\n", len(pkt), hex.Dump(pkt))
		}
	}
	return nil
}

func (s *Session) authenticate(username string, password string) {
	newCharaReq := false

	if username[len(username)-1] == 43 { // '+'
		username = username[:len(username)-1]
		newCharaReq = true
	}

	var id int
	var hash string
	bf := byteframe.NewByteFrame()

	err := s.server.db.QueryRow("SELECT id, password FROM users WHERE username = $1", username).Scan(&id, &hash)
	switch {
	case err == sql.ErrNoRows:
		s.logger.Info("User not found", zap.String("Username", username))
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.AutoCreateAccount {
			s.logger.Info("Creating user", zap.String("Username", username))
			err = s.server.registerDBAccount(username, password)
			if err == nil {
				bf.WriteBytes(s.makeSignResponse(id))
			}
		} else {
			bf.WriteUint8(uint8(SIGN_EAUTH))
		}
	case err != nil:
		bf.WriteUint8(uint8(SIGN_EABORT))
		s.logger.Error("Error getting user details", zap.Error(err))
	default:
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil || s.client == VITA || s.client == PS3 {
			s.logger.Debug("Passwords match!")
			if newCharaReq {
				err = s.server.newUserChara(username)
				if err != nil {
					s.logger.Error("Error adding new character to user", zap.Error(err))
					bf.WriteUint8(uint8(SIGN_EABORT))
					break
				}
			}
			// TODO: Need to auto delete user tokens after inactivity
			// exists, err := s.server.checkToken(id)
			// if err != nil {
			// 	s.logger.Info("Error checking for live tokens", zap.Error(err))
			// 	serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
			// 	break
			// }
			bf.WriteBytes(s.makeSignResponse(id))
		} else {
			s.logger.Warn("Incorrect password")
			bf.WriteUint8(uint8(SIGN_EPASS))
		}
	}

	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.LogOutboundMessages {
		fmt.Printf("\n[Server] -> [Client]\nData [%d bytes]:\n%s\n", len(bf.Data()), hex.Dump(bf.Data()))
	}

	err = s.cryptConn.SendPacket(bf.Data())
}

func (s *Session) handlePSSGN(bf *byteframe.ByteFrame) {
	_ = bf.ReadNullTerminatedBytes() // VITA = 0000000256, PS3 = 0000000255
	_ = bf.ReadBytes(2)              // VITA = 1, PS3 = !
	_ = bf.ReadBytes(82)
	psnUser := string(bf.ReadNullTerminatedBytes())
	var reqUsername string
	err := s.server.db.QueryRow(`SELECT username FROM users WHERE psn_id = $1`, psnUser).Scan(&reqUsername)
	if err == sql.ErrNoRows {
		resp := byteframe.NewByteFrame()
		resp.WriteUint8(uint8(SIGN_ECOGLINK))
		s.cryptConn.SendPacket(resp.Data())
		return
	}
	s.authenticate(reqUsername, "")
}

func (s *Session) handleDSGN(bf *byteframe.ByteFrame) {
	reqUsername := string(bf.ReadNullTerminatedBytes())
	reqPassword := string(bf.ReadNullTerminatedBytes())
	_ = string(bf.ReadNullTerminatedBytes()) // Unk
	s.authenticate(reqUsername, reqPassword)
}
