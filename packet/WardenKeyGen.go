package packet

func NewSHA1Randx(buff []byte, size uint32) *SHA1Randx {
	x := &SHA1Randx{}
	halfSize := size / 2
	x.O1 = Hash(buff[:halfSize])
	x.O2 = Hash(buff[halfSize:])
	x.O0 = make([]byte, 20)
	x.FillUp()
	return x
}

type SHA1Randx struct {
	Taken      uint32
	O0, O1, O2 []byte //20 long
}

func (x *SHA1Randx) Generate(buf []byte, size int) {
	for i := 0; i < int(size); i++ {
		if x.Taken == 20 {
			x.FillUp()
		}

		buf[i] = x.O0[x.Taken]
		x.Taken++
	}
}

func (x *SHA1Randx) FillUp() {
	x.O0 = Hash(x.O1, x.O0, x.O2)
	x.Taken = 0
}
