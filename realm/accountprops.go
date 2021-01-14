package realm

import "github.com/superp00t/gophercraft/realm/wdb"

var NullProp wdb.PropID

var (
	ObjectDebug = wdb.MakePropID("dbgobj")
)

func (s *Session) HasProp(id wdb.PropID) bool {
	s.GuardProps.Lock()
	defer s.GuardProps.Unlock()
	for _, prop := range s.Props {
		if prop == id {
			return true
		}
	}
	return false
}

func (s *Session) AddProp(id wdb.PropID) {
	if s.HasProp(id) {
		return
	}
	s.GuardProps.Lock()
	defer s.GuardProps.Unlock()
	s.Props = append(s.Props, id)
	s.DB().Insert(&wdb.AccountProp{
		ID:   s.Account,
		Prop: id,
	})
}

func (s *Session) RemoveProp(id wdb.PropID) {
	s.GuardProps.Lock()
	defer s.GuardProps.Unlock()
	var index int = -1
	for i, element := range s.Props {
		if element == id {
			index = i
		}
	}

	if index == -1 {
		return
	}

	// Remove the element at index  from s.Props.
	s.Props[index] = s.Props[len(s.Props)-1] // Copy last element to index.
	s.Props[len(s.Props)-1] = NullProp       // Erase last element (write zero value).
	s.Props = s.Props[:len(s.Props)-1]       // Truncate slice.

	s.DB().Where("id = ?", s.Account).Where("prop = ?", id).Delete(new(wdb.AccountProp))
}
