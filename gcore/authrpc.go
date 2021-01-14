package gcore

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/superp00t/gophercraft/crypto"
	"github.com/superp00t/gophercraft/crypto/srp"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/vsn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	authCheckSeed        = []byte{0xC5, 0xC6, 0x98, 0x95, 0x76, 0x3F, 0x1D, 0xCD, 0xB6, 0xA1, 0x37, 0x28, 0xB3, 0x12, 0xFF, 0x8A}
	sessionKeySeed       = []byte{0x58, 0xCB, 0xCF, 0x40, 0xFE, 0x2E, 0xCE, 0xA6, 0x5A, 0x90, 0xB8, 0x01, 0x68, 0x6C, 0x28, 0x0B}
	continuedSessionSeed = []byte{0x16, 0xAD, 0x0C, 0xD4, 0x46, 0xF9, 0x4F, 0xB2, 0xEF, 0x7D, 0xEA, 0x2A, 0x17, 0x66, 0x4D, 0x2F}
	encryptionKeySeed    = []byte{0xE9, 0x75, 0x3C, 0x50, 0x90, 0x93, 0x61, 0xDA, 0x3B, 0x07, 0xEE, 0xFA, 0xFF, 0x9D, 0x41, 0xB8}
)

func (c *Core) GetVersionData(ctx context.Context, _ *empty.Empty) (*sys.VersionData, error) {
	return &sys.VersionData{
		CoreVersion: Version,
	}, nil
}

