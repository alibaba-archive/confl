package confl

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/teambition/confl/etcd"
	"github.com/teambition/confl/vault"
)

const (
	DefaultChangeChan = 10
)

// storeWatcher watch the changes from backend storages
type storeWatcher struct {
	confPath string
	// c the config struct user defined
	c         interface{}
	oCopyed   interface{}
	nCopyL    sync.RWMutex
	nCopyed   interface{}
	etcd      *etcd.Client
	vault     *vault.Client
	changeCh  chan struct{}
	hookL     sync.Mutex
	hooks     []Hook
	errHandle func(error)
}

type (
	Etcd  etcd.Config
	Vault vault.Config
	// options for etcd and vault
	Options struct {
		Etcd  Etcd
		Vault Vault
	}
)

// NewStoreWatcher new a Watcher for backend storages
func NewStoreWatcher(c interface{}, confPath string, opts Options) (Watcher, error) {
	s := &storeWatcher{
		c:         c,
		confPath:  confPath,
		changeCh:  make(chan struct{}, DefaultChangeChan),
		hooks:     []Hook{},
		errHandle: defautlOnError,
	}

	var err error
	if s.etcd, err = etcd.NewClient(etcd.Config(opts.Etcd), s.onError); err != nil {
		return nil, err
	}

	if s.vault, err = vault.New(vault.Config(opts.Vault), s.changeCh, s.onError); err != nil {
		return nil, err
	}

	if err = s.loadConfig(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *storeWatcher) Config() interface{} {
	s.nCopyL.RLock()
	defer s.nCopyL.RUnlock()
	return s.nCopyed
}

func (s *storeWatcher) AddHook(hooks ...Hook) {
	s.hookL.Lock()
	s.hooks = append(s.hooks, hooks...)
	s.hookL.Unlock()
}

func (s *storeWatcher) onError(err error) {
	s.errHandle(err)
}

func (s *storeWatcher) OnError(h func(error)) {
	s.errHandle = h
}

func (s *storeWatcher) Watch() {
	go s.etcd.WatchKey(s.confPath, s.changeCh)
	s.procHooks()
}

// Change channel must be closed finally in case of panic
func (s *storeWatcher) Close() error {
	s.etcd.Close()
	s.vault.Close()
	close(s.changeCh)
	return nil
}

// loadConfig load configuration from etcd by given conf_path
func (s *storeWatcher) loadConfig() error {
	v, err := s.etcd.Key(s.confPath)
	if err != nil {
		return err
	}

	// now configuration only support json type
	if err = json.Unmarshal([]byte(v), s.c); err != nil {
		return err
	}

	// scan the struct and replace the key to value from vault
	if err = s.vault.Scan(s.c); err != nil {
		return err
	}

	s.oCopyed = s.nCopyed
	// s.c must be ptr type
	i := reflect.Indirect(reflect.ValueOf(s.c)).Interface()
	s.nCopyL.Lock()
	s.nCopyed = i
	s.nCopyL.Unlock()
	return nil
}

// procHooks reloads config and runs the hooks when the watched value has changed
func (s *storeWatcher) procHooks() {
	for range s.changeCh {
		if err := s.loadConfig(); err != nil {
			s.onError(err)
			continue
		}

		// hooks must be called one by one
		// bcs there may be dependencies
		for _, hook := range s.hooks {
			hook(s.oCopyed, s.Config())
		}
	}
}
