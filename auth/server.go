package auth

import (
	"crypto/rand"
	"net"
	"time"

	"github.com/superp00t/etc/yo"

	"github.com/superp00t/gophercraft/crypto/srp"
	"github.com/superp00t/gophercraft/gcore"
	"github.com/superp00t/gophercraft/vsn"
)

var (
	VersionChallenge = [...]byte{0xBA, 0xA3, 0x1E, 0x99, 0xA0, 0x0B, 0x21, 0x57, 0xFC, 0x37, 0x3F, 0xB3, 0x69, 0xCD, 0xD2, 0xF1}
)

type Backend interface {
	GetAccount(user string) (*gcore.Account, []gcore.GameAccount, error)
	ListRealms() []gcore.Realm
	StoreKey(user, locale, platform string, K []byte)
	HandleSpecialConn(protocol string, conn net.Conn)
}

type Server struct {
	Listen string
	Name   string

	Backend
}

func (server *Server) Handle(cn net.Conn) {
	var (
		acc    *gcore.Account
		g      *srp.BigNum
		salt   *srp.BigNum
		N      *srp.BigNum
		v      *srp.BigNum
		b      *srp.BigNum
		B      *srp.BigNum
		alc    *AuthLogonChallenge_C
		locale string
		// gameAccounts []gcore.GameAccount
		build        vsn.Build
		platformOS   string = "Wn"
		platformArch string = "32"
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
			// Possibly we're dealing with a TLS connection, in which case pass to the GRPC server.
			data, err := c.Peek(1)
			if err != nil {
				yo.Warn(err)
				return
			}

			if AuthType(data[0]) != AUTH_LOGON_CHALLENGE {
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

		at := AuthType(buf[0])

		switch at {
		case AUTH_LOGON_CHALLENGE:
			alc, err = UnmarshalAuthLogonChallenge_C(buf[:rd])
			if err != nil {
				c.Close()
				return
			}

			if alc.Platform == "x86" {
				platformArch = "32"
			}
			if alc.Platform == "x64" {
				platformArch = "64"
			} else {
				platformArch = "??"
			}

			if alc.OS == "Win" {
				platformOS = "Wn"
			}
			if alc.OS == "OSX" {
				platformOS = "Mc"
			}

			locale = alc.Country

			build = vsn.Build(alc.Build)

			// These builds have been tested and confirmed to function as intended.
			// Add more if you find out they work as well.
			validBuilds := []vsn.Build{
				vsn.V1_12_1,
				vsn.V2_4_3,
				vsn.V3_3_5a,
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
					uint8(AUTH_LOGON_CHALLENGE),
					0x00,
					uint8(WOW_FAIL_VERSION_INVALID),
				})
				time.Sleep(1 * time.Second)
				c.Close()
				return
			}

			acc, _, err = server.GetAccount(string(alc.I))
			if err != nil {
				// User could not be found.
				c.Write([]byte{
					uint8(AUTH_LOGON_CHALLENGE),
					0x00,
					uint8(WOW_FAIL_UNKNOWN_ACCOUNT),
				})
				continue
			}

			state = stateChallenging

			// Generate parameters
			salt = srp.BigNumFromRand(32)
			g = srp.Generator.Copy()
			N = srp.Prime.Copy()

			// Compute ephemeral (temporary) variables
			_, v = srp.CalculateVerifier(acc.IdentityHash, g, N, salt)
			b, B = srp.ServerGenerateEphemeralValues(g, N, v)

			pkt := &AuthLogonChallenge_S{
				Error:            WOW_SUCCESS,
				B:                B.ToArray(32),
				G:                g.ToArray(1),
				N:                N.ToArray(32),
				S:                salt.ToArray(32),
				VersionChallenge: VersionChallenge[:],
			}

			yo.Spew(pkt)

			data := pkt.Encode(build)

			i, err := c.Write(data)
			if err != nil || i != len(data) {
				yo.Warn("COULD NOT TRANSFER ALL", i, len(data), err)
			}
		// Client has posted cryptographic material to the server to complete SRP.
		case AUTH_LOGON_PROOF:
			if state != stateChallenging {
				break
			}

			alpc, err := UnmarshalAuthLogonProof_C(buf[:rd])
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
					uint8(AUTH_LOGON_PROOF),
					uint8(WOW_FAIL_UNKNOWN_ACCOUNT),
				}

				if build.AddedIn(vsn.V2_4_3) {
					resp = append(resp, 0, 0)
				}

				c.Write(resp)
				continue
			}

			server.StoreKey(acc.Username, locale, platformOS+platformArch, K)

			proof := &AuthLogonProof_S{
				Cmd:          AUTH_LOGON_PROOF,
				Error:        WOW_SUCCESS,
				M2:           M2,
				AccountFlags: 0x00800000,
				SurveyID:     0,
				Unk3:         0,
			}

			_, err = c.Write(proof.Encode(build))
			if err != nil {
				c.Close()
				return
			}

			state = stateAuthorized
		// Client requested a list of realms from the server.
		case REALM_LIST:
			if state != stateAuthorized {
				break
			}

			realms := []gcore.Realm{}

			realmList := server.ListRealms()
			for _, v := range realmList {
				if v.Version == build {
					realms = append(realms, v)
				}
			}

			realmListS := MakeRealmlist(realms)

			data := realmListS.Encode(vsn.Build(alc.Build))
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
