package realm

import (
	"reflect"
	"sort"

	"github.com/superp00t/etc/yo"
)

type PluginInfo struct {
	ID string
	// Display name
	Name       string
	Descriptor string
	Authors    []string
	Version    string
}

type LoadedPlugin struct {
	PluginInfo
	Plugin
}

type Plugin interface {
	// For checking whether a plugin is actually operating, i.e. determine if there is a non-fatal error
	Activated() (activated bool, reason error)
	// Gracefully shutdown a plugin. The operation should fully complete before this function returns
	Terminate() error
	// Initialize the plugin's main functionality, and set the metadata
	Init(server *Server, plg *PluginInfo) error
}

var plugins = map[string]reflect.Type{}

func RegisterPlugin(name string, plugin Plugin) {
	if plugins[name] != nil {
		panic("realm: Plugin " + name + " is already installed.")
	}

	t := reflect.TypeOf(plugin)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	plugins[name] = t
}

// safely (re)load plugins
func (s *Server) loadPlugins() error {
	for _, v := range s.Plugins {
		if err := v.Terminate(); err != nil {
			yo.Warn(err)
		}
	}

	s.Plugins = []*LoadedPlugin{}

	for name, plugin := range plugins {
		newPlugin := LoadedPlugin{}
		newPlugin.Plugin = reflect.New(plugin).Interface().(Plugin)
		newPlugin.PluginInfo.ID = name
		s.Plugins = append(s.Plugins, &newPlugin)
	}

	sort.Slice(s.Plugins, func(i, j int) bool {
		return s.Plugins[i].ID < s.Plugins[j].ID
	})

	for _, loadedPlugin := range s.Plugins {
		if err := loadedPlugin.Init(s, &loadedPlugin.PluginInfo); err != nil {
			return err
		}
	}

	return nil
}
