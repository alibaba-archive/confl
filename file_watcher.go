package confl

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"sync"

	"github.com/howeyc/fsnotify"
)

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
}

// NewFileWatcher new a Watcher for file system
func NewFileWatcher(c interface{}, confPath string) (*fileWatcher, error) {
	f := &fileWatcher{
		confPath:  confPath,
		c:         c,
		hooks:     []Hook{},
		errHandle: defautlOnError,
	}

	var err error
	if f.w, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}

	if err = f.w.WatchFlags(confPath, fsnotify.FSN_MODIFY); err != nil {
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

	if err = json.Unmarshal(fileData, f.c); err != nil {
		return err
	}

	f.oCopyed = f.nCopyed

	// w.c must be ptr type
	i := reflect.Indirect(reflect.ValueOf(f.c)).Interface()
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
		case ev := <-f.w.Event:
			if ev.IsAttrib() {
				continue
			}
			if err := f.loadConfig(); err != nil {
				f.errHandle(err)
				continue
			}

			for _, hook := range f.hooks {
				hook(f.oCopyed, f.Config())
			}
		case err := <-f.w.Error:
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
