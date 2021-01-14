package update

import "fmt"

func (mb *MovementBlock) writeDataAlpha(e *Encoder) error {
	if err := EncodeMovementInfo(e.Build, e, mb.Info); err != nil {
		return err
	}

	writeUint32(e, 0) // fall time?

	slist, ok := SpeedLists[e.Build]
	if !ok {
		return fmt.Errorf("update: no speed list found for %s", e.Build)
	}

	for _, sType := range slist {
		writeFloat32(e, mb.Speeds[sType])
	}

	if mb.UpdateFlags&UpdateFlagSelf != 0 {
		writeUint32(e, 1)
	} else {
		writeUint32(e, 0)
	}

	writeUint32(e, 1)
	writeUint32(e, 0)

	mb.Victim.EncodeUnpacked(e.Build, e)

	return nil
}
