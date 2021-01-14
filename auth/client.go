package auth

import (
	"fmt"
	"net"
	"strings"

	"github.com/superp00t/gophercraft/crypto/srp"
	"github.com/superp00t/gophercraft/vsn"
)

type Client struct {
	Version    vsn.Build
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

	if _, err := cl.conn.Write(LogonChallengePacket_C(cl.Version, user)); err != nil {
		return err
	}

	buf := make([]byte, 512)
	_, err = cl.conn.Read(buf)
	if err != nil {
		return err
	}

	auth, err := UnmarshalAuthLogonChallenge_S(buf)
	if err != nil {
		return err
	}

	if auth.Error != WOW_SUCCESS {
		return fmt.Errorf("auth: server returned %s", auth.Error)
	}

	if len(auth.N) != 32 {
		return fmt.Errorf("auth: server sent invalid prime number length %d", len(auth.N))
	}

	_, K, A, M1 := srp.SRPCalculate(user, pass, auth.B, auth.N, auth.S)
	cl.SessionKey = K
	proof := &AuthLogonProof_C{
		A:            A,
		M1:           M1,
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

	alps, err := UnmarshalAuthLogonProof_S(cl.Version, buf)
	if err != nil {
		return err
	}

	if alps.Error != WOW_SUCCESS {
		return fmt.Errorf("Server returned %s", alps.Error)
	}

	return nil
}

func (cl *Client) GetRealmlist() (*RealmList_S, error) {
	if _, err := cl.conn.Write(RealmList_C); err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	if _, err := cl.conn.Read(buf); err != nil {
		return nil, err
	}
	rls, err := UnmarshalRealmList_S(cl.Version, buf)
	if err != nil {
		return nil, err
	}
	return rls, nil
}

func Login(version vsn.Build, address, username, password string) (*Client, error) {
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
