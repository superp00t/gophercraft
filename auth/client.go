package auth

import (
	"fmt"
	"net"
	"strings"

	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/srp"
)

type Client struct {
	Version    uint32
	Address    string
	Username   string
	Password   string
	SessionKey []byte
	conn       net.Conn
}

// connect using the vanilla auth protocol.
func (cl *Client) connectClassic() error {
	var err error
	cl.conn, err = net.Dial("tcp", cl.Address)
	if err != nil {
		return err
	}

	user := strings.ToUpper(cl.Username)
	pass := strings.ToUpper(cl.Password)

	if _, err := cl.conn.Write(packet.LogonChallengePacket_C(cl.Version, user)); err != nil {
		return err
	}

	buf := make([]byte, 512)
	_, err = cl.conn.Read(buf)
	if err != nil {
		return err
	}

	auth, err := packet.UnmarshalAuthLogonChallenge_S(buf)
	if err != nil {
		return err
	}

	if auth.Error != packet.WOW_SUCCESS {
		return fmt.Errorf("auth: server returned %s", auth.Error)
	}

	_, K, A, M1 := srp.SRPCalculate(user, pass, auth.B, auth.N, auth.S)
	cl.SessionKey = K
	proof := &packet.AuthLogonProof_C{
		Cmd:          packet.AUTH_LOGON_PROOF,
		A:            A,
		M1:           M1,
		CRC:          make([]byte, 20),
		NumberOfKeys: 0,
		SecFlags:     0,
	}

	if _, err = cl.conn.Write(proof.Encode()); err != nil {
		return err
	}

	buf = make([]byte, 512)
	_, err = cl.conn.Read(buf)
	if err != nil {
		return err
	}

	alps, err := packet.UnmarshalAuthLogonProof_S(cl.Version, buf)
	if err != nil {
		return err
	}

	if alps.Error != packet.WOW_SUCCESS {
		return fmt.Errorf("Server returned %s", alps.Error)
	}

	return nil
}

func (cl *Client) GetRealmlist() (*packet.RealmList_S, error) {
	if _, err := cl.conn.Write(packet.RealmList_C); err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	if _, err := cl.conn.Read(buf); err != nil {
		return nil, err
	}
	rls, err := packet.UnmarshalRealmList_S(cl.Version, buf)
	if err != nil {
		return nil, err
	}
	return rls, nil
}

func Login(version uint32, address, username, password string) (*Client, error) {
	cl := &Client{
		Version:  version,
		Address:  address,
		Username: username,
		Password: password,
	}

	switch version {
	case 5875, 12340:
		err := cl.connectClassic()
		if err != nil {
			return nil, err
		}

		return cl, nil
	// todo: add bnet client support
	default:
		return nil, fmt.Errorf("auth: unsupported protocol version %d", version)
	}
}
