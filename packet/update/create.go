package update

import "github.com/superp00t/gophercraft/guid"

func (cb *CreateBlock) WriteTo(objectGUID guid.GUID, e *Encoder) error {
	err := guid.EncodeTypeID(e.Version, cb.ObjectType, e.Buffer)
	if err != nil {
		return err
	}
	if err = cb.MovementBlock.WriteTo(objectGUID, e); err != nil {
		return err
	}
	if err = cb.ValuesBlock.WriteTo(objectGUID, e); err != nil {
		return err
	}
	return nil
}

func (cb *CreateBlock) Type() BlockType {
	return cb.BlockType
}
