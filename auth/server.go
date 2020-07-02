package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/superp00t/etc/yo"

	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/srp"
	"github.com/superp00t/gophercraft/vsn"
)

type Backend interface {
	GetAccount(user string) *Account
	ListRealms(user string, buildnumber vsn.Build) []packet.RealmListing
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
		acc  *Account
		g    *srp.BigNum
		salt *srp.BigNum
		N    *srp.BigNum
		v    *srp.BigNum
		b    *srp.BigNum
		B    *srp.BigNum
		alc  *packet.AuthLogonChallenge_C
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

			fmt.Println("Challenge buffer")
			yo.Spew(buf[:rd])
			yo.Spew(alc)

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

			// Generate parameters
			salt = srp.BigNumFromRand(32)
			g = srp.Generator.Copy()
			N = srp.Prime.Copy()

			// Compute temporary variables
			_, v = srp.CalculateVerifier(acc.IdentityHash, g, N, salt)
			b, B = srp.ServerGenerateEphemeralValues(g, N, v)

			pkt := &packet.AuthLogonChallenge_S{
				Cmd:              packet.AUTH_LOGON_CHALLENGE,
				Error:            packet.WOW_SUCCESS,
				B:                B.ToArray(32),
				G:                g.ToArray(1),
				N:                N.ToArray(32),
				S:                salt.ToArray(32),
				VersionChallenge: srp.BigNumFromRand(16).ToArray(16),
			}

			c.Write(pkt.Encode())
		// Client has posted cryptographic material to the server to complete SRP.
		case packet.AUTH_LOGON_PROOF:
			if state != stateChallenging {
				break
			}

			alpc, err := packet.UnmarshalAuthLogonProof_C(buf)
			if err != nil {
				c.Close()
				return
			}

			yo.Spew(alpc)

			K, valid, M2 := srp.ServerLogonProof(acc.Username,
				srp.BigNumFromArray(alpc.A),
				srp.BigNumFromArray(alpc.M1),
				b,
				B,
				salt,
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
				continue
			}

			server.StoreKey(acc.Username, K)

			yo.Println(acc.Username, "successfully authenticated")
			yo.Println("Client A", hex.EncodeToString(alpc.A))
			yo.Println("Client m1", hex.EncodeToString(alpc.M1))

			proof := &packet.AuthLogonProof_S{
				Cmd:          packet.AUTH_LOGON_PROOF,
				Error:        packet.WOW_SUCCESS,
				M2:           M2,
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

			rls := server.ListRealms(acc.Username, vsn.Build(alc.Build))
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
