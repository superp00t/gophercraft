package update

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/guid"
)

func Marshal(version uint32, mask ValueMask, data *Data) (out []byte, err error) {
	e := &Encoder{
		etc.NewBuffer(),
		mask,
		version,
	}

	count := uint32(len(data.Blocks))
	e.WriteUint32(count)
	e.WriteBool(data.HasTransport)

	// delete blocks should be serialized first (or so I think, MaNGOS does it like this)
	for _, pDelete := range data.Blocks {
		t := pDelete.Data.Type()

		if t == DeleteNearObjects || t == DeleteFarObjects {
			e.WriteByte(uint8(t))
			if err := pDelete.Data.WriteTo(guid.None, e); err != nil {
				return nil, err
			}
		}
	}

	// blocks with GUID prefixes
	for _, block := range data.Blocks {
		t := block.Data.Type()
		if t != DeleteNearObjects && t != DeleteFarObjects {
			e.WriteByte(uint8(t))
			block.GUID.EncodePacked(version, e)
			if err := block.Data.WriteTo(block.GUID, e); err != nil {
				return nil, err
			}
		}
	}

	return e.Buffer.Bytes(), nil
}
