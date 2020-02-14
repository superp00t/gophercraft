package worldserver

import (
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"

	"github.com/superp00t/gophercraft/worldserver/script"
)

type EventType int

const (
	AreaTriggerEvent EventType = iota
)

// describes the types exported to Gopherlua

type PlayerScriptingInterface interface {
	Teleport(uint32, float32, float32, float32, float32)
	GetLevel() int
	HasItem(string) bool
	QuestDone(uint32) bool
	SendAlertText(string)
	SendRequiredLevelZoneError(int)
	SendRequiredQuestZoneError(uint32)
	SendRequiredItemZoneError(string)
}

func (ws *WorldServer) CallScript(et EventType, key interface{}, args ...interface{}) {
	eventType, ok := ws.eventMgr.Load(et)
	if !ok {
		return
	}

	handlers, ok := eventType.(*sync.Map).Load(key)
	if !ok {
		return
	}

	for _, h := range handlers.([]interface{}) {
		switch fn := h.(type) {
		case *lua.LFunction:
			ws.scriptFunc <- func() error {
				return ws.ScriptEngine.Invoke(fn, args)
			}
		}
	}
}

func (ws *WorldServer) AddEventHandler(et EventType, key, handler interface{}) {
	// load event type handler
	m, _ := ws.eventMgr.LoadOrStore(et, new(sync.Map))
	eventType := m.(*sync.Map)

	ihandlers, _ := eventType.LoadOrStore(key, []interface{}{})

	handlers := ihandlers.([]interface{})
	handlers = append(handlers, handler)

	eventType.Store(key, handlers)
}

func (s *Session) toPlayerData() *lua.LUserData {
	return s.WS.ScriptEngine.GetUserData("Player", PlayerScriptingInterface(s))
}

func (ws *WorldServer) loadScripts() error {
	ws.ScriptEngine = script.NewEngine()
	// ws.ScriptEngine.SetCallback

	ws.ScriptEngine.WrapInterface("Player", PlayerScriptingInterface(&Session{}), nil)

	ws.ScriptEngine.SetCallback("OnAreaTrigger", func(triggerID uint32, cb *lua.LFunction) {
		fmt.Println("adding area trigger", triggerID)
		ws.AddEventHandler(AreaTriggerEvent, triggerID, cb)
	})

	// load script
	for _, v := range ws.PackLoader.Volumes {
		for _, sc := range v.ServerScripts {
			entryPath := v.Name + ":Scripts/" + sc

			if ws.PackLoader.Exists(entryPath) {
				fl, err := ws.PackLoader.Open(entryPath)
				if err != nil {
					return err
				}

				err = ws.ScriptEngine.DoReader(entryPath, fl)
				if err != nil {
					return err
				}

				fl.Close()
			}
		}
	}

	go ws.scriptTask()

	return nil
}

func (ws *WorldServer) scriptTask() {
	for {
		fn := <-ws.scriptFunc
		err := fn()
		if err != nil {
			panic(err)
		}
	}
}
