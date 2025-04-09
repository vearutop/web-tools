// Package nethttp manages application http interface.
package nethttp

import (
	"net/http"

	"github.com/bool64/brick"
	"github.com/swaggest/rest/nethttp"
	"github.com/vearutop/web-tools/internal/infra/service"
	"github.com/vearutop/web-tools/internal/usecase"
)

// NewRouter creates an instance of router filled with handlers and docs.
func NewRouter(deps *service.Locator) http.Handler {
	r := brick.NewBaseWebService(deps.BaseLocator)

	r.Post("/create-mock", usecase.CreateMock(deps))

	r.Handle("/og.html", nethttp.NewHandler(usecase.OG(deps)))

	r.Handle("/mock", nethttp.NewHandler(usecase.Mock(deps)))
	r.Get("/logs/{key}", usecase.Logz(deps))
	r.Post("/compress", usecase.Compress(deps))

	r.Get("/.well-known/assetlinks.json", usecase.AssetLinksJSON(deps))

	r.OnNotFound(usecase.ShortMock(deps))

	return r
}
