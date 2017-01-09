package confl

import "context"

type Confl interface {
	LoadConfig(ctx context.Context, config interface{}) error
	WatchConfig(ctx context.Context, config interface{}, reload func() error) <-chan error
}
