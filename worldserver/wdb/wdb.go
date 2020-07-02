package wdb

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"sync"

	_ "github.com/superp00t/gophercraft/gcore/dbsupport"
	"xorm.io/xorm"
)

var EntryStore sync.Map
var DataStores = map[reflect.Type]*Store{}

type Store struct {
	store sync.Map
}

func (s *Store) Range(fn func(k, v interface{}) bool) {
	s.store.Range(fn)
}

type Core struct {
	*xorm.Engine
}

func ClearData(typ reflect.Type) {
	if DataStores[typ] != nil {
		delete(DataStores, typ)
	}
}

func GetData(key, ptrTo interface{}) {
	ptr := reflect.ValueOf(ptrTo)
	if ptr.Kind() != reflect.Ptr && ptr.Elem().Kind() != reflect.Ptr {
		panic("must be ptr to ptr")
	}

	datType := ptr.Type().Elem().Elem()

	store := DataStores[datType]

	if store == nil {
		fmt.Println("no data store found for", datType)
		for k := range DataStores {
			fmt.Println("However, ", k, "data store does exist")
		}
		return
	}

	result, ok := store.store.Load(key)
	if !ok {
		fmt.Println("no data found for", key)
		return
	}

	ptr.Elem().Set(reflect.ValueOf(result))
}

func StoreData(value reflect.Value) {
	if value.Kind() != reflect.Ptr {
		panic("not a pointer " + value.Type().String())
	}
	base := value.Elem()
	store := DataStores[base.Type()]

	if store == nil {
		store = &Store{}
		DataStores[base.Type()] = store
	}

	idField := base.FieldByName("ID")
	if idField.Kind() == reflect.Uint32 {
		store.store.Store(uint32(idField.Uint()), value.Interface())
	} else if idField.Kind() == reflect.String {
		store.store.Store(idField.String(), value.Interface())
	} else {
		panic(idField.Kind())
	}

	entryField := base.FieldByName("Entry")
	if entryField.IsValid() {
		entry := uint32(base.FieldByName("Entry").Uint())
		store.store.Store(entry, value.Interface())
	}
}

func NewCore(driver, source string) (*Core, error) {
	var err error
	cn := new(Core)
	cn.Engine, err = xorm.NewEngine(driver, source)
	if err != nil {
		return nil, err
	}

	err = cn.Engine.Sync2(
		new(Character),
		new(Item),
		new(Inventory),
		new(ObjectTemplateRegistry),

		new(Contact),
	)

	if err != nil {
		return nil, err
	}

	_, err = cn.Count(new(Character))
	if err != nil {
		return nil, err
	}

	return cn, nil
}

func GetID(value reflect.Value) string {
	return value.FieldByName("ID").String()
}

func GetName(value reflect.Value) string {
	name := value.FieldByName("Name")
	if name.IsValid() == false {
		return GetID(value)
	}

	return name.String()
}

func SetEntry(value reflect.Value, entry uint32) {
	value.FieldByName("Entry").SetUint(uint64(entry))
}

func (c *Core) GetGameObjectTemplate(id string) (*GameObjectTemplate, error) {
	var gobj *GameObjectTemplate
	GetData(id, &gobj)
	if gobj == nil {
		return nil, fmt.Errorf("no game object template by the ID %s", id)
	}
	return gobj, nil
}

func (c *Core) GetGameObjectTemplateByEntry(entry uint32) (*GameObjectTemplate, error) {
	var gobj *GameObjectTemplate
	GetData(entry, &gobj)
	if gobj == nil {
		return nil, fmt.Errorf("no game object template by the entry %d", entry)
	}
	return gobj, nil
}

func (c *Core) GetItemTemplate(id string) (*ItemTemplate, error) {
	var item *ItemTemplate
	GetData(id, &item)
	if item == nil {
		return nil, fmt.Errorf("no ItemTemplate by the ID %s", id)
	}
	return item, nil
}

func (c *Core) GetItemTemplateByEntry(entry uint32) (*ItemTemplate, error) {
	var item *ItemTemplate
	GetData(entry, &item)
	if item == nil {
		return nil, fmt.Errorf("no ItemTemplate by the entry %d", entry)
	}
	return item, nil
}

type sortableTemplateSlice struct {
	reflect.Value
}

func (v sortableTemplateSlice) Swap(i, j int) {
	x, y := v.Index(i).Interface(), v.Index(j).Interface()
	v.Index(i).Set(reflect.ValueOf(y))
	v.Index(j).Set(reflect.ValueOf(x))
}

func (a sortableTemplateSlice) Less(i, j int) bool {
	return GetName(a.Index(i).Elem()) < GetName(a.Index(j).Elem())
}

func SortNamedTemplates(value reflect.Value) {
	sort.Sort(sortableTemplateSlice{value})
}

func SearchTemplates(nameString string, max int64, sliceInterface interface{}) error {
	slice := reflect.ValueOf(sliceInterface)

	//                   ptr    slice  ptr
	elementType := slice.Elem().Type().Elem().Elem()

	regex, err := regexp.Compile("(?i)" + nameString)
	if err != nil {
		return err
	}

	store := DataStores[elementType]
	if store == nil {
		return fmt.Errorf("wdb: could not find data store for %s", elementType)
	}

	idType, ok := elementType.FieldByName("ID")

	if !ok {
		panic(elementType.String() + " has no ID field")
	}

	scanned := int64(0)

	store.store.Range(func(k, v interface{}) bool {
		if scanned > max {
			return false
		}

		keyType := reflect.TypeOf(k)
		if keyType == idType.Type {
			if regex.MatchString(GetName(reflect.ValueOf(v).Elem())) {
				slice.Elem().Set(reflect.Append(slice.Elem(), reflect.ValueOf(v)))
				scanned++
			}
		}

		return true
	})

	SortNamedTemplates(slice.Elem())

	return nil
}

func GetStore(in interface{}) *Store {
	val := reflect.TypeOf(in)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Slice {
		val = val.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	st := DataStores[val]
	if st == nil {
		panic("no datastore for type " + val.String())
	}

	return st
}
