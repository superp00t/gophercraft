package realm

import (
	"fmt"
	"reflect"
	"sync"
)

type EventType int

const (
	AreaTriggerEvent EventType = iota
	GossipEvent
	ChatEvent
)

// Think on a particular scripting event.
func (ws *Server) ThinkOn(et EventType, scriptId interface{}, parameters ...interface{}) (shouldStop bool, err error) {
	eventType, ok := ws.eventMgr.Load(et)
	if !ok {
		err = fmt.Errorf("No event handler map for %d", et)
		return
	}

	handlers, ok := eventType.(*sync.Map).Load(scriptId)
	if !ok {
		err = fmt.Errorf("no script handler for %v", scriptId)
		return
	}

	for _, handler := range handlers.([]interface{}) {
		function := reflect.ValueOf(handler)

		reflectparameters := []reflect.Value{}
		for _, param := range parameters {
			reflectparameters = append(reflectparameters, reflect.ValueOf(param))
		}
		results := function.Call(reflectparameters)
		if len(results) > 0 {
			if results[0].Bool() == true && shouldStop == false {
				shouldStop = true
			}
		}
	}

	return
}

// describes the types exported to Gopherlua

// func (ws *Server) CallScript(returns interface{}, et EventType, key interface{}, args ...interface{}) (chan bool, bool) {
// 	var returnChan chan bool
// 	var outArg reflect.Value
// 	if returns != nil {
// 		fmt.Println("created return chan.")
// 		returnChan = make(chan bool)
// 		outArg = reflect.ValueOf(returns).Elem()
// 	}

// 	eventType, ok := ws.eventMgr.Load(et)
// 	if !ok {
// 		return nil, false
// 	}

// 	handlers, ok := eventType.(*sync.Map).Load(key)
// 	if !ok {
// 		return nil, false
// 	}

// 	ok = false

// 	for _, h := range handlers.([]interface{}) {
// 		switch fn := h.(type) {
// 		case *lua.LFunction:
// 			ws.scriptFunc <- func() error {
// 				fmt.Println("Invoking...")
// 				err := ws.ScriptEngine.Invoke(fn, args, outArg)
// 				fmt.Println("Invoked")
// 				if err != nil {
// 					fmt.Println("err", err, returnChan == nil)
// 					if returnChan != nil {
// 						returnChan <- false
// 					}
// 					return err
// 				}
// 				fmt.Println("callscript ok", err, returnChan == nil)
// 				if returnChan != nil {
// 					returnChan <- true
// 				}
// 				return nil
// 			}
// 			ok = true
// 		}
// 	}

// 	return returnChan, ok
// }

func (ws *Server) On(et EventType, key, handler interface{}) {
	// load event type handler
	m, _ := ws.eventMgr.LoadOrStore(et, new(sync.Map))
	eventType := m.(*sync.Map)

	ihandlers, _ := eventType.LoadOrStore(key, []reflect.Value{})

	handlers := ihandlers.([]reflect.Value)
	handlers = append(handlers, reflect.ValueOf(handler))

	eventType.Store(key, handlers)
}

func (ws *Server) OverrideEventHandler(et EventType, key, handler interface{}) {
	// load event type handler
	m, _ := ws.eventMgr.LoadOrStore(et, new(sync.Map))
	eventType := m.(*sync.Map)

	eventType.Store(key, []reflect.Value{
		reflect.ValueOf(handler),
	})
}

// func (s *Session) toPlayerData() *lua.LUserData {
// 	return s.WS.ScriptEngine.GetUserData("Player", s)
// }

// func (ws *Server) loadScripts() error {
// 	ws.ScriptEngine = script.NewEngine()
// 	// ws.ScriptEngine.SetCallback

// 	ws.ScriptEngine.WrapInterface("Map", &Map{}, nil, []string{
// 		"PlayObjectSound",
// 		"PlaySound",
// 		"PlayMusic",
// 	})

// 	ws.ScriptEngine.WrapInterface("GUID", guid.GUID{}, nil, []string{
// 		"String",
// 	})

// 	ws.ScriptEngine.WrapInterface("Creature", &Creature{}, nil, []string{
// 		"GUID",
// 	})

// 	ws.ScriptEngine.WrapInterface("Gossip", &packet.Gossip{}, packet.NewGossip, []string{
// 		"AddItem",
// 		"SetTextEntry",
// 		"GetSpeaker",
// 	})

// 	ws.ScriptEngine.WrapInterface("Player", &Session{}, nil, []string{
// 		"GUID",
// 		"Map",
// 		"Teleport",
// 		"GetLevel",
// 		"HasItem",
// 		"QuestDone",
// 		"SendAlertText",
// 		"SendRequiredLevelZoneError",
// 		"SendRequiredQuestZoneError",
// 		"SendRequiredItemZoneError",
// 		"SendGossip",
// 		"SendPlaySound",
// 		"SendPlayMusic",
// 		"GetLoc",
// 	})

