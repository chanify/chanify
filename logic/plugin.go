package logic

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	lua "github.com/yuin/gopher-lua"
)

type luaFunc struct {
	modTime time.Time
	lfunc   *lua.FunctionProto
}

// Webhook item
type Webhook struct {
	name  string
	env   map[string]interface{}
	lfunc *luaFunc
}

type pluginManager struct {
	watcher  *fsnotify.Watcher
	luaMutex sync.Mutex
	luaFiles map[string]*luaFunc
	webHooks map[string]*Webhook
}

func loadWebhookPlugin(path string, wbOpts []map[string]interface{}) *pluginManager {
	log.Println("Load plugins:", path)
	plugin := &pluginManager{
		luaFiles: make(map[string]*luaFunc),
		webHooks: make(map[string]*Webhook),
	}
	watcher, err := fsnotify.NewWatcher()
	if err == nil {
		plugin.watcher = watcher
		go plugin.luaWatch()
	}
	for _, opts := range wbOpts {
		if name, ok := readOptString(opts, "name"); ok && len(name) > 0 {
			if wh, ok := plugin.webHooks[name]; ok && wh.lfunc.lfunc != nil {
				log.Println("Webhook plugin conflict:", name)
				continue
			}
			if file, ok := readOptString(opts, "file"); ok && len(file) > 0 {
				if lfunc := plugin.loadLuaFile(file, path); lfunc != nil {
					webhook := &Webhook{
						name:  name,
						env:   make(map[string]interface{}),
						lfunc: lfunc,
					}
					for k, v := range readOptTable(opts, "env") {
						webhook.env[strings.ToLower(k)] = v
					}
					plugin.webHooks[name] = webhook
					log.Println("Load webhook plugin:", name)
				}
			}
		}
	}
	return plugin
}

func (p *pluginManager) Close() {
	if p.watcher != nil {
		p.watcher.Close()
		p.watcher = nil
	}
}

func (p *pluginManager) GetWebhook(name string) (*Webhook, error) {
	p.luaMutex.Lock()
	defer p.luaMutex.Unlock()
	if webhook, ok := p.webHooks[strings.ToLower(name)]; ok {
		lfunc := webhook.lfunc
		if lfunc != nil && lfunc.lfunc != nil {
			return webhook, nil
		}
	}
	return nil, ErrNotFound
}

func (p *pluginManager) loadLuaFile(file string, path string) *luaFunc {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file = filepath.Join(path, file)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return nil
		}
	}
	file, _ = filepath.Abs(file)
	p.luaMutex.Lock()
	defer p.luaMutex.Unlock()
	lfunc, ok := p.luaFiles[file]
	if !ok {
		lfunc = &luaFunc{}
		if err := lfunc.Reload(file); err != nil {
			log.Println("Load lua failed:", err)
		}
		p.luaFiles[file] = lfunc
		p.watcher.Add(file) // nolint: errcheck
	}
	return lfunc
}

func (p *pluginManager) luaWatch() {
	watcher := p.watcher
	if watcher != nil {
		for {
			select {
			case err, ok := <-watcher.Errors:
				if ok {
					log.Println("Watch lua file failed:", err)
				}
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					p.ReloadWebhook(event.Name)
				}
			}
		}
	}
}

func (p *pluginManager) ReloadWebhook(file string) {
	p.luaMutex.Lock()
	defer p.luaMutex.Unlock()
	if lfunc, ok := p.luaFiles[file]; ok {
		if err := lfunc.Reload(file); err != nil {
			log.Println("Reload lua failed:", err)
		} else {
			log.Println("Reload lua file:", file)
		}
	}
}

func (lf *luaFunc) Reload(file string) error {
	if s, err := os.Stat(file); err == nil {
		modTime := s.ModTime()
		if !modTime.Equal(lf.modTime) {
			lfunc, err := compileLua(file)
			if err != nil {
				return err
			} else {
				lf.modTime = modTime
				lf.lfunc = lfunc
			}
		}
	}
	return nil
}

func (w *Webhook) DoCall(l *lua.LState) error {
	if w.lfunc == nil {
		return ErrNotFound
	}
	lfunc := w.lfunc.lfunc
	if lfunc == nil {
		return ErrNotFound
	}
	initLua(l)
	if ctx, ok := l.GetGlobal("ctx").(*lua.LUserData); ok {
		if mtbl, ok := ctx.Metatable.(*lua.LTable); ok {
			if ftbl, ok := mtbl.RawGetString("__index").(*lua.LTable); ok {
				ftbl.RawSetString("env", l.NewFunction(func(ll *lua.LState) int {
					key := strings.ToLower(ll.CheckString(2))
					if val, ok := w.env[key]; ok {
						ll.Push(luaInterface2LValue(val))
						return 1
					}
					ll.Push(lua.LNil)
					return 1
				}))
			}
		}
	}
	l.Push(l.NewFunctionFromProto(lfunc))
	return l.PCall(0, lua.MultRet, nil)
}
