package usecase

import (
	"context"
	"github.com/bool64/logz"
	"net/http"

	"github.com/bool64/ctxd"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
	"github.com/swaggest/usecase"
	"golang.org/x/net/html"
)

type ogDeps interface {
	CtxdLogger() ctxd.Logger
}

func OG(deps interface {
	AccessLogzObserver(ctx context.Context, key string) (*logz.Observer, error)
}) usecase.Interactor {
	type req struct {
		request.EmbeddedSetter
		LogsKey   string `query:"logs_key" json:"-" description:"Access log key to collect request data"`
		TargetURL string `query:"target_url"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input req, output *response.EmbeddedSetter) error {
		if input.LogsKey != "" {
			o, err := deps.AccessLogzObserver(ctx, input.LogsKey)
			if err != nil {
				return err
			}

			o.ObserveMessage("og page requested", ctxd.Tuples{
				"input", input,
				"header", input.Request().Header,
				"requestUri", input.Request().RequestURI,
			})
		}

		rw := output.ResponseWriter()

		if input.TargetURL != "" {
			http.Redirect(rw, input.Request(), input.TargetURL, http.StatusMovedPermanently)
		}

		rw.Header().Set("Content-Type", "text/html")

		hd := ""
		for k, v := range input.Request().Header {
			hd += "<p>" + k + ": " + v[0] + "</p>"
		}

		_, _ = rw.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <title>OpenGraph Echo</title>

    <meta property="og:title" content="User-Agent: ` + html.EscapeString(input.Request().Header.Get("User-Agent")) + `"/>
    <meta property="og:description" content="URL: ` + html.EscapeString(input.Request().RequestURI) +
			`, IP: ` + input.Request().Header.Get("X-Forwarded-For") +
			`, Accept-Language: ` + input.Request().Header.Get("Accept-Language") +
			`"/>
    <meta property="og:type" content="website"/>
</head>

<p>User-Agent:  ` + html.EscapeString(input.Request().Header.Get("User-Agent")) + `</p>
<p>IP: ` + input.Request().Header.Get("X-Forwarded-For") + `</p>
<p>URL:  ` + html.EscapeString(input.Request().RequestURI) + `</p>
<p>Referer:  ` + html.EscapeString(input.Request().Header.Get("Referer")) + `</p>

Headers: ` + hd + `

</html>
`))

		return nil
	})

	u.SetTags("OpenGraph")

	return u
}
