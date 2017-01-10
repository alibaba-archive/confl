package confl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func (w *Watcher) AddHook(hooks ...Hook) {
	w.Lock()
	w.hooks = append(w.hooks, hooks...)
	w.Unlock()
}

func (w *Watcher) GoWatch() {
	go w.etcd.WatchKey(w.confPath, w.changeCh)
	w.runReloaders()
}

func (w *Watcher) Close() error {
	w.etcd.Close()
	vault.Close()
	close(w.changeCh)
	return nil
}

func (w *Watcher) loadConfig() error {
	v, err := w.etcd.Key(w.confPath)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(v), w.c)
}

// runReloaders run reloaders when the value changes
// which contained etcd and vault background storage
func (w *Watcher) runReloaders() {
	for range w.changeCh {
		if err := w.loadConfig(); err != nil {
			w.onError(err)
			continue
		}

		// reloaders have dependency order
		// need run reload one by one
		for _, hook := range w.hooks {
			hook(w.c)
		}
	}
}
