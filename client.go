package confl

type Client interface {
	Key(key string) (string, error)
	WatchKey(key string, reloadCh chan<- struct{}, stopCh <-chan struct{}, errCh chan<- error)
	Close() error
}
