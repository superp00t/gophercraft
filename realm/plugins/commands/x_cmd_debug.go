package commands

import (
	"os"
	"runtime"
	"strings"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/realm"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func cmdDebugInv(s *realm.Session) {
	s.Warnf("Player:")

	for i := 0; i < 39; i++ {
		gid := s.GetGUID("InventorySlots", i)
		if gid != guid.Nil {
			s.Warnf(" %d: %s", i, gid)
		}
	}

	for i := 19; i < 23; i++ {
		g := s.GetGUID("InventorySlots", i)
		if g != guid.Nil {
			s.Warnf("Bag %d:", i)

			gArray := s.GetBagItem(uint8(i)).Get("Slots")

			for idx := 0; idx < gArray.Len(); idx++ {
				it := gArray.Index(idx).Interface().(guid.GUID)
				if it != guid.Nil {
					s.Warnf(" %d: %s", idx, it)
				}
			}
		}
	}
}

// func getObjectArgument(c *C) (guid.GUID, error) {

// }

// func x_ForceUpdate(c *C) {

// }

func cmdGoroutines(s *realm.Session) {
	// pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	buf := make([]byte, 1<<20)
	stacklen := runtime.Stack(buf, true)
	os.Stdout.Write(buf[:stacklen])
}

func cmdShowSQL(s *realm.Session, on bool) {
	s.DB().ShowSQL(on)
}

func cmdTrackedGUIDs(s *realm.Session) {
	s.GuardTrackedGUIDs.Lock()
	defer s.GuardTrackedGUIDs.Unlock()

	s.Warnf("%d tracked GUIDs:", len(s.TrackedGUIDs))

	for _, id := range s.TrackedGUIDs {
		s.Warnf("%s", s.DebugGUID(id))
	}
}

func cmdListProps(s *realm.Session) {
	var list []string
	for _, prop := range s.Props {
		list = append(list, prop.String())
	}
	s.Warnf("Props: %s", strings.Join(list, ", "))
}

func cmdAddProp(s *realm.Session, propId string) {
	if len(propId) > 8 {
		s.Warnf("too many characters for a prop ID")
		return
	}

	id := wdb.MakePropID(propId)
	s.AddProp(id)
}

func cmdRemoveProp(s *realm.Session, propId string) {
	if len(propId) > 8 {
		s.Warnf("too many characters for a prop ID")
		return
	}
	id := wdb.MakePropID(propId)

	hp := s.HasProp(id)

	s.RemoveProp(id)

	if hp {
		s.Warnf("Prop %s removed", id)
	}
}
