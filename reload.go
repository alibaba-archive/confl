package confl

type ReloadFunc func() error

func (f ReloadFunc) Reload() error {
	return f()
}

type Reloader interface {
	Reload() error
}
