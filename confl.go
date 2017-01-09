package confl

import "context"

type ReloadFunc func() error

func (f ReloadFunc) Reload() error {
	return f()
}

type Reloader interface {
	Reload() error
}

type Confl interface {
	LoadConfig(context.Context, interface{}) error
	WatchConfig(context.Context, interface{}, Reloader, chan<- error) error
}
