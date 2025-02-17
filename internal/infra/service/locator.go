package service

import (
	"context"
	"github.com/bool64/brick"
	"github.com/bool64/cache"
	"github.com/bool64/logz"
	"time"
)

// Locator defines application resources.
type Locator struct {
	*brick.BaseLocator

	AccessLogs *cache.FailoverOf[*logz.Observer]
}

func (l *Locator) AccessLogzObserver(ctx context.Context, key string) (*logz.Observer, error) {
	return l.AccessLogs.Get(ctx, []byte(key), func(ctx context.Context) (*logz.Observer, error) {
		o := logz.Observer{}

		o.SamplingInterval = time.Nanosecond
		o.MaxSamples = 500

		return &o, nil
	})
}
