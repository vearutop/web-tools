package usecase

import (
	"context"
	"net/http"

	"github.com/bool64/ctxd"
	"github.com/bool64/logz"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
	"github.com/swaggest/usecase"
	"golang.org/x/net/html"
)

type ogInput struct {
	request.EmbeddedSetter
	LogsKey   string `query:"logs_key" json:"-" description:"Access log key to collect request data"`
	TargetURL string `query:"target_url"`
}

// OG serves a page with OpenGraph tags.
func OG(deps interface {
	CtxdLogger() ctxd.Logger
	AccessLogzObserver(ctx context.Context, key string) (*logz.Observer, error)
},
) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input ogInput, output *response.EmbeddedSetter) error {
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

		if _, err := rw.Write(ogBody(input, hd)); err != nil {
			deps.CtxdLogger().Error(ctx, "failed to write og body", "error", err)
		}

		return nil
	})

	u.SetTags("OpenGraph")

	return u
}

func ogBody(input ogInput, hd string) []byte {
	return []byte(`
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
`)
}
