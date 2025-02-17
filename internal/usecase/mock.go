package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/bool64/ctxd"
	"github.com/bool64/logz"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
	"github.com/swaggest/usecase"
)

func Mock(deps interface {
	AccessLogzObserver(ctx context.Context, key string) (*logz.Observer, error)
}) usecase.Interactor {
	type req struct {
		request.EmbeddedSetter
		LogsKey         string            `query:"logs_key" json:"-" description:"Access log key to collect request data, put something random here"`
		R301            string            `query:"r301" json:"r301,omitempty" description:"URL to redirect to with 302"`
		R302            string            `query:"r302"  json:"r302,omitempty" description:"URL to redirect to with 302"`
		Status          int               `query:"status" json:"status,omitempty" description:"Response status code, for non 301/302 cases, default 200"`
		RespBody        string            `query:"body" json:"respBody,omitempty" description:"Response body, text or base64(brotli) - see /compress to prepare body, default empty"`
		RespContentType string            `query:"ct" json:"ct,omitempty" description:"Response Content-Type, default text/plain"`
		RespHeaders     map[string]string `query:"headers" json:"respHeaders,omitempty" description:"Response headers, default empty"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input req, output *response.EmbeddedSetter) error {
		if input.LogsKey != "" {
			o, err := deps.AccessLogzObserver(ctx, input.LogsKey)
			if err != nil {
				return err
			}

			o.ObserveMessage("mock page served", ctxd.Tuples{
				"input", input,
				"header", input.Request().Header,
				"requestUri", input.Request().RequestURI,
			})
		}

		rw := output.ResponseWriter()

		if input.R301 != "" {
			http.Redirect(rw, input.Request(), input.R301, http.StatusMovedPermanently)
			return nil
		}

		if input.R302 != "" {
			http.Redirect(rw, input.Request(), input.R302, http.StatusFound)
			return nil
		}

		if input.RespContentType != "" {
			rw.Header().Set("Content-Type", input.RespContentType)
		} else {
			rw.Header().Set("Content-Type", "text/plain")
		}

		for k, v := range input.RespHeaders {
			rw.Header().Set(k, v)
		}

		if input.Status > 0 {
			rw.WriteHeader(input.Status)
		}

		if b64, err := base64.StdEncoding.DecodeString(input.RespBody); err == nil {
			b := bytes.NewReader(b64)

			r := brotli.NewReader(b)
			raw, err := io.ReadAll(r)
			if err != nil {
				return err
			}

			input.RespBody = string(raw)
		}

		_, _ = rw.Write([]byte(input.RespBody))

		return nil
	})

	u.SetTags("Mock")
	u.SetDescription("Mock serves pre-defined response based on query parameters.")

	return u
}
