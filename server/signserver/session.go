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

// Session holds state for the sign server connection.
type Session struct {
	sync.Mutex
	logger    *zap.Logger
	server    *Server
	rawConn   net.Conn
	cryptConn *network.CryptConn
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
	case "DLTSKEYSIGN:100":
		fallthrough
	case "DSGN:100":
		err := s.handleDSGNRequest(bf)
		if err != nil {
			return nil
		}
	case "DELETE:100":
		loginTokenString := string(bf.ReadNullTerminatedBytes())
		characterID := int(bf.ReadUint32())
		_ = int(bf.ReadUint32()) // login_token_number
		s.server.deleteCharacter(characterID, loginTokenString)
		s.logger.Info("Deleted character", zap.Int("CharacterID", characterID))
		err := s.cryptConn.SendPacket([]byte{0x01}) // DEL_SUCCESS
		if err != nil {
			return nil
		}
	default:
		s.logger.Warn("Unknown sign request", zap.String("reqType", reqType))
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.LogInboundMessages {
			fmt.Printf("\n[Client] -> [Server]\nData [%d bytes]:\n%s\n", len(pkt), hex.Dump(pkt))
		}
	}

	return nil
}

func (s *Session) handleDSGNRequest(bf *byteframe.ByteFrame) error {

	reqUsername := string(bf.ReadNullTerminatedBytes())
	reqPassword := string(bf.ReadNullTerminatedBytes())
	_ = string(bf.ReadNullTerminatedBytes()) // Unk

	newCharaReq := false

	if reqUsername[len(reqUsername)-1] == 43 { // '+'
		reqUsername = reqUsername[:len(reqUsername)-1]
		newCharaReq = true
	}

	var (
		id       int
		password string
	)
	err := s.server.db.QueryRow("SELECT id, password FROM users WHERE username = $1", reqUsername).Scan(&id, &password)
	var serverRespBytes []byte
	switch {
	case err == sql.ErrNoRows:
		s.logger.Info("User not found", zap.String("Username", reqUsername))
		serverRespBytes = makeSignInFailureResp(SIGN_EAUTH)

		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.AutoCreateAccount {
			s.logger.Info("Creating user", zap.String("Username", reqUsername))
			err = s.server.registerDBAccount(reqUsername, reqPassword)
			if err != nil {
				s.logger.Error("Error registering new user", zap.Error(err))
				serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
				break
			}
		} else {
			break
		}

		var id int
		err = s.server.db.QueryRow("SELECT id FROM users WHERE username = $1", reqUsername).Scan(&id)
		if err != nil {
			s.logger.Error("Error getting new user ID", zap.Error(err))
			serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
			break
		}

		serverRespBytes = s.makeSignInResp(id)
		break
	case err != nil:
		serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
		s.logger.Error("Error getting user details", zap.Error(err))
		break
	default:
		if bcrypt.CompareHashAndPassword([]byte(password), []byte(reqPassword)) == nil {
			s.logger.Debug("Passwords match!")
			if newCharaReq {
				err = s.server.newUserChara(reqUsername)
				if err != nil {
					s.logger.Error("Error adding new character to user", zap.Error(err))
					serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
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
			serverRespBytes = s.makeSignInResp(id)
		} else {
			s.logger.Warn("Incorrect password")
			serverRespBytes = makeSignInFailureResp(SIGN_EPASS)
		}

	}

	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.LogOutboundMessages {
		fmt.Printf("\n[Server] -> [Client]\nData [%d bytes]:\n%s\n", len(serverRespBytes), hex.Dump(serverRespBytes))
	}

	err = s.cryptConn.SendPacket(serverRespBytes)
	if err != nil {
		return err
	}

	return nil
}
