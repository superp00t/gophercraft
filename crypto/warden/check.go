package warden

type Check struct {
	ID      uint16
	Type    uint8
	Data    string
	Str     string
	Address int
	Length  uint8
	Result  string
	Comment string
}

func usableType(i uint8) bool {
	return true
	// return i == packet.MODULE_CHECK || i == packet.DRIVER_CHECK || i == packet.PAGE_CHECK_B || i == packet.MEM_CHECK || i == packet.MPQ_CHECK || i == packet.LUA_STR_CHECK
	// return i == packet.MODULE_CHECK || i == packet.MEM_CHECK || i == packet.MPQ_CHECK || i == packet.LUA_STR_CHECK
}

func GetChecks() []Check {
	var c []Check
	for _, v := range ChecksDB[1:] {
		if usableType(v.Type) {
			c = append(c, v)
		}
	}

	return c
}
