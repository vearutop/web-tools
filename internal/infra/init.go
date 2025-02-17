package infra

import (
	"context"

	"github.com/bool64/brick"
	"github.com/bool64/cache"
	"github.com/bool64/logz"
	"github.com/swaggest/rest/response/gzip"
	"github.com/vearutop/web-tools/internal/infra/schema"
	"github.com/vearutop/web-tools/internal/infra/service"
)

// NewServiceLocator creates application service locator.
func NewServiceLocator(cfg service.Config) (loc *service.Locator, err error) {
	l := &service.Locator{}

	defer func() {
		if err != nil && l != nil && l.LoggerProvider != nil {
			l.CtxdLogger().Error(context.Background(), err.Error())
		}
	}()

	l.BaseLocator, err = brick.NewBaseLocator(cfg.BaseConfig)
	if err != nil {
		return nil, err
	}

	schema.SetupOpenapiCollector(l.OpenAPI)

	l.HTTPServerMiddlewares = append(l.HTTPServerMiddlewares, gzip.Middleware)

	if err := l.TransferCache(context.Background()); err != nil {
		l.CtxdLogger().Warn(context.Background(), "failed to transfer cache", "error", err)
	}

	l.AccessLogs = cache.NewFailoverOf[*logz.Observer](func(cfg *cache.FailoverConfigOf[*logz.Observer]) {
		cfg.BackendConfig.CountSoftLimit = 1000
		cfg.BackendConfig.EvictionStrategy = cache.EvictLeastRecentlyUsed
		cfg.BackendConfig.TimeToLive = cache.UnlimitedTTL
	})

	return l, nil
}
