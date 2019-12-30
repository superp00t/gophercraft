package packet

const (
	AUTH_OK                     uint8 = 0x0C
	AUTH_FAILED                 uint8 = 0x0D
	AUTH_REJECT                 uint8 = 0x0E
	AUTH_BAD_SERVER_PROOF       uint8 = 0x0F
	AUTH_UNAVAILABLE            uint8 = 0x10
	AUTH_SYSTEM_ERROR           uint8 = 0x11
	AUTH_BILLING_ERROR          uint8 = 0x12
	AUTH_BILLING_EXPIRED        uint8 = 0x13
	AUTH_VERSION_MISMATCH       uint8 = 0x14
	AUTH_UNKNOWN_ACCOUNT        uint8 = 0x15
	AUTH_INCORRECT_PASSWORD     uint8 = 0x16
	AUTH_SESSION_EXPIRED        uint8 = 0x17
	AUTH_SERVER_SHUTTING_DOWN   uint8 = 0x18
	AUTH_ALREADY_LOGGING_IN     uint8 = 0x19
	AUTH_LOGIN_SERVER_NOT_FOUND uint8 = 0x1A
	AUTH_WAIT_QUEUE             uint8 = 0x1B
	AUTH_BANNED                 uint8 = 0x1C
	AUTH_ALREADY_ONLINE         uint8 = 0x1D
	AUTH_NO_TIME                uint8 = 0x1E
	AUTH_DB_BUSY                uint8 = 0x1F
	AUTH_SUSPENDED              uint8 = 0x20
	AUTH_PARENTAL_CONTROL       uint8 = 0x21
	AUTH_LOCKED_ENFORCED        uint8 = 0x22
)
