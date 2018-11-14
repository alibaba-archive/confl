package confl

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Unmarshal - unmarshal function for FileWatcher
type Unmarshal func([]byte, interface{}) error

// fileWatcher watch the changes of configuration file
type fileWatcher struct {
	confPath  string
	c         interface{}
	oCopyed   interface{}
	nCopyL    sync.RWMutex
	nCopyed   interface{}
	w         *fsnotify.Watcher
	hookL     sync.Mutex
	hooks     []Hook
	errHandle func(error)
	unmarshal Unmarshal
}

// NewFileWatcher returns new a Watcher for file system
func NewFileWatcher(c interface{}, confPath string, fns ...Unmarshal) (Watcher, error) {
	unmarshal := json.Unmarshal
	if len(fns) > 0 {
		unmarshal = fns[0]
	}

	f := &fileWatcher{
		confPath:  confPath,
		c:         c,
		hooks:     []Hook{},
		errHandle: defautlOnError,
		unmarshal: unmarshal,
	}

	var err error
	if f.w, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}

	if err = f.w.Add(confPath); err != nil {
		return nil, err
	}

	if err = f.loadConfig(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *fileWatcher) loadConfig() error {
	fileData, err := ioutil.ReadFile(f.confPath)
	if err != nil {
		return err
	}
	fc := reflect.New(reflect.TypeOf(f.c).Elem()).Interface()
	if err = f.unmarshal(fileData, fc); err != nil {
		return err
	}

	f.oCopyed = f.nCopyed

	// w.c must be ptr type
	i := reflect.Indirect(reflect.ValueOf(fc)).Interface()
	f.nCopyL.Lock()
	f.nCopyed = i
	f.nCopyL.Unlock()
	return nil
}

func (f *fileWatcher) Config() interface{} {
	f.nCopyL.RLock()
	defer f.nCopyL.RUnlock()
	return f.nCopyed
}

func (f *fileWatcher) Watch() {
	for {
		select {
		case ev := <-f.w.Events:
			if ev.Op != fsnotify.Write {
				continue
			}
			if err := f.loadConfig(); err != nil {
				f.errHandle(err)
				continue
			}

			for _, hook := range f.hooks {
				hook(f.oCopyed, f.Config())
			}
		case err := <-f.w.Errors:
			f.errHandle(err)
		}
	}
}

func (f *fileWatcher) AddHook(h ...Hook) {
	f.hookL.Lock()
	f.hooks = append(f.hooks, h...)
	f.hookL.Unlock()
}

func (f *fileWatcher) OnError(h func(error)) {
	f.errHandle = h
}

func (f *fileWatcher) Close() error {
	return f.w.Close()
}
