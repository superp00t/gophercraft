package gcore

import (
	"bytes"
	"context"
	"crypto/sha1"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/sys"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	rlm.Version = req.Build
	rlm.Description = req.RealmDescription
	rlm.ActivePlayers = req.ActivePlayers
	rlm.Type = req.Type
	rlm.LastUpdated = time.Now()

	if !found {
		rlm.ID = req.RealmID
		c.DB.Insert(&rlm)
	} else {
		c.DB.AllCols().Update(&rlm)
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

	var user Account
	found, _ := c.DB.Where("username = ?", req.Account).Get(&user)
	if !found {
		return &sys.VerifyWorldResponse{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	var sk SessionKey
	found, err := c.DB.Where("id = ?", user.ID).Get(&sk)
	if !found {
		yo.Println("Database error: ", err)
		yo.Warn("connection authentication failure: no session key", req.Account, "from", req.IP)
		return &sys.VerifyWorldResponse{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	digest := hash(
		[]byte(req.Account),
		[]byte{0, 0, 0, 0},
		req.Seed,
		req.Salt,
		sk.K,
	)

	if !bytes.Equal(digest, req.Digest) {
		yo.Warn("connection authentication failure: phony connection attempt to", req.Account, "from", req.IP)
		return &sys.VerifyWorldResponse{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	var ga GameAccount
	found, _ = c.DB.Where("owner = ?", user.ID).Where("name = ?", req.GameAccount).Get(&ga)
	if !found {
		return &sys.VerifyWorldResponse{
			Status: sys.Status_SysUnauthorized,
		}, nil
	}

	return &sys.VerifyWorldResponse{
		Status:      sys.Status_SysOK,
		SessionKey:  sk.K,
		GameAccount: ga.ID,
	}, nil
}

func (c *Core) ReportInfo(ctx context.Context, req *sys.Info) (*sys.StatusMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportInfo not implemented")
}

func hash(input ...[]byte) []byte {
	bt := sha1.Sum(bytes.Join(input, nil))
	return bt[:]
}
