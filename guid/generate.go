package guid

func New(high, low uint64) GUID {
	return GUID{high, low}
}

func Global(t HighType, low uint64) GUID {
	return New(uint64(t)<<58, low)
}

func RealmSpecific(t HighType, realmID, low uint64) GUID {
	return New((uint64(t)<<58)|(realmID<<42), low)
}

func MapSpecific(t HighType, realmID uint64, subType uint8, mapID uint16, serverID uint32, entry uint32, counter uint32) GUID {
	return New(uint64((uint64(t)<<58)|(uint64(realmID&0x1FFF)<<42)|(uint64(mapID&0x1FFF)<<29)|(uint64(entry&0x7FFFFF)<<6)|(uint64(subType)&0x3F)),
		uint64((uint64(serverID&0xFFFFFF)<<40)|(uint64(counter)&0xFFFFFFFFFF)))
}
