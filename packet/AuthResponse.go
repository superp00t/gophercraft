package packet

import "github.com/superp00t/gophercraft/vsn"

type VirtualRealmNameInfo struct {
	IsLocal             bool
	IsInternalRealm     bool
	RealmNameActual     string
	RealmNameNormalized string
}

type VirtualRealmInfo struct {
	VirtualAddress uint32
	VirtualRealmNameInfo
}

type ClassAvailability struct {
	ClassID               uint8
	ActiveExpansionLevel  uint8
	AccountExpansionLevel uint8
}

type RaceClassAvailability struct {
	RaceID  uint8
	Classes []ClassAvailability
}

type AuthResponse struct {
	Result          uint32
	SuccessInfoInit bool
	WaitInfoInit    bool

	SuccessInfo struct {
		VirtualRealmAddress    uint32
		TimeRested             uint32
		ActiveExpansionLevel   uint32
		AccountExpansionLevel  uint32
		TimeSecondsUntilPCKick uint32
		CurrencyID             uint32
		Time                   uint32

		AvailableClasses []RaceClassAvailability

		IsExpansionTrial bool
		SuccessInfoTrial
	}
}

func (ar *AuthResponse) Encode(build vsn.Build, to *WorldPacket) error {

}