// 	ws.ScriptEngine.WrapInterface("Instance", &Phase{}, nil, []string{
// 		"AddCreature",
// 	})

// 	ws.ScriptEngine.SetCallback("RandUint32", crypto.RandUint32)

// 	ws.ScriptEngine.SetCallback("GetInstance", func(instanceName string) *Phase {
// 		return ws.Phase(instanceName)
// 	})

// 	ws.ScriptEngine.SetCallback("OnAreaTrigger", func(triggerID uint32, cb *lua.LFunction) {
// 		fmt.Println("adding area trigger", triggerID)
// 		ws.AddEventHandler(AreaTriggerEvent, triggerID, cb)
// 	})

// 	ws.ScriptEngine.SetCallback("OnGossip", func(menuID string, cb *lua.LFunction) {
// 		ws.AddEventHandler(GossipEvent, menuID, cb)
// 	})

// 	ws.ScriptEngine.SetCallback("GetNPCTextEntry", func(textID string) uint32 {
// 		var npcText *wdb.NPCText
// 		ws.DB.GetData(textID, &npcText)
// 		if npcText == nil {
// 			return 0
// 		}

// 		return npcText.Entry
// 	})

// 	// Enums
// 	ws.ScriptEngine.SetEnum("GossipIconChat", packet.GossipIconChat)
// 	ws.ScriptEngine.SetEnum("GossipIconVendor", packet.GossipIconVendor)       // 1 Brown bag
// 	ws.ScriptEngine.SetEnum("GossipIconTaxi", packet.GossipIconTaxi)           // 2 Flight
// 	ws.ScriptEngine.SetEnum("GossipIconTrainer", packet.GossipIconTrainer)     // 3 Book
// 	ws.ScriptEngine.SetEnum("GossipIconInteract1", packet.GossipIconInteract1) // 4	Interaction wheel
// 	ws.ScriptEngine.SetEnum("GossipIconInteract2", packet.GossipIconInteract2) // 5	Interaction wheel
// 	ws.ScriptEngine.SetEnum("GossipIconGold", packet.GossipIconGold)           // 6 Brown bag with yellow dot (gold)
// 	ws.ScriptEngine.SetEnum("GossipIconTalk", packet.GossipIconTalk)           // White chat bubble with black dots (...)
// 	ws.ScriptEngine.SetEnum("GossipIconTabard", packet.GossipIconTabard)       // 8 Tabard
// 	ws.ScriptEngine.SetEnum("GossipIconBattle", packet.GossipIconBattle)       // 9 Two swords
// 	ws.ScriptEngine.SetEnum("GossipIconDot", packet.GossipIconDot)             // 10 Yellow dot
// 	ws.ScriptEngine.SetEnum("GossipIconChat11", packet.GossipIconChat11)       // 11	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconChat12", packet.GossipIconChat12)       // 12	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconChat13", packet.GossipIconChat13)       // 13	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconInvalid14", packet.GossipIconInvalid14) // 14	INVALID - DO NOT USE
// 	ws.ScriptEngine.SetEnum("GossipIconInvalid15", packet.GossipIconInvalid15) // 15	INVALID - DO NOT USE
// 	ws.ScriptEngine.SetEnum("GossipIconChat16", packet.GossipIconChat16)       // 16	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconChat17", packet.GossipIconChat17)       // 17	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconChat18", packet.GossipIconChat18)       // 18	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconChat19", packet.GossipIconChat19)       // 19	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconChat20", packet.GossipIconChat20)       // 20	White chat bubble
// 	ws.ScriptEngine.SetEnum("GossipIconTransmog", packet.GossipIconTransmog)   // 21	Transmogrifier?

// 	// load script
// 	for _, pack := range ws.PackLoader.Volumes {
// 		for _, sc := range pack.ServerScripts {
// 			entryPath := "Scripts/" + sc
// 			if pack.Exists(entryPath) {
// 				fl, err := pack.ReadFile(entryPath)
// 				if err != nil {
// 					return err
// 				}

// 				err = ws.ScriptEngine.DoReader(entryPath, fl)
// 				if err != nil {
// 					return err
// 				}

// 				fl.Close()
// 			} else {
// 				return fmt.Errorf("datapack %s's Pack.txt does not have a script named %s", pack.Name, entryPath)
// 			}
// 		}
// 	}

// 	go ws.scriptTask()

// 	return nil
// }

// func (ws *Server) scriptTask() {
// 	for {
// 		fn := <-ws.scriptFunc
// 		err := fn()
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }
