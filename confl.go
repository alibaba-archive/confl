package confl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/teambition/confl/etcd"
	"github.com/teambition/confl/vault"
)

var (
	DefaultChangeChan = 10
	ConfPathEnv       = "CONFL_CONF_PATH"
	ErrorNoConfPath   = fmt.Errorf("required env %s missing value", ConfPathEnv)
)

var (
	defautlOnError = func(err error) {
		if err != nil {
			log.Println(err)
		}
	}
)

// Hook hook type
// When configuration updates, then pass the copy of configuration to it
type Hook func(oldCfg, newCfg interface{})

// Watcher manage the watch states
type Watcher struct {
	confPath string
	// c the config struct user defined
	c        interface{}
	oCopyed  interface{}
	nCopyL   sync.RWMutex
	nCopyed  interface{}
	etcd     *etcd.Client
	changeCh chan struct{}
	hookL    sync.Mutex
	hooks    []Hook
	// OnError error handler
	OnError func(error)
}

// NewFromEnv create a config watcher from env
func NewFromEnv(c interface{}) (*Watcher, error) {
	confPath, _ := os.LookupEnv(ConfPathEnv)
	var (
		err       error
		etcdConf  = &etcd.Config{}
		vaultConf = &vault.Config{}
	)

	if err = envconfig.Process("", etcdConf); err != nil {
		return nil, err
	}

	if err = envconfig.Process("", vaultConf); err != nil {
		return nil, err
	}

	return New(c, confPath, etcdConf, vaultConf)
}

func New(c interface{}, confPath string, etcdConf *etcd.Config, vaultConf *vault.Config) (*Watcher, error) {
	var err error
	if confPath == "" {
		return nil, ErrorNoConfPath
	}

	w := &Watcher{
		c:        c,
		confPath: confPath,
		changeCh: make(chan struct{}, DefaultChangeChan),
		hooks:    []Hook{},
	}

	if w.etcd, err = etcd.NewClient(etcdConf, w.onError); err != nil {
		return nil, err
	}

	if err = vault.Init(vaultConf, w.changeCh, w.onError); err != nil {
		return nil, err
	}

	if err = w.loadConfig(); err != nil {
		return nil, err
	}
	return w, nil
}

// Config return the copy of w.c
// Example:
//   cfg := w.Config().(MyConfigStruct)
func (w *Watcher) Config() interface{} {
	w.nCopyL.RLock()
	defer w.nCopyL.RUnlock()
	return w.nCopyed
}

// AddHook add hooks for the update events of configuration
func (w *Watcher) AddHook(hooks ...Hook) {
	w.hookL.Lock()
	w.hooks = append(w.hooks, hooks...)
	w.hookL.Unlock()
}

func (w *Watcher) onError(err error) {
	if w.OnError != nil {
		w.OnError(err)
	} else {
		defautlOnError(err)
	}
}

// GoWatch start watch the update events
// It is blocked until the watcher is closed
func (w *Watcher) GoWatch() {
	go w.etcd.WatchKey(w.confPath, w.changeCh)
	w.procHooks()
}

// Close close the watcher
// Change channel must be closed finally in case of panic
func (w *Watcher) Close() error {
	w.etcd.Close()
	vault.Close()
	close(w.changeCh)
	return nil
}

// loadConfig load configuration from etcd by given conf_path
func (w *Watcher) loadConfig() error {
	v, err := w.etcd.Key(w.confPath)
	if err != nil {
		return err
	}

	// now configuration only support json type
	if err = json.Unmarshal([]byte(v), w.c); err != nil {
		return err
	}

	val := reflect.ValueOf(w.c)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}

	i := val.Interface()

	if reflect.ValueOf(w.nCopyed).IsValid() {
		w.oCopyed = w.nCopyed
	} else {
		w.oCopyed = i
	}

	w.nCopyL.Lock()
	w.nCopyed = i
	w.nCopyL.Unlock()
	return nil
}

// procHooks reloads config and runs the hooks when the watched value has changed
func (w *Watcher) procHooks() {
	for range w.changeCh {
		if err := w.loadConfig(); err != nil {
			w.onError(err)
			continue
		}

		// hooks must be called one by one
		// bcs there may be dependencies
		for _, hook := range w.hooks {
			hook(w.oCopyed, w.Config())
		}
	}
}
