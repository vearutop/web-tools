package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/bool64/ctxd"
	"github.com/bool64/logz"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
	"github.com/swaggest/usecase"
)

type mockInput struct {
	request.EmbeddedSetter
	LogsKey         string            `query:"logs_key" json:"-" description:"Access log key to collect request data, put something random here"`
	R301            string            `query:"r301" json:"r301,omitempty" description:"URL to redirect to with 302"`
	R302            string            `query:"r302"  json:"r302,omitempty" description:"URL to redirect to with 302"`
	Status          int               `query:"status" json:"status,omitempty" description:"Response status code, for non 301/302 cases, default 200"`
	RespBody        string            `query:"body" json:"respBody,omitempty" description:"Response body, text or base64(brotli) - see /compress to prepare body, default empty"`
	RespContentType string            `query:"ct" json:"ct,omitempty" description:"Response Content-Type, default text/plain"`
	RespHeaders     map[string]string `query:"headers" json:"respHeaders,omitempty" description:"Response headers, default empty"`
}

func (input *mockInput) prepare() error {
	if input.RespContentType == "" {
		input.RespContentType = "text/plain"
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

	return nil
}

// Mock serves a request with a predefined response.
func Mock(deps interface {
	AccessLogzObserver(ctx context.Context, key string) (*logz.Observer, error)
	CtxdLogger() ctxd.Logger
},
) usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, input mockInput, output *response.EmbeddedSetter) error {
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

		if err := input.prepare(); err != nil {
			return err
		}

		rw.Header().Set("Content-Type", input.RespContentType)

		for k, v := range input.RespHeaders {
			rw.Header().Set(k, v)
		}

		if input.Status > 0 {
			rw.WriteHeader(input.Status)
		}

		if _, err := rw.Write([]byte(input.RespBody)); err != nil {
			deps.CtxdLogger().Error(ctx, "failed to write response", "error", err)
		}

		return nil
	})

	u.SetTags("Mock")
	u.SetDescription("Mock serves pre-defined response based on query parameters.")

	return u
}
