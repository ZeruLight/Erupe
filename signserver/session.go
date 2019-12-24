package signserver

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"sync"

	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// Session holds state for the sign server connection.
type Session struct {
	sync.Mutex
	sid       int
	server    *Server
	rawConn   *net.Conn
	cryptConn *network.CryptConn
}

func (session *Session) fail() {
	session.server.Lock()
	delete(session.server.sessions, session.sid)
	session.server.Unlock()

}

func (session *Session) work() {
	for {
		pkt, err := session.cryptConn.ReadPacket()
		if err != nil {
			session.fail()
			return
		}

		err = session.handlePacket(pkt)
		if err != nil {
			session.fail()
			return
		}
	}
}

func (session *Session) handlePacket(pkt []byte) error {
	bf := byteframe.NewByteFrameFromBytes(pkt)
	reqType := string(bf.ReadNullTerminatedBytes())
	switch reqType {
	case "DSGN:100":
		session.handleDSGNRequest(bf)
		break
	case "DELETE:100":
		loginTokenString := string(bf.ReadNullTerminatedBytes())
		_ = loginTokenString
		characterID := bf.ReadUint32()

		fmt.Printf("Got delete request for character ID: %v\n", characterID)
		fmt.Printf("remaining unknown data:\n%s\n", hex.Dump(bf.DataFromCurrent()))
	default:
		fmt.Printf("Got unknown request type %s, data:\n%s\n", reqType, hex.Dump(bf.DataFromCurrent()))
	}

	return nil
}

func (session *Session) handleDSGNRequest(bf *byteframe.ByteFrame) error {
	reqUsername := string(bf.ReadNullTerminatedBytes())
	reqPassword := string(bf.ReadNullTerminatedBytes())
	reqUnk := string(bf.ReadNullTerminatedBytes())
	fmt.Printf("Got sign in request:\n\tUsername: %s\n\tPassword %s\n\tUnk: %s\n", reqUsername, reqPassword, reqUnk)

	// TODO(Andoryuuta): remove plaintext password storage if this ever becomes more than a toy project.
	var (
		id       int
		password string
	)
	err := session.server.db.QueryRow("SELECT id, password FROM users WHERE username = $1", reqUsername).Scan(&id, &password)
	var serverRespBytes []byte
	switch {
	case err == sql.ErrNoRows:
		fmt.Printf("No rows for username %s\n", reqUsername)
		serverRespBytes = makeSignInFailureResp(SIGN_EAUTH)
		break
	case err != nil:
		serverRespBytes = makeSignInFailureResp(SIGN_EABORT)
		fmt.Println("Got error on SQL query!")
		fmt.Println(err)
		break
	default:
		if reqPassword == password {
			fmt.Println("Passwords match!")
			serverRespBytes = makeSignInResp(reqUsername)
		} else {
			fmt.Println("Passwords don't match!")
			serverRespBytes = makeSignInFailureResp(SIGN_EPASS)
		}

	}

	err = session.cryptConn.SendPacket(serverRespBytes)
	if err != nil {
		return err
	}

	return nil
}
