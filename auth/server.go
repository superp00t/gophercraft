package auth

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net"
	"time"

	"github.com/superp00t/etc/yo"

	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/srp"
)

type Backend interface {
	GetAccount(user string) *Account
	ListRealms(user string, buildnumber uint32) []packet.RealmListing
	StoreKey(user string, K []byte)
	HandleSpecialConn(protocol string, conn net.Conn)
}

type Account struct {
	Username     string
	IdentityHash []byte
}

type Server struct {
	Listen string
	Name   string

	Backend
}

// func (server *Server) handleTCP(c net.Conn) {
// 	rdr := bufio.NewReader(c)

// 	dat := string(data)

// 	psk := strings.TrimRight(dat, "\r\n")

// 	if psk != server.APIKey() {
// 		c.Write([]byte("Invalid API key.\n"))
// 		c.Close()
// 		return
// 	}

// 	fmt.Fprintln(c, "Gophercraft Admin Shell")

// 	for {
// 		str, err := rdr.ReadString('\n')
// 		if err != nil {
// 			yo.Warn(err)
// 			return
// 		}

// 		str = strings.TrimRight(str, "\r\n")

// 		if _, err = c.Write([]byte{'>'}); err != nil {
// 			yo.Warn(err)
// 			return
// 		}

// 		fmt.Println("recv command", str)
// 	}
// }

func (server *Server) Handle(cn net.Conn) {
	var (
		acc *Account
		g   *srp.BigNum
		s   *srp.BigNum
		N   *srp.BigNum
		v   *srp.BigNum
		b   *srp.BigNum
		B   *srp.BigNum
		alc *packet.AuthLogonChallenge_C
	)

	c := wrapConn(cn)

	yo.Ok("New authserver connection from", c.RemoteAddr())

	const (
		stateUnauthorized = iota
		stateChallenging
		stateAuthorized
	)

	state := stateUnauthorized

	// every iteration of this loop reads an opcode from the TCP socket, and associated data.
	for {
		if state == stateUnauthorized {
			data, err := c.Peek(1)
			if err != nil {
				yo.Warn(err)
				return
			}

			if packet.AuthType(data[0]) != packet.AUTH_LOGON_CHALLENGE {
				// SSH server for remote administration
				server.Backend.HandleSpecialConn("grpc", c)
				return
			}
		}

		buf := make([]byte, 2048)
		rd, err := c.Read(buf)
		if err != nil {
			c.Close()
			return
		}

		at := packet.AuthType(buf[0])

		yo.Ok("recv", at)
		switch at {
		case packet.AUTH_LOGON_CHALLENGE:
			alc, err = packet.UnmarshalAuthLogonChallenge_C(buf[:rd])
			if err != nil {
				c.Close()
				return
			}

			build := alc.Build
			validBuilds := []uint16{
				5875,
				12340,
			}

			invalid := true

			for _, v := range validBuilds {
				if v == build {
					invalid = false
					break
				}
			}

			// TODO: apply ratelimiting to prevent brute force attacks.
			if invalid {
				yo.Warn("User attempted to log in with invalid client", alc.Build)
				c.Write([]byte{
					uint8(packet.AUTH_LOGON_CHALLENGE),
					0x00,
					uint8(packet.WOW_FAIL_VERSION_INVALID),
				})
				time.Sleep(1 * time.Second)
				c.Close()
				return
			}

			acc = server.GetAccount(string(alc.I))
			if acc == nil {
				// User could not be found.
				c.Write([]byte{
					uint8(packet.AUTH_LOGON_CHALLENGE),
					0x00,
					uint8(packet.WOW_FAIL_UNKNOWN_ACCOUNT),
				})
				continue
			}

			state = stateChallenging

			// Perform SRP check
			nh := "894B645E89E1535BBDAD5B8B290650530801B18EBFBF5E8FAB3C82872A3E9BB7"
			nb, _ := hex.DecodeString(nh)
			N = &srp.BigNum{X: new(big.Int).SetBytes(nb)}
			v, s, _ = srp.ServerCalcVSX(acc.IdentityHash, N)

			g = srp.BigNumFromInt(7)
			b = srp.BigNumFromRand(19)
			gmod := g.ModExp(b, N)
			B = ((v.Multiply(srp.BigNumFromInt(3))).Add(gmod)).Mod(N)
			pkt := &packet.AuthLogonChallenge_S{
				Cmd:   packet.AUTH_LOGON_CHALLENGE,
				Error: packet.WOW_SUCCESS,
				B:     B.ToArray(),
				G:     7,
				N:     N.ToArray(),
				S:     s.ToArray(),
				Unk3:  srp.BigNumFromRand(16).ToArray(),
			}

			c.Write(pkt.Encode())
		// Client posts cryptographic material to the server to complete SRP.
		case packet.AUTH_LOGON_PROOF:
			if state != stateChallenging {
				break
			}

			alpc, err := packet.UnmarshalAuthLogonProof_C(buf)
			if err != nil {
				c.Close()
				return
			}

			K, valid, M3 := srp.ServerLogonProof(acc.Username,
				srp.BigNumFromArray(alpc.A),
				srp.BigNumFromArray(alpc.M1),
				b,
				B,
				s,
				N,
				v)

			if !valid {
				yo.Println(acc.Username, "Invalid login")
				var resp = []byte{
					uint8(packet.AUTH_LOGON_PROOF),
					uint8(packet.WOW_FAIL_UNKNOWN_ACCOUNT),
				}

				if alc.Build == 12340 {
					resp = append(resp, 0, 0)
				}

				c.Write(resp)
				// time.Sleep(1 * time.Second)
				// c.Close()
				continue
			}

			server.StoreKey(acc.Username, K)

			yo.Println(acc.Username, "successfully authenticated")

			proof := &packet.AuthLogonProof_S{
				Cmd:          packet.AUTH_LOGON_PROOF,
				Error:        packet.WOW_SUCCESS,
				M2:           M3,
				AccountFlags: 0x00800000,
				SurveyID:     0,
				Unk3:         0,
			}

			_, err = c.Write(proof.Encode(uint32(alc.Build)))
			if err != nil {
				c.Close()
				return
			}

			state = stateAuthorized
		// Client requested a list of realms from the server.
		case packet.REALM_LIST:
			if state != stateAuthorized {
				break
			}

			rls := server.ListRealms(acc.Username, uint32(alc.Build))
			rlst := &packet.RealmList_S{
				Cmd:    packet.REALM_LIST,
				Realms: rls,
			}
			data := rlst.Encode(uint32(alc.Build))
			_, err := c.Write(data)
			if err != nil {
				break
			}
		}
	}
}

func (server *Server) Run() error {
	l, err := net.Listen("tcp", server.Listen)
	if err != nil {
		return err
	}

	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go server.Handle(c)
	}
}

func rnd(l int) []byte {
	b := make([]byte, l)
	rand.Read(b)
	return b
}

func (h *Server) Close() error {
	panic("cannot close")
	return nil
}
