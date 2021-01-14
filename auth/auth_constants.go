package auth

//go:generate gcraft_stringer -type=AuthType
//go:generate gcraft_stringer -type=ErrorType
const (
	WOW_SUCCESS                              ErrorType = 0x00
	WOW_FAIL_BANNED                          ErrorType = 0x03
	WOW_FAIL_UNKNOWN_ACCOUNT                 ErrorType = 0x04
	WOW_FAIL_INCORRECT_PASSWORD              ErrorType = 0x05
	WOW_FAIL_ALREADY_ONLINE                  ErrorType = 0x06
	WOW_FAIL_NO_TIME                         ErrorType = 0x07
	WOW_FAIL_DB_BUSY                         ErrorType = 0x08
	WOW_FAIL_VERSION_INVALID                 ErrorType = 0x09
	WOW_FAIL_VERSION_UPDATE                  ErrorType = 0x0A
	WOW_FAIL_INVALID_SERVER                  ErrorType = 0x0B
	WOW_FAIL_SUSPENDED                       ErrorType = 0x0C
	WOW_FAIL_FAIL_NOACCESS                   ErrorType = 0x0D
	WOW_SUCCESS_SURVEY                       ErrorType = 0x0E
	WOW_FAIL_PARENTCONTROL                   ErrorType = 0x0F
	WOW_FAIL_LOCKED_ENFORCED                 ErrorType = 0x10
	WOW_FAIL_TRIAL_ENDED                     ErrorType = 0x11
	WOW_FAIL_USE_BATTLENET                   ErrorType = 0x12
	WOW_FAIL_ANTI_INDULGENCE                 ErrorType = 0x13
	WOW_FAIL_EXPIRED                         ErrorType = 0x14
	WOW_FAIL_NO_GAME_ACCOUNT                 ErrorType = 0x15
	WOW_FAIL_CHARGEBACK                      ErrorType = 0x16
	WOW_FAIL_INTERNET_GAME_ROOM_WITHOUT_BNET ErrorType = 0x17
	WOW_FAIL_GAME_ACCOUNT_LOCKED             ErrorType = 0x18
	WOW_FAIL_UNLOCKABLE_LOCK                 ErrorType = 0x19
	WOW_FAIL_CONVERSION_REQUIRED             ErrorType = 0x20
	WOW_FAIL_DISCONNECTED                    ErrorType = 0xFF

	LOGIN_OK               uint8 = 0x00
	LOGIN_FAILED           uint8 = 0x01
	LOGIN_FAILED2          uint8 = 0x02
	LOGIN_BANNED           uint8 = 0x03
	LOGIN_UNKNOWN_ACCOUNT  uint8 = 0x04
	LOGIN_UNKNOWN_ACCOUNT3 uint8 = 0x05
	LOGIN_ALREADYONLINE    uint8 = 0x06
	LOGIN_NOTIME           uint8 = 0x07
	LOGIN_DBBUSY           uint8 = 0x08
	LOGIN_BADVERSION       uint8 = 0x09
	LOGIN_DOWNLOAD_FILE    uint8 = 0x0A
	LOGIN_FAILED3          uint8 = 0x0B
	LOGIN_SUSPENDED        uint8 = 0x0C
	LOGIN_FAILED4          uint8 = 0x0D
	LOGIN_CONNECTED        uint8 = 0x0E
	LOGIN_PARENTALCONTROL  uint8 = 0x0F
	LOGIN_LOCKED_ENFORCED  uint8 = 0x10

	POST_BC_EXP_FLAG  uint8 = 0x2
	PRE_BC_EXP_FLAG   uint8 = 0x1
	NO_VALID_EXP_FLAG uint8 = 0x0

	AUTH_LOGON_CHALLENGE     AuthType = 0x00
	AUTH_LOGON_PROOF         AuthType = 0x01
	AUTH_RECONNECT_CHALLENGE AuthType = 0x02
	AUTH_RECONNECT_PROOF     AuthType = 0x03
	REALM_LIST               AuthType = 0x10
	XFER_INITIATE            AuthType = 0x30
	XFER_DATA                AuthType = 0x31
	XFER_ACCEPT              AuthType = 0x32
	XFER_RESUME              AuthType = 0x33
	XFER_CANCEL              AuthType = 0x34

	REALM_GREEN  uint8 = 0
	REALM_YELLOW uint8 = 1
	REALM_RED    uint8 = 2
)

type AuthType uint8
type ErrorType uint8
