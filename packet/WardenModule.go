package packet

import (
	"github.com/superp00t/etc"
)

type WardenModuleInitRequest struct {
	Command1       uint8
	Size1          uint16
	Checksum1      uint32
	Unk1           uint8
	Unk2           uint8
	Type           uint8
	StringLibrary1 uint8
	Function1      []uint32

	Command2       uint8
	Size2          uint16
	Checksum2      uint32
	Unk3           uint8
	Unk4           uint8
	StringLibrary2 uint8
	Function2      uint32
	Function2Set   uint8

	Command3       uint8
	Size3          uint16
	Checksum3      uint32
	Unk5           uint8
	Unk6           uint8
	StringLibrary3 uint8
	Function3      uint32
	Function3Set   uint8
}

func (w *WardenModuleInitRequest) Encode() []byte {
	p := etc.NewBuffer()
	p.WriteByte(w.Command1)
	p.WriteUint16(w.Size1)
	p.WriteUint32(w.Checksum1)
	p.WriteByte(w.Unk1)
	p.WriteByte(w.Unk2)
	p.WriteByte(w.Type)
	p.WriteByte(w.StringLibrary1)
	for _, v := range w.Function1 {
		p.WriteUint32(v)
	}

	p.WriteByte(w.Command2)
	p.WriteUint16(w.Size2)
	p.WriteUint32(w.Checksum2)
	p.WriteByte(w.Unk3)
	p.WriteByte(w.Unk4)
	p.WriteByte(w.StringLibrary2)
	p.WriteUint32(w.Function2)
	p.WriteByte(w.Function2Set)

	p.WriteByte(w.Command3)
	p.WriteUint16(w.Size3)
	p.WriteUint32(w.Checksum3)
	p.WriteByte(w.Unk5)
	p.WriteByte(w.Unk6)
	p.WriteByte(w.StringLibrary3)
	p.WriteUint32(w.Function3)
	p.WriteByte(w.Function3Set)

	return p.Bytes()
}

func UnmarshalWardenModuleInitRequest(input []byte) (*WardenModuleInitRequest, error) {
	p := etc.FromBytes(input)
	w := new(WardenModuleInitRequest)

	w.Command1 = p.ReadByte()
	w.Size1 = p.ReadUint16()
	w.Checksum1 = p.ReadUint32()
	w.Unk1 = p.ReadByte()
	w.Unk2 = p.ReadByte()
	w.Type = p.ReadByte()
	w.StringLibrary1 = p.ReadByte()
	w.Function1 = make([]uint32, 4)
	for i := range w.Function1 {
		w.Function1[i] = p.ReadUint32()
	}

	w.Command2 = p.ReadByte()
	w.Size2 = p.ReadUint16()
	w.Checksum2 = p.ReadUint32()
	w.Unk3 = p.ReadByte()
	w.Unk4 = p.ReadByte()
	w.StringLibrary2 = p.ReadByte()
	w.Function2 = p.ReadUint32()
	w.Function2Set = p.ReadByte()

	w.Command3 = p.ReadByte()
	w.Size3 = p.ReadUint16()
	w.Checksum3 = p.ReadUint32()
	w.Unk5 = p.ReadByte()
	w.Unk6 = p.ReadByte()
	w.StringLibrary3 = p.ReadByte()
	w.Function3 = p.ReadUint32()
	w.Function3Set = p.ReadByte()

	return w, nil
}

type WardenModuleUse struct {
	Command   uint8
	ModuleID  []byte
	ModuleKey []byte
	Size      uint32
}

func UnmarshalWardenModuleUse(input []byte) (*WardenModuleUse, error) {
	e := etc.FromBytes(input)
	w := new(WardenModuleUse)
	w.Command = e.ReadByte()
	w.ModuleID = e.ReadBytes(16)
	w.ModuleKey = e.ReadBytes(16)
	w.Size = e.ReadUint32()
	return w, nil
}

func (w *WardenModuleUse) Encode() []byte {
	e := etc.NewBuffer()
	e.WriteByte(w.Command)
	e.Write(w.ModuleID)
	e.Write(w.ModuleKey)
	e.WriteUint32(w.Size)
	return e.Bytes()
}

type WardenModuleTransfer struct {
	Command  uint8
	DataSize uint16
	Data     []byte
}

func (w *WardenModuleTransfer) Encode() []byte {
	e := etc.NewBuffer()
	e.WriteByte(w.Command)
	e.WriteUint16(w.DataSize)
	e.Write(w.Data)
	return e.Bytes()
}

// func (w *WardenModuleInitRequest) Encode() []byte {
// 	g := new(bytes.Buffer)
// 	g.WriteByte(w.Command1)
// }
