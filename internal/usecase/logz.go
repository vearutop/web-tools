package usecase

import (
	"context"
	"github.com/bool64/logz"
	"github.com/bool64/logz/logzpage"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
	"github.com/swaggest/usecase"
)

func Logz(deps interface {
	AccessLogzObserver(ctx context.Context, key string) (*logz.Observer, error)
}) usecase.Interactor {
	type req struct {
		request.EmbeddedSetter

		Key string `path:"key"`
	}
	u := usecase.NewInteractor(func(ctx context.Context, input req, output *response.EmbeddedSetter) error {
		o, err := deps.AccessLogzObserver(ctx, input.Key)
		if err != nil {
			return err
		}

		logzpage.Handler(o).ServeHTTP(output.ResponseWriter(), input.Request())

		return nil
	})

	return u
}
