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
type Hook func(config interface{})

// Watcher manage the watch states
type Watcher struct {
	sync.Mutex
	confPath string
	// c the config struct user defined
	c        interface{}
	etcd     *etcd.Client
	changeCh chan struct{}
	hooks    []Hook
	onError  func(error)
}

// NewFromEnv create a config watcher from env
func NewFromEnv(c interface{}, onError func(error)) (*Watcher, error) {
	if onError == nil {
		onError = defautlOnError
	}

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

	return New(c, confPath, etcdConf, vaultConf, onError)
}

func New(c interface{}, confPath string, etcdConf *etcd.Config, vaultConf *vault.Config, onError func(error)) (*Watcher, error) {
	var err error
	if confPath == "" {
		return nil, ErrorNoConfPath
	}

	if onError == nil {
		onError = defautlOnError
	}

	if etcdConf.OnError == nil {
		etcdConf.OnError = onError
	}

	if vaultConf.OnError == nil {
		vaultConf.OnError = onError
	}

	w := &Watcher{
		c:        c,
		confPath: confPath,
		changeCh: make(chan struct{}, DefaultChangeChan),
		hooks:    []Hook{},
		onError:  onError,
	}

	if w.etcd, err = etcd.NewClient(etcdConf); err != nil {
		return nil, err
	}

	vaultConf.ChangeCh = w.changeCh
	if err = vault.Init(vaultConf); err != nil {
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
	val := reflect.ValueOf(w.c)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	return val.Interface()
}

// AddHook add hooks for the update events of configuration
func (w *Watcher) AddHook(hooks ...Hook) {
	w.Lock()
	w.hooks = append(w.hooks, hooks...)
	w.Unlock()
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
	return json.Unmarshal([]byte(v), w.c)
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
			hook(w.Config())
		}
	}
}
