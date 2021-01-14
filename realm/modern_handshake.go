package realm

import (
	"bufio"
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	gcrypto "github.com/superp00t/gophercraft/crypto"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/i18n"
	"github.com/superp00t/gophercraft/packet"
)

var (
	clientHelloV2 = []byte("WORLD OF WARCRAFT CONNECTION - CLIENT TO SERVER - V2")
	serverHelloV2 = []byte("WORLD OF WARCRAFT CONNECTION - SERVER TO CLIENT - V2\n")
)

func (rs *Server) handleModern(c net.Conn) {
	reader := bufio.NewReader(c)

	// Send server hello
	if _, err := c.Write(serverHelloV2); err != nil {
		yo.Warn(err)
		c.Close()
		return
	}

	// Get client hello
	cHello := make([]byte, len(clientHelloV2))
	_, err := reader.Read(cHello[:])
	if err != nil {
		yo.Warn(err)
		c.Close()
		return
	}

	if !bytes.Equal(cHello, clientHelloV2) {
		yo.Warn("Connection sent invalid client hello")
		yo.Spew(cHello)
		c.Close()
		return
	}

	newline, err := reader.ReadByte()
	if err != nil {
		yo.Warn(err)
		c.Close()
		return
	}

	if newline != '\n' {
		yo.Warn("Invalid terminator", newline)
		c.Close()
		return
	}

	// Send crypto data to client.
	smsgAuthChallenge, err := packet.ConvertWorldTypeToUint(rs.Build(), packet.SMSG_AUTH_CHALLENGE)
	if err != nil {
		panic(err)
	}

	ac := etc.NewBuffer()
	ac.WriteInt32(2 + 16 + 32 + 1)
	ac.Write(make([]byte, 12))
	ac.WriteUint16(uint16(smsgAuthChallenge))
	serverChallenge := make([]byte, 16)
	dosChallenge := make([]byte, 32)
	io.ReadFull(rand.Reader, serverChallenge)
	io.ReadFull(rand.Reader, dosChallenge)
	ac.Write(dosChallenge)
	ac.Write(serverChallenge)
	ac.WriteByte(1)
	authChallengePacket := ac.Bytes()

	if _, err := c.Write(authChallengePacket[:]); err != nil {
		yo.Warn(err)
		c.Close()
		return
	}

	// Read client response (CMSG_AUTH_SESSION)
	// Provides additional cryptography data.
	var asHeader [16]byte
	if _, err := reader.Read(asHeader[:]); err != nil {
		yo.Warn(err)
		c.Close()
		return
	}
	asSize := int32(binary.LittleEndian.Uint32(asHeader[:4]))
	authSessionData := make([]byte, asSize)
	if _, err := reader.Read(authSessionData[:]); err != nil {
		yo.Warn(err)
		c.Close()
		return
	}
	yo.Spew(authSessionData)
	authSession := etc.FromBytes(authSessionData)
	opcode := authSession.ReadUint16()
	wt, err := packet.LookupWorldType(rs.Build(), uint32(opcode))
	if err != nil {
		panic(err)
	}
	if wt != packet.CMSG_AUTH_SESSION {
		yo.Warn(wt)
		return
	}

	connectToRsa := gcrypto.GetConnectionRSAKey()

	dosResponse := authSession.ReadUint64()
	regionID := authSession.ReadUint32()
	battlegroupID := authSession.ReadUint32()
	realmID := authSession.ReadInt32()
	localChallenge := authSession.ReadBytes(16)
	digest := authSession.ReadBytes(24)
	useIpv6 := authSession.ReadByte()

	fmt.Println(dosResponse, regionID, battlegroupID, realmID, localChallenge, digest[0], serverChallenge[0], dosChallenge[0], useIpv6)

	realmJoinTicketSize := int(authSession.ReadUint32())
	realmJoinTicket := string(authSession.ReadBytes(realmJoinTicketSize))

	ticket := strings.SplitN(realmJoinTicket, ":", 2)
	user := ticket[0]
	gameAccount := ticket[1]

	// Invoke the gcore server.
	// gcore will do the computations for us and return a usable key.
	resp, err := rs.AuthServiceClient.VerifyWorld(context.Background(), &sys.VerifyWorldQuery{
		RealmID:     rs.RealmID(),
		Build:       uint32(rs.Build()),
		Account:     user,
		GameAccount: gameAccount,
		IP:          c.RemoteAddr().String(),
		Digest:      digest,
		Salt:        serverChallenge,
		Seed:        localChallenge,
	})

	if err != nil {
		yo.Warn(err)
		c.Close()
		return
	}

	encryptKey := resp.SessionKey

	if len(encryptKey) != 16 {
		panic(len(encryptKey))
	}

	enableEncryptionSeed := [16]byte{0x90, 0x9C, 0xD0, 0x50, 0x5A, 0x2C, 0x14, 0xDD, 0x5C, 0x2C, 0xC0, 0x64, 0x14, 0xF3, 0xFE, 0xC9}

	hmc := hmac.New(sha256.New, encryptKey)
	hmc.Write([]byte{1}) // enabled
	hmc.Write(enableEncryptionSeed[:])
	enableEncryptionResult := hmc.Sum(nil)

	signature, err := rsa.SignPKCS1v15(nil, connectToRsa, crypto.SHA256, enableEncryptionResult)
	if err != nil {
		panic(err)
	}
	yo.Spew(signature)

	if len(signature) != 256 {
		panic(len(signature))
	}

	if err := rsa.VerifyPKCS1v15(connectToRsa.Public().(*rsa.PublicKey), crypto.SHA256, enableEncryptionResult, signature); err != nil {
		panic(err)
	}

	gcrypto.ReverseBytes(signature)

	enableEncryption, err := packet.ConvertWorldTypeToUint(rs.Build(), packet.SMSG_ENTER_ENCRYPTED_MODE)
	if err != nil {
		panic(err)
	}
	header := make([]byte, 16)

	body := etc.NewBuffer()
	body.WriteUint16(uint16(enableEncryption))
	body.Write(signature)
	body.WriteByte(1 << 7)

	binary.LittleEndian.PutUint32(header[0:4], uint32(int32(body.Len())))

	fullPacket := append(header, body.Bytes()...)

	_, err = c.Write(fullPacket)
	if err != nil {
		yo.Warn(err)
		c.Close()
		return
	}

	for {

		var eemHeader [16]byte
		if _, err := reader.Read(eemHeader[:]); err != nil {
			yo.Warn(err)
			c.Close()
			return
		}

		size := int32(binary.LittleEndian.Uint32(eemHeader[0:4]))
		if size > 50 || size < 2 {
			yo.Warn(size)
			c.Close()
			return
		}
		eemAckPacket := make([]byte, size)
		if _, err := reader.Read(eemAckPacket); err != nil {
			yo.Warn(err)
			c.Close()
			return
		}
		wt, err = packet.LookupWorldType(rs.Build(), uint32(binary.LittleEndian.Uint16(eemAckPacket[0:2])))
		if err != nil {
			yo.Warn(err)
			c.Close()
			return
		}

		yo.Ok(wt)

		switch wt {
		case packet.CMSG_LOG_DISCONNECT:
			var dc error = packet.DCReason(binary.LittleEndian.Uint32(eemAckPacket[2:6]))
			yo.Ok("Client logging disconnect reason", dc)
			// yo.Spew(eemAckPacket)
			// c.Close()
			// return
		case packet.CMSG_ENABLE_NAGLE:
			yo.Ok("Client requested to enable Nagle's algorithm (delay)")
			c.(*net.TCPConn).SetNoDelay(false)
		case packet.CMSG_ENTER_ENCRYPTED_MODE_ACK:
			break
		default:
			panic(wt)
		}
	}

	reader = nil

	conn := packet.NewConnection(rs.Build(), c, encryptKey, true)

	sess := &Session{
		WS:          rs,
		Account:     resp.Account,
		GameAccount: resp.GameAccount,
		Connection:  conn,
		Tier:        resp.Tier,
		Locale:      i18n.Locale(resp.Locale),
	}
	sess.Init()
}
