package sys

func Code(s Status) *StatusMsg {
	return &StatusMsg{
		Status: s,
	}
}
