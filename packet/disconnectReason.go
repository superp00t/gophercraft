package packet

import "fmt"

type DCReason uint32

const (
	InvalidServerHeader = 3
	RSAVerifyFailed     = 4
	ServerCheckFailed   = 24
)

func (dc DCReason) Error() string {
	reason := "client disconnected "
	switch dc {
	case InvalidServerHeader:
		reason += "because of a protocol error"
	case RSAVerifyFailed:
		reason += "due to an RSA signature failure"
	case ServerCheckFailed:
		reason += "because an HMAC server check failed"
	default:
		reason += "for an unknown reason"
	}

	reason += fmt.Sprintf("(error code %d)", dc)
	return reason
}