func (c *Core) Ping(ctx context.Context, req *sys.PingMsg) (*sys.PingMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func (c *Core) CheckPeerIdentity(ctx context.Context, realmID uint64) (*sys.StatusMsg, error) {
	finger, err := sys.GetPeerFingerprint(ctx)
	if err != nil {
		return nil, err
	}

	rf, err := c.Auth.RealmsFile()
	if err != nil {
		return nil, err
	}

	rm, ok := rf.Realms[realmID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no realm exists corresponding to %d. You can solve this by adding your server to gcraft_auth/realms.yml", realmID)
	}

	if finger != rm.FP {
		return nil, status.Errorf(codes.PermissionDenied, "realm %d exists, but this realm's fingerprint does not match the one on record.", realmID)
	}

	return sys.Code(sys.Status_SysOK), nil
}

func (c *Core) AnnounceRealm(ctx context.Context, req *sys.AnnounceRealmMsg) (*sys.StatusMsg, error) {
	smsg, err := c.CheckPeerIdentity(ctx, req.RealmID)
	if err != nil {
		return smsg, err
	}

	var rlm Realm
	found, err := c.DB.Where("id = ?", req.RealmID).Get(&rlm)
	if err != nil {
		panic(err)
	}

	rlm.Name = req.RealmName
	rlm.Address = req.Address
	rlm.Version = vsn.Build(req.Build)
	rlm.Description = req.RealmDescription
	rlm.ActivePlayers = req.ActivePlayers
	rlm.Type = config.RealmType(req.Type)
	rlm.LastUpdated = time.Now()

	if !found {
		rlm.ID = req.RealmID
		c.DB.Insert(&rlm)
	} else {
		if _, err := c.DB.ID(req.RealmID).Cols("name", "address", "version", "description", "active_players", "type", "last_updated").Update(&rlm); err != nil {
			panic(err)
		}
	}

	return smsg, err
}

func (c *Core) VerifyWorld(ctx context.Context, req *sys.VerifyWorldQuery) (*sys.VerifyWorldResponse, error) {
	smsg, err := c.CheckPeerIdentity(ctx, req.RealmID)
	if err != nil {
		return &sys.VerifyWorldResponse{
			Status: smsg.Status,
		}, err
	}

	var sessionKey []byte

	var user Account
	found, _ := c.DB.Where("username = ?", req.Account).Get(&user)
	if !found {
		return &sys.VerifyWorldResponse{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	build := vsn.Build(req.Build)

	// Simplified check
	if vsn.Build(req.Build) == vsn.Alpha {
		hash, err := hex.DecodeString(string(req.Digest))
		if err != nil {
			return &sys.VerifyWorldResponse{
				Status: smsg.Status,
			}, err
		}

		if subtle.ConstantTimeCompare(hash, user.IdentityHash) == 0 {
			return &sys.VerifyWorldResponse{
				Status: smsg.Status,
			}, err
		}
	} else {
		var sk SessionKey
		found, err = c.DB.Where("id = ?", user.ID).Get(&sk)
		if !found {
			yo.Println("Database error: ", err)
			yo.Warn("connection authentication failure: no session key", req.Account, "from", req.IP)
			return &sys.VerifyWorldResponse{
				Status: sys.Status_SysUnauthorized,
			}, nil
		}

		// Use new calculations
		if build.AddedIn(vsn.NewAuthSystem) {
			buildInfo := build.BuildInfo()
			if buildInfo == nil {
				err := fmt.Errorf("build info for %s not found", buildInfo)
				return &sys.VerifyWorldResponse{
					Status: sys.Status_SysUnauthorized,
				}, err
			}
			if len(buildInfo.Win64AuthSeed) == 0 || len(buildInfo.Mac64AuthSeed) == 0 {
				err := fmt.Errorf("auth seed for %s not found", buildInfo)
				return &sys.VerifyWorldResponse{
					Status: sys.Status_SysUnauthorized,
				}, err
			}

			localChallenge := req.Seed
			serverChallenge := req.Salt
			digest := req.Digest

			sessionKeyHash := sha256.New()
			skl, _ := sessionKeyHash.Write(sk.K)
			if skl != 64 {
				panic("invalid key length")
			}

			yo.Spew(req.Digest)
			yo.Spew(localChallenge)
			yo.Spew(serverChallenge)
			yo.Spew(buildInfo.Win64AuthSeed)

			if user.Platform == "Wn64" {
				sessionKeyHash.Write(buildInfo.Win64AuthSeed)
			} else if user.Platform == "Mc64" {
				sessionKeyHash.Write(buildInfo.Mac64AuthSeed)
			} else {
				return &sys.VerifyWorldResponse{
					Status: sys.Status_SysUnauthorized,
				}, fmt.Errorf("invalid user platform %s", user.Platform)
			}

			digestKeyHash := sessionKeyHash.Sum(nil)

			hmc := hmac.New(sha256.New, digestKeyHash)
			hmc.Write(localChallenge)  //localChallenge
			hmc.Write(serverChallenge) //serverChallenge
			hmc.Write(authCheckSeed)
			authCheckHash := hmc.Sum(nil)

			if subtle.ConstantTimeCompare(authCheckHash[:24], digest[:24]) == 0 {
				err := errors.New(fmt.Sprintln("connection authentication failure: phony connection attempt to", req.Account, "from", req.IP))
				yo.Warn(err)
				return &sys.VerifyWorldResponse{
					Status: sys.Status_SysUnauthorized,
				}, err
			}

			keyDataDigest := sha256.Sum256(sk.K)

			sessionKeyHmac := hmac.New(sha256.New, keyDataDigest[:])
			sessionKeyHmac.Write(serverChallenge)
			sessionKeyHmac.Write(localChallenge)
			sessionKeyHmac.Write(sessionKeySeed)

			sessionKey = make([]byte, 40)
			skg := crypto.NewSessionKeyGenerator(sha256.New, sessionKeyHmac.Sum(nil))
			skg.Read(sessionKey)

			yo.Spew(sessionKey)

			encryptKeyGen := hmac.New(sha256.New, sessionKey)
			encryptKeyGen.Write(localChallenge)
			encryptKeyGen.Write(serverChallenge)
			encryptKeyGen.Write(encryptionKeySeed)

			encryptKeyHash := encryptKeyGen.Sum(nil)

			sessionKey = encryptKeyHash[:16]
		} else {
			digest := hash(
				[]byte(req.Account),
				[]byte{0, 0, 0, 0},
				req.Seed,
				req.Salt,
				sk.K,
			)

			if subtle.ConstantTimeCompare(digest, req.Digest) == 0 {
				err := errors.New(fmt.Sprintln("connection authentication failure: phony connection attempt to", req.Account, "from", req.IP))
				yo.Warn(err)
				return &sys.VerifyWorldResponse{
					Status: sys.Status_SysUnauthorized,
				}, err
			}

			sessionKey = sk.K
		}
	}

	var ga GameAccount
	found, _ = c.DB.Where("owner = ?", user.ID).Where("name = ?", req.GameAccount).Get(&ga)
	if !found {
		return &sys.VerifyWorldResponse{
			Status: sys.Status_SysUnauthorized,
		}, fmt.Errorf("no GameAccount detected")
	}

	return &sys.VerifyWorldResponse{
		Status:      sys.Status_SysOK,
		Tier:        user.Tier,
		SessionKey:  sessionKey,
		Account:     user.ID,
		GameAccount: ga.ID,
	}, nil
}

func (c *Core) ReportInfo(ctx context.Context, req *sys.Info) (*sys.StatusMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportInfo not implemented")
}

func fixedSleep(start time.Time, dur time.Duration) {
	now := time.Now()
	if now.Sub(start) > dur {
		return
	}

	time.Sleep(dur - now.Sub(start))
}

func (c *Core) CheckCredentials(ctx context.Context, req *sys.Credentials) (*sys.CredentialsResponse, error) {
	start := time.Now()
	timeout := 5 * time.Second

	// TODO: impose increased penalty for incorrect login

	var acc Account
	found, err := c.DB.Where("username = ?", req.Account).Get(&acc)
	if err != nil || !found {
		fixedSleep(start, timeout)
		if err != nil {
			yo.Warn(err)
		}
		return &sys.CredentialsResponse{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	identityHash := srp.HashCredentials(req.Account, req.Password)

	if subtle.ConstantTimeCompare(identityHash, acc.IdentityHash) == 0 {
		fixedSleep(start, timeout)
		return &sys.CredentialsResponse{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	return &sys.CredentialsResponse{
		Status: sys.Status_SysOK,
		Tier:   acc.Tier,
	}, nil
}

func (c *Core) GetNextRealmID(ctx context.Context, _ *empty.Empty) (*sys.AddRealmRequest, error) {
	var last uint64
	rfile, err := c.Auth.RealmsFile()
	if err != nil {
		panic(err)
	}
	for k := range rfile.Realms {
		if k >= last {
			last = k + 1
		}
	}
	return &sys.AddRealmRequest{
		RealmID: last,
	}, nil
}

func (c *Core) AddRealmToConfig(ctx context.Context, req *sys.AddRealmRequest) (*sys.StatusMsg, error) {
	start := time.Now()
	timeout := 5 * time.Second

	// TODO: impose increased penalty for incorrect login

	var acc Account
	found, err := c.DB.Where("username = ?", req.Credentials.Account).Get(&acc)
	if err != nil || !found {
		fixedSleep(start, timeout)
		if err != nil {
			yo.Warn(err)
		}
		return &sys.StatusMsg{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	identityHash := srp.HashCredentials(req.Credentials.Account, req.Credentials.Password)

	if subtle.ConstantTimeCompare(identityHash, acc.IdentityHash) == 0 {
		fixedSleep(start, timeout)
		return &sys.StatusMsg{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	rfile, err := c.Auth.RealmsFile()
	if err != nil {
		panic(err)
	}

	_, realmExists := rfile.Realms[req.RealmID]
	if realmExists {
		return nil, fmt.Errorf("gcore: Realm ID %d already exists. To make changes to this realm, you must edit Auth/Realms.txt locally.")
	}

	c.Auth.SetRealm(req.RealmID, config.Realm{
		FP: req.GetRealmFingerprint(),
	})

	return &sys.StatusMsg{
		Status: sys.Status_SysOK,
	}, nil
}

func hash(input ...[]byte) []byte {
	bt := sha1.Sum(bytes.Join(input, nil))
	return bt[:]
}
