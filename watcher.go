package confl

import (
	"errors"
	"sync"

	"github.com/teambition/confl/etcd"
	"github.com/teambition/confl/vault"
)

var (
	DefaultReloaderChan = 10
)

// Watcher manage the watch states
type Watcher struct {
	sync.RWMutex
	c              Configuration
	etcd, vault    Client
	reloadCh       chan struct{}
	reloaders      []Reloader
	errCh          chan error
	doneCh, stopCh chan struct{}
}

func NewWatcherFromEnv(c Configuration, doneCh, stopCh chan struct{}, errCh chan error) (*Watcher, error) {
	var err error
	if c.Path() == "" {
		return nil, errors.New("need config path")
	}

	w := &Watcher{
		c:         c,
		reloadCh:  make(chan struct{}, DefaultReloaderChan),
		reloaders: []Reloader{},
		errCh:     errCh,
		doneCh:    doneCh,
		stopCh:    stopCh,
	}

	if w.etcd, err = etcd.NewClientFromEnv(); err != nil {
		return nil, err
	}

	if w.vault, err = vault.NewClientFromEnv(); err != nil {
		return nil, err
	}

	if err = w.loadConfig(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) AddReloaders(rs ...Reloader) {
	w.Lock()
	w.reloaders = append(w.reloaders, rs...)
	w.Unlock()
}

func (w *Watcher) GoWatch() {
	go w.etcd.WatchKey(w.c.Path(), w.reloadCh, w.stopCh, w.errCh)
	go w.vault.WatchKey("", w.reloadCh, w.stopCh, w.errCh)
	w.runReloaders()
}

func (w *Watcher) Close() error {
	close(w.reloadCh)
	var err error
	if err = w.etcd.Close(); err != nil {
		return err
	}

	return w.vault.Close()
}

func (w *Watcher) loadConfig() error {
	v, err := w.etcd.Key(w.c.Path())
	if err != nil {
		return err
	}

	return w.c.Unmarshal([]byte(v))
}

// runReloaders run reloaders when the value changes
// which contained etcd and vault background storage
func (w *Watcher) runReloaders() {
	for range w.reloadCh {
		w.Lock()
		if err := w.loadConfig(); err != nil {
			w.errCh <- err
			continue
		}

		// reloaders have dependency order
		// need run reload one by one
		for _, r := range w.reloaders {
			err := r.Reload()
			if err != nil {
				w.errCh <- err
			}

		}
		w.Unlock()
	}
}
