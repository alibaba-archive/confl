package confl

import "context"

type Confl interface {
	LoadConfig(ctx context.Context, config interface{}, key string) error
	WatchConfig(ctx context.Context, config interface{}, key string, reload func() error) <-chan error
}
