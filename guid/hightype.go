package guid

//go:generate gcraft_stringer -type=HighType
type HighType uint64

const (
	Null             HighType = 0
	Uniq             HighType = 1
	Player           HighType = 2
	Item             HighType = 3
	WorldTransaction HighType = 4
	StaticDoor       HighType = 5 //NYI
	Transport        HighType = 6
	Conversation     HighType = 7
	Creature         HighType = 8
	Vehicle          HighType = 9
	Pet              HighType = 10
	GameObject       HighType = 11
	DynamicObject    HighType = 12
	AreaTrigger      HighType = 13
	Corpse           HighType = 14
	LootObject       HighType = 15
	SceneObject      HighType = 16
	Scenario         HighType = 17
	AIGroup          HighType = 18
	DynamicDoor      HighType = 19
	ClientActor      HighType = 20 //NYI
	Vignette         HighType = 21
	CallForHelp      HighType = 22
	AIResource       HighType = 23
	AILock           HighType = 24
	AILockTicket     HighType = 25
	ChatChannel      HighType = 26
	Party            HighType = 27
	Guild            HighType = 28
	WowAccount       HighType = 29
	BNetAccount      HighType = 30
	GMTask           HighType = 31
	MobileSession    HighType = 32 //NYI
	RaidGroup        HighType = 33
	Spell            HighType = 34
	Mail             HighType = 35
	WebObj           HighType = 36 //NYI
	LFGObject        HighType = 37 //NYI
	LFGList          HighType = 38 //NYI
	UserRouter       HighType = 39
	PVPQueueGroup    HighType = 40
	UserClient       HighType = 41
	PetBattle        HighType = 42 //NYI
	UniqUserClient   HighType = 43
	BattlePet        HighType = 44
	CommerceObj      HighType = 45
	ClientSession    HighType = 46
	Cast             HighType = 47
	ClientConnection HighType = 48
	// virtual gophercraft codes for back compatibility
	Mo_Transport HighType = 60
	Instance     HighType = 61
	Group        HighType = 62
)
